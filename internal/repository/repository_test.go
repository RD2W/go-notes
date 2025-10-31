package repository

import (
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepository_Save тестирует метод Save с различными типами сущностей
func TestRepository_Save(t *testing.T) {
	// Перехватываем вывод лога для проверки
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	repo := NewRepository()
	entityChan := make(chan Entity, 10)
	done := make(chan struct{})
	defer close(done)

	// Запускаем горутину метода Save
	go repo.Save(entityChan, done)

	// Тестируем сохранение заметки
	note := model.NewNote("Test Note", "Test Content")
	entityChan <- note

	// Даем время на обработку
	time.Sleep(10 * time.Millisecond)

	// Проверяем, что заметка была сохранена
	notes := repo.GetAllNotes()
	require.Len(t, notes, 1, "Должна быть одна заметка")
	assert.Equal(t, note.GetID(), notes[0].GetID())
	assert.Equal(t, note.GetTitle(), notes[0].GetTitle())

	// Проверяем вывод в лог
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Репозиторий: сохранена заметка ID="+note.GetID())
}

// TestRepository_SaveMultipleNotes тестирует сохранение нескольких заметок
func TestRepository_SaveMultipleNotes(t *testing.T) {
	repo := NewRepository()
	entityChan := make(chan Entity, 10)
	done := make(chan struct{})
	defer close(done)

	go repo.Save(entityChan, done)

	// Сохраняем несколько заметок
	notes := []*model.Note{
		model.NewNote("Note 1", "Content 1"),
		model.NewNote("Note 2", "Content 2"),
		model.NewNote("Note 3", "Content 3"),
	}

	for _, note := range notes {
		entityChan <- note
	}

	// Даем время на обработку
	time.Sleep(20 * time.Millisecond)

	// Проверяем, что все заметки были сохранены
	savedNotes := repo.GetAllNotes()
	assert.Len(t, savedNotes, 3, "Должно быть 3 заметки")

	// Проверяем содержимое заметок
	for i, note := range notes {
		assert.Equal(t, note.GetID(), savedNotes[i].GetID())
		assert.Equal(t, note.GetTitle(), savedNotes[i].GetTitle())
	}
}

// TestRepository_SaveUnsupportedEntity тестирует обработку неподдерживаемых типов сущностей
func TestRepository_SaveUnsupportedEntity(t *testing.T) {
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	repo := NewRepository()
	entityChan := make(chan Entity, 10)
	done := make(chan struct{})
	defer close(done)

	go repo.Save(entityChan, done)

	// Создаем неподдерживаемую сущность
	unsupportedEntity := &mockEntity{id: "test", entityType: "unsupported"}
	entityChan <- unsupportedEntity

	time.Sleep(10 * time.Millisecond)

	// Проверяем, что заметки не были сохранены для неподдерживаемых типов
	assert.Equal(t, 0, repo.GetNotesCount(), "Не должно быть сохраненных заметок для неподдерживаемых сущностей")

	// Проверяем вывод в лог
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Репозиторий: неподдерживаемый тип сущности")
}

// TestRepository_GetAllNotes тестирует метод GetAllNotes
func TestRepository_GetAllNotes(t *testing.T) {
	repo := NewRepository()
	entityChan := make(chan Entity, 10)
	done := make(chan struct{})
	defer close(done)

	go repo.Save(entityChan, done)

	// Добавляем заметки
	note1 := model.NewNote("Note 1", "Content 1")
	note2 := model.NewNote("Note 2", "Content 2")

	entityChan <- note1
	entityChan <- note2

	time.Sleep(10 * time.Millisecond)

	// Тестируем GetAllNotes
	notes := repo.GetAllNotes()
	require.Len(t, notes, 2)

	// Проверяем, что возвращаются копии, а не ссылки на внутренний слайс
	notes[0] = nil // Это не должно повлиять на внутренний слайс репозитория

	internalNotes := repo.GetAllNotes()
	assert.NotNil(t, internalNotes[0], "Изменение возвращенного слайса не должно влиять на репозиторий")
	assert.Equal(t, note1.GetID(), internalNotes[0].GetID())
}

// TestRepository_GetNotesCount тестирует метод GetNotesCount
func TestRepository_GetNotesCount(t *testing.T) {
	repo := NewRepository()
	entityChan := make(chan Entity, 10)
	done := make(chan struct{})
	defer close(done)

	go repo.Save(entityChan, done)

	// Начальное количество должно быть 0
	assert.Equal(t, 0, repo.GetNotesCount())

	// Добавляем заметки и проверяем увеличение счетчика
	note1 := model.NewNote("Note 1", "Content 1")
	entityChan <- note1
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, 1, repo.GetNotesCount())

	note2 := model.NewNote("Note 2", "Content 2")
	entityChan <- note2
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, 2, repo.GetNotesCount())
}

