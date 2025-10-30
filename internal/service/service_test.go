package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/repository"
)

func TestNewService(t *testing.T) {
	repo := repository.NewRepository()
	service := NewService(repo)

	if service == nil {
		t.Fatal("NewService returned nil")
	}

	if service.repo != repo {
		t.Error("Service should use the provided repository")
	}
}
func TestStartDataGeneration(t *testing.T) {
	repo := repository.NewRepository()
	service := NewService(repo)

	// Запускаем генерацию и ждем ее завершения СИНХРОННО
	interval := 1 * time.Millisecond
	service.StartDataGeneration(interval) // Запускаем в той же горутине

	// Теперь безопасно проверяем результаты
	notesCount := repo.GetNotesCount()
	if notesCount != 5 {
		t.Errorf("Expected 5 notes, got %d", notesCount)
	}

	// Проверяем содержимое заметок
	notes := repo.GetAllNotes()
	for i, note := range notes {
		expectedTitle := fmt.Sprintf("Тестовая заметка %d", i+1)
		expectedContent := fmt.Sprintf("Это содержимое тестовой заметки номер %d", i+1)

		if note.GetTitle() != expectedTitle {
			t.Errorf("Note %d: expected title %q, got %q", i+1, expectedTitle, note.GetTitle())
		}

		if note.GetContent() != expectedContent {
			t.Errorf("Note %d: expected content %q, got %q", i+1, expectedContent, note.GetContent())
		}

		// Проверяем что заметка имеет ID
		if note.GetID() == "" {
			t.Errorf("Note %d: ID should not be empty", i+1)
		}

		// Проверяем временные метки
		if note.GetCreatedAt().IsZero() {
			t.Errorf("Note %d: CreatedAt should be set", i+1)
		}
		if note.GetUpdatedAt().IsZero() {
			t.Errorf("Note %d: UpdatedAt should be set", i+1)
		}
	}
}

func TestStartDataGenerationStopsAfterFiveNotes(t *testing.T) {
	repo := repository.NewRepository()
	service := NewService(repo)

	// Запускаем генерацию синхронно
	interval := 1 * time.Millisecond
	startTime := time.Now()
	service.StartDataGeneration(interval)

	// Проверяем что выполнение заняло разумное время
	executionTime := time.Since(startTime)
	if executionTime > time.Second {
		t.Errorf("Data generation should complete quickly, took %v", executionTime)
	}

	// Проверяем что создалось ровно 5 заметок
	notesCount := repo.GetNotesCount()
	if notesCount != 5 {
		t.Errorf("Expected exactly 5 notes, got %d", notesCount)
	}
}

func TestServiceIsolation(t *testing.T) {
	// Тестируем что разные сервисы работают независимо
	repo1 := repository.NewRepository()
	repo2 := repository.NewRepository()

	service1 := NewService(repo1)
	service2 := NewService(repo2)

	service1.StartDataGeneration(1 * time.Millisecond)
	service2.StartDataGeneration(1 * time.Millisecond)

	// Оба репозитория должны иметь по 5 заметок
	if repo1.GetNotesCount() != 5 {
		t.Errorf("Repo1 should have 5 notes, got %d", repo1.GetNotesCount())
	}
	if repo2.GetNotesCount() != 5 {
		t.Errorf("Repo2 should have 5 notes, got %d", repo2.GetNotesCount())
	}

	// Заметки в разных репозиториях должны быть независимы
	notes1 := repo1.GetAllNotes()
	notes2 := repo2.GetAllNotes()

	for i := 0; i < 5; i++ {
		if notes1[i].GetID() == notes2[i].GetID() {
			t.Errorf("Notes in different repositories should have different IDs")
		}
	}
}

func TestServiceWithNilRepository(t *testing.T) {
	// Тестируем что сервис не паникует при работе с nil репозиторием
	service := NewService(nil)

	// Запускаем синхронно - должно завершиться сразу
	service.StartDataGeneration(1 * time.Millisecond)

	// Если не было паники - тест пройден
}

func TestNoteCounterIncrementsCorrectly(t *testing.T) {
	repo := repository.NewRepository()
	service := NewService(repo)

	// Запускаем генерацию синхронно
	service.StartDataGeneration(1 * time.Millisecond)

	// Проверяем что заметки имеют правильную нумерацию
	notes := repo.GetAllNotes()

	for i, note := range notes {
		expectedNumber := i + 1
		expectedTitle := fmt.Sprintf("Тестовая заметка %d", expectedNumber)
		expectedContent := fmt.Sprintf("Это содержимое тестовой заметки номер %d", expectedNumber)

		if note.GetTitle() != expectedTitle {
			t.Errorf("Note %d has wrong title: %q", expectedNumber, note.GetTitle())
		}
		if note.GetContent() != expectedContent {
			t.Errorf("Note %d has wrong content: %q", expectedNumber, note.GetContent())
		}
	}
}
