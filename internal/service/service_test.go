package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/repository/storage/fs"
	"github.com/rd2w/go-notes/internal/repository/storage/ram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestService_StopWithContext тестирует остановку генерации через контекст
func TestService_StopWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := ram.NewRamRepository()
	service := NewService(repo, ctx, 10*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание хотя бы одной заметки
	time.Sleep(15 * time.Millisecond)

	// Останавливаем сервис
	cancel()

	// Ждем завершения
	time.Sleep(20 * time.Millisecond)

	// Проверяем, что сервис остановился и создал заметки до остановки
	notes := repo.GetAllNotes()
	assert.Greater(t, len(notes), 0, "Должны быть созданы заметки до остановки")
}

// TestService_LimitTenNotes тестирует ограничение в 10 заметок
func TestService_LimitTenNotes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := ram.NewRamRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем, пока будут созданы все заметки (ограничение в 10)
	time.Sleep(100 * time.Millisecond)

	// Проверяем что было создано ровно 10 заметок
	assert.Len(t, repo.GetAllNotes(), 10, "Должно быть создано ровно 10 заметок")
}

// TestService_NoteTitles тестирует корректность заголовков заметок
func TestService_NoteTitles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := ram.NewRamRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание нескольких заметок
	time.Sleep(30 * time.Millisecond)

	// Проверяем заголовки
	notes := repo.GetAllNotes()
	require.GreaterOrEqual(t, len(notes), 3, "Должно быть создано как минимум 3 заметки")

	for i := 0; i < len(notes) && i < 3; i++ {
		expectedTitle := fmt.Sprintf("Тестовая заметка %d", i+1)
		assert.Equal(t, expectedTitle, notes[i].GetTitle(),
			"Заголовок заметки %d должен быть '%s'", i+1, expectedTitle)
	}
}

// TestService_ConcurrentSafety тестирует безопасность конкурентного доступа
func TestService_ConcurrentSafety(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := ram.NewRamRepository()

	// Создаем сервис
	service := NewService(repo, ctx, 10*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем немного времени
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что сервис работает без паники и создает заметки
	notesCount := repo.GetNotesCount()
	assert.Greater(t, notesCount, 0, "Должны быть созданы заметки")
	t.Logf("Создано заметок: %d", notesCount)
}

// TestService_ImmediateStop тестирует немедленную остановку сервиса
func TestService_ImmediateStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Останавливаем сервис сразу же
	cancel()

	repo := ram.NewRamRepository()
	service := NewService(repo, ctx, 10*time.Millisecond)
	service.Start()

	// Даем время на обработку
	time.Sleep(20 * time.Millisecond)

	// Проверяем, что не было создано заметок
	notes := repo.GetAllNotes()
	assert.Len(t, notes, 0, "Не должно быть созданных заметок при немедленной остановке")
}

// TestService_NoteContent тестирует содержимое заметок
func TestService_NoteContent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := ram.NewRamRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание заметок
	time.Sleep(40 * time.Millisecond)

	// Проверяем содержимое заметок
	notes := repo.GetAllNotes()
	require.GreaterOrEqual(t, len(notes), 3, "Должно быть создано как минимум 3 заметки")

	for i := 0; i < len(notes) && i < 3; i++ {
		expectedContent := fmt.Sprintf("Это содержимое тестовой заметки номер %d", i+1)
		assert.Equal(t, expectedContent, notes[i].GetContent(),
			"Содержимое заметки %d должно быть '%s'", i+1, expectedContent)
	}
}

// TestService_SimpleCase тестирует простой сценарий работы сервиса
func TestService_SimpleCase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := ram.NewRamRepository()
	service := NewService(repo, ctx, 20*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание хотя бы одной заметки
	time.Sleep(30 * time.Millisecond)

	// Проверяем, что создана хотя бы одна заметка
	assert.GreaterOrEqual(t, repo.GetNotesCount(), 1, "Должна быть создана хотя бы одна заметка")
}

// TestService_MultipleInstances тестирует работу нескольких экземпляров сервиса
func TestService_MultipleInstances(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := ram.NewRamRepository()

	// Создаем два сервиса
	service1 := NewService(repo, ctx, 15*time.Millisecond)
	service2 := NewService(repo, ctx, 20*time.Millisecond)

	// Запускаем оба сервиса
	service1.Start()
	service2.Start()

	// Ждем некоторое время
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что оба сервиса создают заметки
	notesCount := repo.GetNotesCount()
	assert.Greater(t, notesCount, 0, "Должны быть созданы заметки от обоих сервисов")
	t.Logf("Создано заметок от обоих сервисов: %d", notesCount)
}

