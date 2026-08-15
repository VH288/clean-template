package sample

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/pkg/constant"
	apperrors "clean-template/internal/pkg/errors"

	goredis "github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *goredis.Client
	ttl    time.Duration
}

func NewRedisRepository(client *goredis.Client) *RedisRepository {
	return &RedisRepository{
		client: client,
		ttl:    time.Duration(constant.SampleCacheTTLSec) * time.Second,
	}
}

func (r *RedisRepository) key(id string) string {
	return constant.SampleCacheKeyPrefix + id
}

func (r *RedisRepository) Set(ctx context.Context, sample *entity.Sample) error {
	payload, err := json.Marshal(sample)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(sample.ID), payload, r.ttl).Err()
}

func (r *RedisRepository) Get(ctx context.Context, id string) (*entity.Sample, error) {
	val, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err != nil {
		if err == goredis.Nil {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("redis get sample: %w", err)
	}
	var sample entity.Sample
	if err := json.Unmarshal(val, &sample); err != nil {
		return nil, err
	}
	return &sample, nil
}

func (r *RedisRepository) Delete(ctx context.Context, id string) error {
	return r.client.Del(ctx, r.key(id)).Err()
}
