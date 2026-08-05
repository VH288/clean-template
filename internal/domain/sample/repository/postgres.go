package repository

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

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, sample *entity.Sample) error {
	if sample.ID == "" {
		sample.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	sample.CreatedAt = now
	sample.UpdatedAt = now

	const q = `
		INSERT INTO samples (id, name, description, status, created_at, updated_at)
		VALUES (:id, :name, :description, :status, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, q, sample)
	if err != nil {
		return fmt.Errorf("insert sample: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	var sample entity.Sample
	const q = `
		SELECT id, name, description, status, created_at, updated_at
		FROM samples WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &sample, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get sample: %w", err)
	}
	return &sample, nil
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]entity.Sample, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM samples`); err != nil {
		return nil, 0, fmt.Errorf("count samples: %w", err)
	}

	var items []entity.Sample
	const q = `
		SELECT id, name, description, status, created_at, updated_at
		FROM samples
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	if err := r.db.SelectContext(ctx, &items, q, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list samples: %w", err)
	}
	if items == nil {
		items = []entity.Sample{}
	}
	return items, total, nil
}

func (r *PostgresRepository) Update(ctx context.Context, sample *entity.Sample) error {
	sample.UpdatedAt = time.Now().UTC()
	const q = `
		UPDATE samples
		SET name = :name, description = :description, status = :status, updated_at = :updated_at
		WHERE id = :id
	`
	res, err := r.db.NamedExecContext(ctx, q, sample)
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
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM samples WHERE id = $1`, id)
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
	return nil
}