// TestService_GoroutineSafety тестирует безопасность при одновременном доступе к репозиторию
func TestService_GoroutineSafety(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := ram.NewRamRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Одновременно читаем заметки из репозитория в нескольких горутинах
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = repo.GetAllNotes()
				time.Sleep(2 * time.Millisecond)
			}
		}()
	}

	// Ждем выполнения чтения
	time.Sleep(50 * time.Millisecond)

	// Отменяем контекст для остановки сервиса
	cancel()

	// Даем время на завершение
	time.Sleep(20 * time.Millisecond)

	// Завершаем горутины чтения
	wg.Wait()

	// Проверяем, что заметки были созданы
	notesCount := repo.GetNotesCount()
	assert.Greater(t, notesCount, 0, "Должны быть созданы заметки")
}

// TestService_DataGenerationAndSaving тестирует полный цикл генерации и сохранения данных
func TestService_DataGenerationAndSaving(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := ram.NewRamRepository()
	service := NewService(repo, ctx, 10*time.Millisecond)

	// Сохраняем начальное количество заметок
	initialCount := repo.GetNotesCount()

	// Запускаем сервис
	service.Start()

	// Ждем создание нескольких заметок
	time.Sleep(60 * time.Millisecond)

	// Проверяем, что количество заметок увеличилось
	finalCount := repo.GetNotesCount()
	assert.Greater(t, finalCount, initialCount, "Количество заметок должно увеличиться после запуска сервиса")

	// Проверяем, что все заметки имеют корректные ID
	notes := repo.GetAllNotes()
	for _, note := range notes {
		assert.NotEmpty(t, note.GetID(), "У заметки должен быть ID")
		assert.Equal(t, "note", note.GetType(), "Тип заметки должен быть 'note'")
	}
}

