package postgres

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/database"
	"github.com/rd2w/go-notes/internal/domain/model"
)

// BenchmarkPostgresUserRepositoryCreate benchmarks the Create method
func BenchmarkPostgresUserRepositoryCreate(b *testing.B) {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем клиент подключения к PostgreSQL
	postgresClient, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer postgresClient.Close()

	repo, err := NewPostgresUserRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория пользователей: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user, err := model.NewUser(fmt.Sprintf("testuser%d", i), fmt.Sprintf("test%d@example.com", i), "password123")
		if err != nil {
			b.Logf("Ошибка создания пользователя: %v", err)
			continue
		}
		user.SetCreatedAt(time.Now())
		user.SetUpdatedAt(time.Now())

		err = repo.Create(user)
		if err != nil {
			// Не прерываем benchmark при ошибках, а просто логируем
			b.Logf("Ошибка создания пользователя: %v", err)
		}
	}
}

// BenchmarkPostgresUserRepositoryGetByID benchmarks the GetByID method
func BenchmarkPostgresUserRepositoryGetByID(b *testing.B) {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем клиент подключения к PostgreSQL
	postgresClient, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer postgresClient.Close()

	repo, err := NewPostgresUserRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория пользователей: %v", err)
	}

	// Подготовим тестового пользователя в базе данных
	user, err := model.NewUser("benchmark_test_user", "benchmark_test@example.com", "password123")
	if err != nil {
		b.Fatalf("Ошибка создания тестового пользователя: %v", err)
	}
	user.SetID("benchmark_test_user_id")
	user.SetCreatedAt(time.Now())
	user.SetUpdatedAt(time.Now())

	// Создаем тестового пользователя
	err = repo.Create(user)
	if err != nil {
		b.Fatalf("Ошибка подготовки тестового пользователя: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByID("benchmark_test_user_id")
		if err != nil {
			b.Logf("Ошибка получения пользователя: %v", err)
		}
	}

	// Удаляем тестового пользователя после завершения
	err = repo.DeleteByID("benchmark_test_user_id")
	if err != nil {
		b.Logf("Ошибка удаления тестового пользователя: %v", err)
	}
}

// BenchmarkPostgresUserRepositoryUpdate benchmarks the Update method
func BenchmarkPostgresUserRepositoryUpdate(b *testing.B) {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем клиент подключения к PostgreSQL
	postgresClient, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer postgresClient.Close()

	repo, err := NewPostgresUserRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория пользователей: %v", err)
	}

	// Подготовим тестового пользователя в базе данных
	user, err := model.NewUser("Original User", "original@example.com", "password123")
	if err != nil {
		b.Fatalf("Ошибка создания тестового пользователя: %v", err)
	}
	user.SetID("benchmark_update_user_id")
	user.SetCreatedAt(time.Now())
	user.SetUpdatedAt(time.Now())

	// Создаем тестового пользователя
	err = repo.Create(user)
	if err != nil {
		b.Fatalf("Ошибка подготовки тестового пользователя: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		updatedUser, err := model.NewUser(fmt.Sprintf("Updated User %d", i), fmt.Sprintf("updated%d@example.com", i), "password123")
		if err != nil {
			b.Logf("Ошибка создания обновленного пользователя: %v", err)
			continue
		}
		updatedUser.SetID("benchmark_update_user_id")
		updatedUser.SetCreatedAt(user.GetCreatedAt()) // Сохраняем оригинальное время создания
		updatedUser.SetUpdatedAt(time.Now())

		err = repo.Update(updatedUser)
		if err != nil {
			b.Logf("Ошибка обновления пользователя: %v", err)
		}
	}

	// Удаляем тестового пользователя после завершения
	err = repo.DeleteByID("benchmark_update_user_id")
	if err != nil {
		b.Logf("Ошибка удаления тестового пользователя: %v", err)
	}
}

// BenchmarkPostgresUserRepositoryGetAllUsers benchmarks the GetAllUsers method
func BenchmarkPostgresUserRepositoryGetAllUsers(b *testing.B) {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем клиент подключения к PostgreSQL
	postgresClient, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer postgresClient.Close()

	repo, err := NewPostgresUserRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория пользователей: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetAllUsers()
		if err != nil {
			b.Logf("Ошибка получения всех пользователей: %v", err)
		}
	}
}

// BenchmarkPostgresUserRepositoryGetUserByEmail benchmarks the GetUserByEmail method
func BenchmarkPostgresUserRepositoryGetUserByEmail(b *testing.B) {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем клиент подключения к PostgreSQL
	postgresClient, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer postgresClient.Close()

	repo, err := NewPostgresUserRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория пользователей: %v", err)
	}

	// Подготовим тестового пользователя в базе данных
	user, err := model.NewUser("Email Test User", "email_test@example.com", "password123")
	if err != nil {
		b.Fatalf("Ошибка создания тестового пользователя: %v", err)
	}
	user.SetID("benchmark_email_test_user_id")
	user.SetCreatedAt(time.Now())
	user.SetUpdatedAt(time.Now())

	// Создаем тестового пользователя
	err = repo.Create(user)
	if err != nil {
		b.Fatalf("Ошибка подготовки тестового пользователя: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetUserByEmail("email_test@example.com")
		if err != nil {
			b.Logf("Ошибка получения пользователя по email: %v", err)
		}
	}

	// Удаляем тестового пользователя после завершения
	err = repo.DeleteByID("benchmark_email_test_user_id")
	if err != nil {
		b.Logf("Ошибка удаления тестового пользователя: %v", err)
	}
}

// BenchmarkPostgresUserRepositoryGetUserByUsername benchmarks the GetUserByUsername method
func BenchmarkPostgresUserRepositoryGetUserByUsername(b *testing.B) {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../../config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию: %v", err)
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем клиент подключения к PostgreSQL
	postgresClient, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer postgresClient.Close()

	repo, err := NewPostgresUserRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория пользователей: %v", err)
	}

	// Подготовим тестового пользователя в базе данных
	user, err := model.NewUser("username_test", "Username Test User", "password123")
	if err != nil {
		b.Fatalf("Ошибка создания тестового пользователя: %v", err)
	}
	user.SetID("benchmark_username_test_user_id")
	user.SetCreatedAt(time.Now())
	user.SetUpdatedAt(time.Now())

	// Создаем тестового пользователя
	err = repo.Create(user)
	if err != nil {
		b.Fatalf("Ошибка подготовки тестового пользователя: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetUserByUsername("username_test")
		if err != nil {
			b.Logf("Ошибка получения пользователя по username: %v", err)
		}
	}

	// Удаляем тестового пользователя после завершения
	err = repo.DeleteByID("benchmark_username_test_user_id")
	if err != nil {
		b.Logf("Ошибка удаления тестового пользователя: %v", err)
	}
}
