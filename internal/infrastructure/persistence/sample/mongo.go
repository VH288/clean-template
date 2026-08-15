package sample

import (
	"context"
	"errors"
	"fmt"
	"time"

	"clean-template/internal/domain/sample/entity"
	apperrors "clean-template/internal/pkg/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	collection *mongo.Collection
}

type mongoSampleDoc struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Status      string    `bson:"status"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{collection: db.Collection("samples")}
}

func (r *MongoRepository) Upsert(ctx context.Context, sample *entity.Sample) error {
	doc := mongoSampleDoc{
		ID:          sample.ID,
		Name:        sample.Name,
		Description: sample.Description,
		Status:      sample.Status,
		CreatedAt:   sample.CreatedAt,
		UpdatedAt:   sample.UpdatedAt,
	}
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": sample.ID},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("mongo upsert sample: %w", err)
	}
	return nil
}

func (r *MongoRepository) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	var doc mongoSampleDoc
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("mongo get sample: %w", err)
	}
	return &entity.Sample{
		ID:          doc.ID,
		Name:        doc.Name,
		Description: doc.Description,
		Status:      doc.Status,
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
	}, nil
}

func (r *MongoRepository) Delete(ctx context.Context, id string) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("mongo delete sample: %w", err)
	}
	if res.DeletedCount == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