// TestService_StopWithContext_WithJSON тестирует остановку генерации через контекст с JSON репозиторием
func TestService_StopWithContext_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()
	service := NewService(repo, ctx, 10*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание хотя бы одной заметки
	time.Sleep(15 * time.Millisecond)

	// Останавливаем сервис
	cancel()

	// Ждем завершения
	time.Sleep(20 * time.Millisecond)

	// Проверяем, что сервис остановился и создал заметки до остановки
	notes := repo.GetAllNotes()
	assert.Greater(t, len(notes), 0, "Должны быть созданы заметки до остановки")

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// TestService_LimitTenNotes_WithJSON тестирует ограничение в 10 заметок с JSON репозиторием
func TestService_LimitTenNotes_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем, пока будут созданы все заметки (ограничение в 10)
	time.Sleep(60 * time.Millisecond)

	// Проверяем что было создано ровно 10 заметок
	assert.Len(t, repo.GetAllNotes(), 10, "Должно быть создано ровно 10 заметок")

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// TestService_NoteTitles_WithJSON тестирует корректность заголовков заметок с JSON репозиторием
func TestService_NoteTitles_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание нескольких заметок
	time.Sleep(30 * time.Millisecond)

	// Проверяем заголовки
	notes := repo.GetAllNotes()
	require.GreaterOrEqual(t, len(notes), 3, "Должно быть создано как минимум 3 заметки")

	for i := 0; i < len(notes) && i < 3; i++ {
		expectedTitle := fmt.Sprintf("Тестовая заметка %d", i+1)
		assert.Equal(t, expectedTitle, notes[i].GetTitle(),
			"Заголовок заметки %d должен быть '%s'", i+1, expectedTitle)
	}

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// TestService_ConcurrentSafety_WithJSON тестирует безопасность конкурентного доступа с JSON репозиторием
func TestService_ConcurrentSafety_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()

	// Создаем сервис
	service := NewService(repo, ctx, 10*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем немного времени
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что сервис работает без паники и создает заметки
	notesCount := repo.GetNotesCount()
	assert.Greater(t, notesCount, 0, "Должны быть созданы заметки")
	t.Logf("Создано заметок: %d", notesCount)

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// TestService_ImmediateStop_WithJSON тестирует немедленную остановку сервиса с JSON репозиторием
func TestService_ImmediateStop_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Останавливаем сервис сразу же
	cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()
	service := NewService(repo, ctx, 10*time.Millisecond)
	service.Start()

	// Даем время на обработку
	time.Sleep(20 * time.Millisecond)

	// Проверяем, что не было создано заметок
	notes := repo.GetAllNotes()
	assert.Len(t, notes, 0, "Не должно быть созданных заметок при немедленной остановке")

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// TestService_NoteContent_WithJSON тестирует содержимое заметок с JSON репозиторием
func TestService_NoteContent_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание заметок
	time.Sleep(40 * time.Millisecond)

	// Проверяем содержимое заметок
	notes := repo.GetAllNotes()
	require.GreaterOrEqual(t, len(notes), 3, "Должно быть создано как минимум 3 заметки")

	for i := 0; i < len(notes) && i < 3; i++ {
		expectedContent := fmt.Sprintf("Это содержимое тестовой заметки номер %d", i+1)
		assert.Equal(t, expectedContent, notes[i].GetContent(),
			"Содержимое заметки %d должно быть '%s'", i+1, expectedContent)
	}

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// TestService_SimpleCase_WithJSON тестирует простой сценарий работы сервиса с JSON репозиторием
func TestService_SimpleCase_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()
	service := NewService(repo, ctx, 20*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание хотя бы одной заметки
	time.Sleep(30 * time.Millisecond)

	// Проверяем, что создана хотя бы одна заметка
	assert.GreaterOrEqual(t, repo.GetNotesCount(), 1, "Должна быть создана хотя бы одна заметка")

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// TestService_MultipleInstances_WithJSON тестирует работу нескольких экземпляров сервиса с JSON репозиторием
func TestService_MultipleInstances_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()

	// Создаем два сервиса
	service1 := NewService(repo, ctx, 15*time.Millisecond)
	service2 := NewService(repo, ctx, 20*time.Millisecond)

	// Запускаем оба сервиса
	service1.Start()
	service2.Start()

	// Ждем некоторое время
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что оба сервиса создают заметки
	notesCount := repo.GetNotesCount()
	assert.Greater(t, notesCount, 0, "Должны быть созданы заметки от обоих сервисов")
	t.Logf("Создано заметок от обоих сервисов: %d", notesCount)

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// TestService_GoroutineSafety_WithJSON тестирует безопасность при одновременном доступе к JSON репозиторию
func TestService_GoroutineSafety_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Одновременно читаем заметки из репозитория в нескольких горутинах
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = repo.GetAllNotes()
				time.Sleep(2 * time.Millisecond)
			}
		}()
	}

	// Ждем выполнения чтения
	time.Sleep(50 * time.Millisecond)

	// Отменяем контекст для остановки сервиса
	cancel()

	// Даем время на завершение
	time.Sleep(20 * time.Millisecond)

	// Завершаем горутины чтения
	wg.Wait()

	// Проверяем, что заметки были созданы
	notesCount := repo.GetNotesCount()
	assert.Greater(t, notesCount, 0, "Должны быть созданы заметки")

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// TestService_DataGenerationAndSaving_WithJSON тестирует полный цикл генерации и сохранения данных с JSON репозиторием
func TestService_DataGenerationAndSaving_WithJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем временный JSON репозиторий
	repo := fs.NewJSONRepository()
	service := NewService(repo, ctx, 10*time.Millisecond)

	// Сохраняем начальное количество заметок
	initialCount := repo.GetNotesCount()

	// Запускаем сервис
	service.Start()

	// Ждем создание нескольких заметок
	time.Sleep(60 * time.Millisecond)

	// Проверяем, что количество заметок увеличилось
	finalCount := repo.GetNotesCount()
	assert.Greater(t, finalCount, initialCount, "Количество заметок должно увеличиться после запуска сервиса")

	// Проверяем, что все заметки имеют корректные ID
	notes := repo.GetAllNotes()
	for _, note := range notes {
		assert.NotEmpty(t, note.GetID(), "У заметки должен быть ID")
		assert.Equal(t, "note", note.GetType(), "Тип заметки должен быть 'note'")
	}

	// Очищаем файлы после теста
	_ = cleanupJSONFiles()
}

// cleanupJSONFiles удаляет все JSON файлы из директории data
func cleanupJSONFiles() error {
	files, err := os.ReadDir(fs.StorageDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" &&
			len(file.Name()) > len(fs.NoteFilePrefix) &&
			file.Name()[:len(fs.NoteFilePrefix)] == fs.NoteFilePrefix {
			filePath := filepath.Join(fs.StorageDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}
