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

func getTestPostgresPool() *database.PostgresClient {
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

	return postgresClient
}

// BenchmarkPostgresNoteRepositoryCreate benchmarks the Create method
func BenchmarkPostgresNoteRepositoryCreate(b *testing.B) {
	postgresClient := getTestPostgresPool()
	defer postgresClient.Close()

	repo, err := NewPostgresNoteRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория заметок: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		note := model.NewNote(fmt.Sprintf("Test Title %d", i), fmt.Sprintf("Test Content %d", i), "user123")
		note.SetID(fmt.Sprintf("note%d", i))
		note.SetCreatedAt(time.Now())
		note.SetUpdatedAt(time.Now())

		err := repo.Create(note)
		if err != nil {
			// Не прерываем benchmark при ошибках, а просто логируем
			b.Logf("Ошибка создания заметки: %v", err)
		}
	}
}

// BenchmarkPostgresNoteRepositoryGetByID benchmarks the GetByID method
func BenchmarkPostgresNoteRepositoryGetByID(b *testing.B) {
	postgresClient := getTestPostgresPool()
	defer postgresClient.Close()

	repo, err := NewPostgresNoteRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория заметок: %v", err)
	}

	// Подготовим тестовую заметку в базе данных
	testNote := model.NewNote("Test Title", "Test Content", "user123")
	testNote.SetID("benchmark_test_note")
	testNote.SetCreatedAt(time.Now())
	testNote.SetUpdatedAt(time.Now())

	// Создаем тестовую заметку
	err = repo.Create(testNote)
	if err != nil {
		b.Fatalf("Ошибка подготовки тестовой заметки: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByID("benchmark_test_note")
		if err != nil {
			b.Logf("Ошибка получения заметки: %v", err)
		}
	}

	// Удаляем тестовую заметку после завершения
	err = repo.DeleteByID("benchmark_test_note")
	if err != nil {
		b.Logf("Ошибка удаления тестовой заметки: %v", err)
	}
}

// BenchmarkPostgresNoteRepositoryUpdate benchmarks the Update method
func BenchmarkPostgresNoteRepositoryUpdate(b *testing.B) {
	postgresClient := getTestPostgresPool()
	defer postgresClient.Close()

	repo, err := NewPostgresNoteRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория заметок: %v", err)
	}

	// Подготовим тестовую заметку в базе данных
	testNote := model.NewNote("Original Title", "Original Content", "user123")
	testNote.SetID("benchmark_update_test")
	testNote.SetCreatedAt(time.Now())
	testNote.SetUpdatedAt(time.Now())

	// Создаем тестовую заметку
	err = repo.Create(testNote)
	if err != nil {
		b.Fatalf("Ошибка подготовки тестовой заметки: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		updatedNote := model.NewNote(fmt.Sprintf("Updated Title %d", i), fmt.Sprintf("Updated Content %d", i), "user123")
		updatedNote.SetID("benchmark_update_test")
		updatedNote.SetCreatedAt(testNote.GetCreatedAt()) // Сохраняем оригинальное время создания
		updatedNote.SetUpdatedAt(time.Now())

		err := repo.Update(updatedNote)
		if err != nil {
			b.Logf("Ошибка обновления заметки: %v", err)
		}
	}

	// Удаляем тестовую заметку после завершения
	err = repo.DeleteByID("benchmark_update_test")
	if err != nil {
		b.Logf("Ошибка удаления тестовой заметки: %v", err)
	}
}

// BenchmarkPostgresNoteRepositoryGetAllNotes benchmarks the GetAllNotes method
func BenchmarkPostgresNoteRepositoryGetAllNotes(b *testing.B) {
	postgresClient := getTestPostgresPool()
	defer postgresClient.Close()

	repo, err := NewPostgresNoteRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория заметок: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetAllNotes()
		if err != nil {
			b.Logf("Ошибка получения всех заметок: %v", err)
		}
	}
}

// BenchmarkPostgresNoteRepositoryGetAllNotesByUserID benchmarks the GetAllNotesByUserID method
func BenchmarkPostgresNoteRepositoryGetAllNotesByUserID(b *testing.B) {
	postgresClient := getTestPostgresPool()
	defer postgresClient.Close()

	repo, err := NewPostgresNoteRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория заметок: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetAllNotesByUserID("user123")
		if err != nil {
			b.Logf("Ошибка получения заметок пользователя: %v", err)
		}
	}
}

// BenchmarkPostgresNoteRepositoryGetListByUserID benchmarks the GetListByUserID method
func BenchmarkPostgresNoteRepositoryGetListByUserID(b *testing.B) {
	postgresClient := getTestPostgresPool()
	defer postgresClient.Close()

	repo, err := NewPostgresNoteRepository(postgresClient.GetPool())
	if err != nil {
		b.Fatalf("Ошибка создания репозитория заметок: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetListByUserID("user123", 10, 0)
		if err != nil {
			b.Logf("Ошибка получения списка заметок пользователя: %v", err)
		}
	}
}
