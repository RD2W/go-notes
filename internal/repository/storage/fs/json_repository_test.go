package fs_test

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository/storage/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// safeBuffer потокобезопасный буфер для логов
type safeBuffer struct {
	buf bytes.Buffer
	mu  sync.RWMutex
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.String()
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

// TestJSONRepository_Save тестирует метод Save с различными типами сущностей
func TestJSONRepository_Save(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Тестируем сохранение заметки
	note := model.NewNote("Test Note", "Test Content")
	repo.Save(note)

	// Проверяем, что заметка была сохранена
	notes := repo.GetAllNotes()
	require.Len(t, notes, 1, "Должна быть одна заметка")
	assert.Equal(t, note.GetID(), notes[0].GetID())
	assert.Equal(t, note.GetTitle(), notes[0].GetTitle())
}

// TestJSONRepository_SaveUnsupportedEntity тестирует обработку неподдерживаемых типов сущностей
func TestJSONRepository_SaveUnsupportedEntity(t *testing.T) {
	// Перехватываем вывод лога для проверки
	var buf safeBuffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Создаем неподдерживаемую сущность
	unsupportedEntity := &mockEntity{id: "test", entityType: "unsupported"}
	repo.Save(unsupportedEntity)

	// Проверяем, что заметки не были сохранены для неподдерживаемых типов
	assert.Equal(t, 0, repo.GetNotesCount(), "Не должно быть сохраненных заметок для неподдерживаемых сущностей")

	// Проверяем, что сущность была сохранена в общий слайс
	entities := repo.GetAllByType("unsupported")
	assert.Len(t, entities, 1, "Должна быть одна неподдерживаемая сущность")

	// Проверяем вывод в лог
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Репозиторий: сохранена сущность")
}

// TestJSONRepository_SaveMultipleNotes тестирует сохранение нескольких заметок
func TestJSONRepository_SaveMultipleNotes(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Сохраняем несколько заметок
	notes := []*model.Note{
		model.NewNote("Note 1", "Content 1"),
		model.NewNote("Note 2", "Content 2"),
		model.NewNote("Note 3", "Content 3"),
	}

	for _, note := range notes {
		repo.Save(note)
	}

	// Проверяем, что все заметки были сохранены
	savedNotes := repo.GetAllNotes()
	assert.Len(t, savedNotes, 3, "Должно быть 3 заметки")

	// Проверяем содержимое заметок
	for i, note := range notes {
		assert.Equal(t, note.GetID(), savedNotes[i].GetID())
		assert.Equal(t, note.GetTitle(), savedNotes[i].GetTitle())
	}
}

// TestJSONRepository_GetAllNotes тестирует метод GetAllNotes
func TestJSONRepository_GetAllNotes(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Добавляем заметки
	note1 := model.NewNote("Note 1", "Content 1")
	note2 := model.NewNote("Note 2", "Content 2")

	repo.Save(note1)
	repo.Save(note2)

	// Тестируем GetAllNotes
	notes := repo.GetAllNotes()
	require.Len(t, notes, 2)

	// Проверяем, что возвращаются копии, а не ссылки на внутренний слайс
	notes[0] = nil // Это не должно повлиять на внутренний слайс репозитория

	internalNotes := repo.GetAllNotes()
	assert.NotNil(t, internalNotes[0], "Изменение возвращенного слайса не должно влиять на репозиторий")
	assert.Equal(t, note1.GetID(), internalNotes[0].GetID())
}

// TestJSONRepository_GetNotesCount тестирует метод GetNotesCount
func TestJSONRepository_GetNotesCount(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Начальное количество должно быть 0
	assert.Equal(t, 0, repo.GetNotesCount())

	// Добавляем заметки и проверяем увеличение счетчика
	note1 := model.NewNote("Note 1", "Content 1")
	repo.Save(note1)
	assert.Equal(t, 1, repo.GetNotesCount())

	note2 := model.NewNote("Note 2", "Content 2")
	repo.Save(note2)
	assert.Equal(t, 2, repo.GetNotesCount())
}

// TestJSONRepository_GetNewNotes тестирует метод GetNewNotes с различными индексами
func TestJSONRepository_GetNewNotes(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Добавляем начальные заметки
	notes := []*model.Note{
		model.NewNote("Note 1", "Content 1"),
		model.NewNote("Note 2", "Content 2"),
		model.NewNote("Note 3", "Content 3"),
	}

	for _, note := range notes {
		repo.Save(note)
	}

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
		{"LastIndex negative", -1, 3}, // при отрицательном индексе должен возвращать все заметки
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newNotes := repo.GetNewNotes(tt.lastIndex)
			assert.Len(t, newNotes, tt.expected)

			// Проверяем, что возвращаются правильные заметки
			if tt.expected > 0 {
				var expectedNoteIndex int
				if tt.lastIndex < 0 {
					expectedNoteIndex = 0 // для отрицательных индексов ожидаем первую заметку
				} else {
					expectedNoteIndex = tt.lastIndex
				}

				if expectedNoteIndex < len(notes) {
					expectedNote := notes[expectedNoteIndex]
					assert.Equal(t, expectedNote.GetID(), newNotes[0].GetID())
				}
			}
		})
	}
}

