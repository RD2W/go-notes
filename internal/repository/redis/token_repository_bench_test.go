package redis

import (
	"context"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/database"
)

func getTestRedisClient() *database.RedisClient {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../../config/config_dev.toml")
	if err != nil {
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем клиент подключения к Redis
	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		panic(err)
	}

	return redisClient
}

// BenchmarkRedisTokenRepositoryAddToBlacklist benchmarks the AddToBlacklist method
func BenchmarkRedisTokenRepositoryAddToBlacklist(b *testing.B) {
	redisClient := getTestRedisClient()
	defer func() {
		if err := redisClient.Close(); err != nil {
			b.Logf("Ошибка закрытия Redis клиента: %v", err)
		}
	}()

	repo, err := NewRedisTokenRepository(redisClient)
	if err != nil {
		b.Fatalf("Ошибка создания репозитория токенов: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenID := "test_token_id_" + string(rune(i))
		expiresAt := time.Now().Add(1 * time.Hour)

		err := repo.AddToBlacklist(tokenID, expiresAt)
		if err != nil {
			// Не прерываем benchmark при ошибках, а просто логируем
			b.Logf("Ошибка добавления токена в черный список: %v", err)
		}
	}
}

// BenchmarkRedisTokenRepositoryIsBlacklisted benchmarks the IsBlacklisted method
func BenchmarkRedisTokenRepositoryIsBlacklisted(b *testing.B) {
	redisClient := getTestRedisClient()
	defer func() {
		if err := redisClient.Close(); err != nil {
			b.Logf("Ошибка закрытия Redis клиента: %v", err)
		}
	}()

	repo, err := NewRedisTokenRepository(redisClient)
	if err != nil {
		b.Fatalf("Ошибка создания репозитория токенов: %v", err)
	}

	// Подготовим тестовый токен в Redis
	testTokenID := "benchmark_test_token"
	testExpiresAt := time.Now().Add(1 * time.Hour)

	// Добавляем тестовый токен в черный список
	err = repo.AddToBlacklist(testTokenID, testExpiresAt)
	if err != nil {
		b.Fatalf("Ошибка подготовки тестового токена: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.IsBlacklisted(testTokenID)
		if err != nil {
			b.Logf("Ошибка проверки токена в черном списке: %v", err)
		}
	}

	// Удаляем тестовый токен из Redis после завершения
	redisClient.GetClient().Del(context.Background(), "blacklist:"+testTokenID)
}
