package sample

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"clean-template/internal/domain/sample/entity"
	apperrors "clean-template/internal/pkg/errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresModel struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toModel(s *entity.Sample) postgresModel {
	return postgresModel{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Status:      s.Status,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func fromModel(m postgresModel) *entity.Sample {
	return &entity.Sample{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, sample *entity.Sample) error {
	return r.CreateWithEvent(ctx, sample, "", nil)
}

func (r *PostgresRepository) CreateWithEvent(ctx context.Context, sample *entity.Sample, eventType string, eventPayload []byte) error {
	if sample.ID == "" {
		sample.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	sample.CreatedAt = now
	sample.UpdatedAt = now

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const q = `
		INSERT INTO samples (id, name, description, status, created_at, updated_at)
		VALUES (:id, :name, :description, :status, :created_at, :updated_at)
	`
	if _, err := tx.NamedExecContext(ctx, q, toModel(sample)); err != nil {
		return fmt.Errorf("insert sample: %w", err)
	}

	if eventType != "" {
		if err := insertOutboxEvent(ctx, tx, sample.ID, eventType, eventPayload); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	var model postgresModel
	const q = `
		SELECT id, name, description, status, created_at, updated_at
		FROM samples WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &model, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get sample: %w", err)
	}
	return fromModel(model), nil
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]entity.Sample, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM samples`); err != nil {
		return nil, 0, fmt.Errorf("count samples: %w", err)
	}

	var models []postgresModel
	const q = `
		SELECT id, name, description, status, created_at, updated_at
		FROM samples
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	if err := r.db.SelectContext(ctx, &models, q, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list samples: %w", err)
	}

	items := make([]entity.Sample, 0, len(models))
	for _, m := range models {
		items = append(items, *fromModel(m))
	}
	return items, total, nil
}

func (r *PostgresRepository) Update(ctx context.Context, sample *entity.Sample) error {
	return r.UpdateWithEvent(ctx, sample, "", nil)
}

func (r *PostgresRepository) UpdateWithEvent(ctx context.Context, sample *entity.Sample, eventType string, eventPayload []byte) error {
	sample.UpdatedAt = time.Now().UTC()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const q = `
		UPDATE samples
		SET name = :name, description = :description, status = :status, updated_at = :updated_at
		WHERE id = :id
	`
	res, err := tx.NamedExecContext(ctx, q, toModel(sample))
	if err != nil {
		return fmt.Errorf("update sample: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return apperrors.ErrNotFound
	}

	if eventType != "" {
		if err := insertOutboxEvent(ctx, tx, sample.ID, eventType, eventPayload); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	return r.DeleteWithEvent(ctx, id, "", nil)
}

func (r *PostgresRepository) DeleteWithEvent(ctx context.Context, id, eventType string, eventPayload []byte) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `DELETE FROM samples WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete sample: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return apperrors.ErrNotFound
	}

	if eventType != "" {
		if err := insertOutboxEvent(ctx, tx, id, eventType, eventPayload); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func insertOutboxEvent(ctx context.Context, tx *sqlx.Tx, aggregateID, eventType string, payload []byte) error {
	const q = `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
	`
	_, err := tx.ExecContext(ctx, q, uuid.NewString(), "sample", aggregateID, eventType, payload)
	if err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}