// TestJSONRepository_GetNewNotesOrder тестирует, что GetNewNotes возвращает заметки в правильном порядке
func TestJSONRepository_GetNewNotesOrder(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Добавляем заметки
	notes := []*model.Note{
		model.NewNote("Note 1", "Content 1"),
		model.NewNote("Note 2", "Content 2"),
		model.NewNote("Note 3", "Content 3"),
	}

	for _, note := range notes {
		repo.Save(note)
	}

	// Получаем новые заметки с индекса 1
	newNotes := repo.GetNewNotes(1)

	// Проверяем, что порядок правильный
	assert.Len(t, newNotes, 2)
	assert.Equal(t, notes[1].GetID(), newNotes[0].GetID())
	assert.Equal(t, notes[2].GetID(), newNotes[1].GetID())
}

// TestJSONRepository_DataIsolation тестирует, что внутренние данные не экспортируются
func TestJSONRepository_DataIsolation(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Добавляем заметку
	note := model.NewNote("Test Note", "Content")
	repo.Save(note)

	// Получаем заметки и изменяем возвращенный слайс
	notes := repo.GetAllNotes()
	originalID := notes[0].GetID()
	notes[0] = nil // Это не должно повлиять на репозиторий

	// Получаем заметки снова - должны быть оригинальные данные
	notesAgain := repo.GetAllNotes()
	assert.NotNil(t, notesAgain[0])
	assert.Equal(t, originalID, notesAgain[0].GetID())
}

// TestJSONRepository_NewNotesIsolation тестирует, что GetNewNotes возвращает копии
func TestJSONRepository_NewNotesIsolation(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Добавляем заметки
	note1 := model.NewNote("Note 1", "Content 1")
	note2 := model.NewNote("Note 2", "Content 2")
	repo.Save(note1)
	repo.Save(note2)

	// Получаем новые заметки и изменяем их
	newNotes := repo.GetNewNotes(0)
	newNotes[0] = nil

	// Проверяем, что данные в репозитории не изменились
	allNotes := repo.GetAllNotes()
	assert.NotNil(t, allNotes[0])
	assert.Equal(t, note1.GetID(), allNotes[0].GetID())
}

// TestJSONRepository_EmptyRepository тестирует поведение репозитория без заметок
func TestJSONRepository_EmptyRepository(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	assert.Equal(t, 0, repo.GetNotesCount())
	assert.Empty(t, repo.GetAllNotes())
	assert.Empty(t, repo.GetNewNotes(0))
	assert.Empty(t, repo.GetNewNotes(5))
}

// TestJSONRepository_NegativeIndex тестирует поведение при отрицательных индексах
func TestJSONRepository_NegativeIndex(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Добавляем заметки
	notes := []*model.Note{
		model.NewNote("Note 1", "Content 1"),
		model.NewNote("Note 2", "Content 2"),
	}

	for _, note := range notes {
		repo.Save(note)
	}

	// Проверяем, что отрицательный индекс ведет себя как 0 (возвращает все заметки)
	allNotes := repo.GetAllNotes()
	negativeIndexNotes := repo.GetNewNotes(-1)

	assert.Len(t, negativeIndexNotes, 2)
	assert.Equal(t, allNotes[0].GetID(), negativeIndexNotes[0].GetID())
	assert.Equal(t, allNotes[1].GetID(), negativeIndexNotes[1].GetID())
}

// TestJSONRepository_SaveToFile тестирует сохранение в файл
func TestJSONRepository_SaveToFile(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()

	// Сохраняем заметку
	note := model.NewNote("Test Note", "Test Content")
	repo.Save(note)

	// Проверяем, что файл был создан
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.NotEmpty(t, files)

	// Проверяем, что файл имеет правильное имя
	var noteFileFound bool
	for _, file := range files {
		if file.Name() == fs.TestNotesFileName {
			noteFileFound = true
			break
		}
	}
	assert.True(t, noteFileFound, "Файл с заметками должен быть создан")
}

// TestJSONRepository_LoadFromStorage тестирует загрузку данных из файла
func TestJSONRepository_LoadFromStorage(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	// Создаем JSON файл с заметками
	note := model.NewNote("Loaded Note", "Loaded Content")
	notes := []*model.Note{note}

	// Формируем имя файла
	filename := filepath.Join(tempDir, fs.TestNotesFileName)

	file, err := os.Create(filename)
	require.NoError(t, err)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(notes)
	require.NoError(t, err)

	if closeErr := file.Close(); closeErr != nil {
		log.Printf("Error closing file %s: %v", filename, closeErr)
	}

	// Создаем новый репозиторий - он должен загрузить данные
	repo := fs.NewJSONRepository()

	// Проверяем, что заметка была загружена
	loadedNotes := repo.GetAllNotes()
	assert.Len(t, loadedNotes, 1)
	assert.Equal(t, note.GetID(), loadedNotes[0].GetID())
	assert.Equal(t, note.GetTitle(), loadedNotes[0].GetTitle())
}

// TestJSONRepository_ConcurrentAccess тестирует конкурентный доступ к репозиторию
func TestJSONRepository_ConcurrentAccess(t *testing.T) {
	// Создаем временный каталог для тестирования
	tempDir := t.TempDir()
	originalStorageDir := fs.TestStorageDir
	fs.TestStorageDir = tempDir
	defer func() { fs.TestStorageDir = originalStorageDir }()

	repo := fs.NewJSONRepository()
	var wg sync.WaitGroup

	// Конкурентные писатели
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			note := model.NewNote("Concurrent Note", "Content")
			repo.Save(note)
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

	// Проверяем, что все заметки были сохранены
	assert.Equal(t, 10, repo.GetNotesCount(), "Все конкурентные записи должны быть обработаны")
}
