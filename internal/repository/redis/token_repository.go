package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rd2w/go-notes/internal/database"
	"github.com/rd2w/go-notes/internal/domain/repository"
	"github.com/redis/go-redis/v9"
)

// RedisTokenRepository реализация хранилища токенов в Redis
type RedisTokenRepository struct {
	client *redis.Client
}

// Ensure RedisTokenRepository implements TokenRepository interface
var _ repository.TokenRepository = (*RedisTokenRepository)(nil)

// NewRedisTokenRepository создает новое хранилище токенов в Redis с указанным клиентом
func NewRedisTokenRepository(redisClient *database.RedisClient) (*RedisTokenRepository, error) {
	return &RedisTokenRepository{
		client: redisClient.GetClient(),
	}, nil
}

// AddToBlacklist добавляет токен в черный список
func (r *RedisTokenRepository) AddToBlacklist(tokenID string, expiresAt time.Time) error {
	ctx := context.Background()

	// Вычисляем время до истечения в секундах
	ttl := time.Until(expiresAt)

	// Добавляем токен в черный список с TTL
	err := r.client.SetEx(ctx, "blacklist:"+tokenID, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("ошибка добавления токена в черный список: %w", err)
	}

	return nil
}

// IsBlacklisted проверяет, находится ли токен в черном списке
func (r *RedisTokenRepository) IsBlacklisted(tokenID string) (bool, error) {
	ctx := context.Background()

	// Проверяем наличие токена в черном списке
	val, err := r.client.Get(ctx, "blacklist:"+tokenID).Result()
	if errors.Is(err, redis.Nil) {
		// Ключ не существует, токен не в черном списке
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("ошибка проверки токена в черном списке: %w", err)
	}

	// Если значение существует, токен в черном списке
	return val == "1", nil
}
