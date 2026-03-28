package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RedisClient struct {
	pool *pgxpool.Pool
}

func NewRedisClient(pool *pgxpool.Pool) *RedisClient {
	return &RedisClient{pool: pool}
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	query := `SELECT data FROM redis_cache WHERE key = $1`
	row := c.pool.QueryRow(ctx, query, key)

	var data string
	err := row.Scan(&data)
	if err != nil {
		return "", fmt.Errorf("key not found: %w", err)
	}
	return data, nil
}

func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	query := `
		INSERT INTO redis_cache (key, data, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE SET data = $2, expires_at = $3
	`

	expiresAt := time.Now().Add(expiration)
	_, err = c.pool.Exec(ctx, query, key, string(data), expiresAt)
	return err
}

func (c *RedisClient) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM redis_cache WHERE key = $1`
	_, err := c.pool.Exec(ctx, query, key)
	return err
}

func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM redis_cache WHERE key = $1 AND (expires_at IS NULL OR expires_at > NOW()))`
	row := c.pool.QueryRow(ctx, query, key)

	var exists bool
	err := row.Scan(&exists)
	return exists, err
}