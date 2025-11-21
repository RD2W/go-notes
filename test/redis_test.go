package main_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/database"
	"github.com/rd2w/go-notes/internal/repository/redis"
)

func TestRedisIntegration(t *testing.T) {
	fmt.Println("Тестирование подключения к Redis и работы с данными...")

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем клиент подключения к Redis
	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	defer redisClient.Close()

	// Создаем репозитории
	tokenRepo, err := redis.NewRedisTokenRepository(redisClient)
	if err != nil {
		t.Fatalf("Ошибка создания репозитория токенов: %v", err)
	}

	// Генерируем уникальный ID токена для теста
	tokenID := "test_token_" + fmt.Sprint(time.Now().Unix())

	// Тестируем добавление токена в черный список
	expirationTime := time.Now().Add(10 * time.Minute)
	err = tokenRepo.AddToBlacklist(tokenID, expirationTime)
	if err != nil {
		t.Fatalf("Ошибка добавления токена в черный список: %v", err)
	}
	fmt.Printf("✓ Токен %s добавлен в черный список\n", tokenID)

	// Проверяем, что токен находится в черном списке
	isBlacklisted, err := tokenRepo.IsBlacklisted(tokenID)
	if err != nil {
		t.Fatalf("Ошибка проверки токена в черном списке: %v", err)
	}
	if isBlacklisted {
		fmt.Printf("✓ Токен %s найден в черном списке\n", tokenID)
	} else {
		t.Errorf("✗ Токен %s не найден в черном списке\n", tokenID)
	}

	// Проверяем, что несуществующий токен не находится в черном списке
	nonExistentTokenID := "non_existent_token_" + fmt.Sprint(time.Now().Unix())
	isBlacklisted, err = tokenRepo.IsBlacklisted(nonExistentTokenID)
	if err != nil {
		t.Fatalf("Ошибка проверки несуществующего токена в черном списке: %v", err)
	}
	if !isBlacklisted {
		fmt.Printf("✓ Несуществующий токен %s не найден в черном списке\n", nonExistentTokenID)
	} else {
		t.Errorf("✗ Несуществующий токен %s найден в черном списке\n", nonExistentTokenID)
	}

	// Тестируем TTL токена в черном списке - устанавливаем короткое время жизни для теста
	shortExpirationTime := time.Now().Add(1 * time.Second)
	shortTokenID := "short_lived_token_" + fmt.Sprint(time.Now().Unix())
	err = tokenRepo.AddToBlacklist(shortTokenID, shortExpirationTime)
	if err != nil {
		t.Fatalf("Ошибка добавления токена с коротким временем жизни в черный список: %v", err)
	}
	fmt.Printf("✓ Токен %s добавлен в черный список с коротким временем жизни\n", shortTokenID)

	// Проверяем, что токен с коротким временем жизни все еще в черном списке
	isBlacklisted, err = tokenRepo.IsBlacklisted(shortTokenID)
	if err != nil {
		t.Fatalf("Ошибка проверки токена с коротким временем жизни в черном списке: %v", err)
	}
	if isBlacklisted {
		fmt.Printf("✓ Токен %s все еще в черном списке\n", shortTokenID)
	} else {
		t.Errorf("✗ Токен %s не найден в черном списке\n", shortTokenID)
	}

	// Ждем, пока токен с коротким временем жизни не исчезнет из черного списка
	fmt.Println("Ожидание истечения срока действия токена...")
	time.Sleep(2 * time.Second)

	// Проверяем, что токен с коротким временем жизни больше не в черном списке
	isBlacklisted, err = tokenRepo.IsBlacklisted(shortTokenID)
	if err != nil {
		t.Fatalf("Ошибка проверки истекшего токена в черном списке: %v", err)
	}
	if !isBlacklisted {
		fmt.Printf("✓ Токен %s больше не находится в черном списке (TTL истек)\n", shortTokenID)
	} else {
		t.Errorf("✗ Токен %s все еще находится в черном списке (TTL не истек)\n", shortTokenID)
	}

	fmt.Println("\n✓ Все тесты Redis пройдены успешно!")
}
