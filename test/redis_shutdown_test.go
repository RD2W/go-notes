package main_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/database"
)

func TestRedisShutdown(t *testing.T) {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем Redis клиент напрямую для тестирования
	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("Ошибка создания Redis клиента: %v", err)
	}

	// Проверяем, что клиент работает
	client := redisClient.GetClient()
	ctx := context.Background()
	err = client.Set(ctx, "test_key", "test_value", 0).Err()
	if err != nil {
		t.Fatalf("Ошибка при записи в Redis: %v", err)
	}

	// Закрываем соединение
	err = redisClient.Close()
	if err != nil {
		t.Fatalf("Ошибка при закрытии Redis соединения: %v", err)
	}

	// Ждем немного, чтобы соединение точно закрылось
	time.Sleep(10 * time.Millisecond)

	// Проверяем, что после закрытия клиент больше не работает
	err = client.Set(ctx, "test_after_close", "should_fail", 0).Err()
	// Ожидаем ошибку, так как соединение закрыто
	if err == nil {
		t.Error("Ожидается ошибка при использовании клиента после закрытия, но ошибки не произошло")
	}
}

func TestTokenManagerClose(t *testing.T) {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем Redis клиент
	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("Ошибка создания Redis клиента: %v", err)
	}

	// Создаем TokenManager
	tokenManager := auth.NewTokenManager(cfg, redisClient)

	// Проверяем, что TokenManager создался без ошибок
	if tokenManager == nil {
		t.Fatal("TokenManager не должен быть nil")
	}

	// Закрываем Redis клиент
	err = redisClient.Close()
	if err != nil {
		t.Fatalf("Ошибка при закрытии Redis клиента: %v", err)
	}

	// Успешное выполнение без паники означает, что закрытие прошло корректно
	t.Log("Redis клиент успешно закрыт")
}
