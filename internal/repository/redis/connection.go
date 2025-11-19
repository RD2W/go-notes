package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisClient обертка для клиента Redis
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient создает новый экземпляр Redis клиента
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	return &RedisClient{
		client: rdb,
	}, nil
}

// GetClient возвращает экземпляр Redis клиента
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Close закрывает соединение с Redis
func (r *RedisClient) Close() error {
	return r.client.Close()
}