// TestRepository_GetNewNotes тестирует метод GetNewNotes
func TestRepository_GetNewNotes(t *testing.T) {
	repo := NewRepository()
	entityChan := make(chan Entity, 10)
	done := make(chan struct{})
	defer close(done)

	go repo.Save(entityChan, done)

	// Добавляем начальные заметки
	notes := []*model.Note{
		model.NewNote("Note 1", "Content 1"),
		model.NewNote("Note 2", "Content 2"),
		model.NewNote("Note 3", "Content 3"),
	}

	for _, note := range notes {
		entityChan <- note
	}

	time.Sleep(10 * time.Millisecond)

	// Тестируем GetNewNotes с различными индексами
	tests := []struct {
		name      string
		lastIndex int
		expected  int
	}{
		{"LastIndex 0", 0, 3},
		{"LastIndex 1", 1, 2},
		{"LastIndex 2", 2, 1},
		{"LastIndex 3", 3, 0},
		{"LastIndex 5", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newNotes := repo.GetNewNotes(tt.lastIndex)
			assert.Len(t, newNotes, tt.expected)

			// Проверяем, что возвращаются правильные заметки
			if tt.expected > 0 {
				expectedNote := notes[tt.lastIndex]
				assert.Equal(t, expectedNote.GetID(), newNotes[0].GetID())
			}
		})
	}
}

// TestRepository_ConcurrentAccess тестирует конкурентный доступ к репозиторию
func TestRepository_ConcurrentAccess(t *testing.T) {
	repo := NewRepository()
	entityChan := make(chan Entity, 100)
	done := make(chan struct{})
	defer close(done)

	go repo.Save(entityChan, done)

	var wg sync.WaitGroup

	// Конкурентные писатели
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			note := model.NewNote("Concurrent Note", "Content")
			entityChan <- note
		}(i)
	}

	// Конкурентные читатели
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				_ = repo.GetNotesCount()
				_ = repo.GetAllNotes()
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	time.Sleep(10 * time.Millisecond)

	// Проверяем, что все заметки были сохранены
	assert.Equal(t, 10, repo.GetNotesCount(), "Все конкурентные записи должны быть обработаны")
}

// TestRepository_StopWithDone тестирует корректное завершение работы через done канал
func TestRepository_StopWithDone(t *testing.T) {
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	repo := NewRepository()
	entityChan := make(chan Entity, 10)
	done := make(chan struct{})

	// Запускаем горутину Save
	go repo.Save(entityChan, done)

	// Отправляем одну заметку
	note := model.NewNote("Test Note", "Content")
	entityChan <- note
	time.Sleep(5 * time.Millisecond)

	// Останавливаем репозиторий
	close(done)
	time.Sleep(5 * time.Millisecond)

	// Проверяем сообщение о остановке в логах
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Репозиторий: завершение работы")

	// Проверяем, что заметка была сохранена до остановки
	assert.Equal(t, 1, repo.GetNotesCount())
}

// TestRepository_EmptyChannel тестирует поведение с пустым каналом
func TestRepository_EmptyChannel(t *testing.T) {
	repo := NewRepository()
	entityChan := make(chan Entity)
	done := make(chan struct{})
	defer close(done)

	// Должен запуститься и ждать ввода
	go repo.Save(entityChan, done)

	// Даем время на запуск
	time.Sleep(5 * time.Millisecond)

	// Должен продолжать работать без паники
	assert.Equal(t, 0, repo.GetNotesCount())
}

// mockEntity реализует интерфейс Entity для тестирования неподдерживаемых типов
type mockEntity struct {
	id         string
	entityType string
}

func (m *mockEntity) GetID() string {
	return m.id
}

func (m *mockEntity) GetType() string {
	return m.entityType
}

// TestRepository_DataIsolation тестирует, что внутренние данные не экспортируются
func TestRepository_DataIsolation(t *testing.T) {
	repo := NewRepository()
	entityChan := make(chan Entity, 10)
	done := make(chan struct{})
	defer close(done)

	go repo.Save(entityChan, done)

	// Добавляем заметку
	note := model.NewNote("Test Note", "Content")
	entityChan <- note
	time.Sleep(5 * time.Millisecond)

	// Получаем заметки и изменяем возвращенный слайс
	notes := repo.GetAllNotes()
	originalID := notes[0].GetID()
	notes[0] = nil // Это не должно повлиять на репозиторий

	// Получаем заметки снова - должны быть оригинальные данные
	notesAgain := repo.GetAllNotes()
	assert.NotNil(t, notesAgain[0])
	assert.Equal(t, originalID, notesAgain[0].GetID())
}

// TestRepository_NewNotesIsolation тестирует, что GetNewNotes возвращает копии
func TestRepository_NewNotesIsolation(t *testing.T) {
	repo := NewRepository()
	entityChan := make(chan Entity, 10)
	done := make(chan struct{})
	defer close(done)

	go repo.Save(entityChan, done)

	// Добавляем заметки
	note1 := model.NewNote("Note 1", "Content 1")
	note2 := model.NewNote("Note 2", "Content 2")
	entityChan <- note1
	entityChan <- note2
	time.Sleep(10 * time.Millisecond)

	// Получаем новые заметки и изменяем их
	newNotes := repo.GetNewNotes(0)
	newNotes[0] = nil

	// Проверяем, что данные в репозитории не изменились
	allNotes := repo.GetAllNotes()
	assert.NotNil(t, allNotes[0])
	assert.Equal(t, note1.GetID(), allNotes[0].GetID())
}
