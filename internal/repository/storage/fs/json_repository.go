package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
)

const (
	StorageDir     = "data"
	NoteFilePrefix = "notes"
	TimeFormat     = "2006-01-02_15-04-05"
)

// JSONRepository реализация репозитория с хранением данных в JSON файлах
type JSONRepository struct {
	notes      []*model.Note
	notesIndex map[string]*model.Note
	mu         sync.RWMutex
	startTime  time.Time // Время запуска репозитория для формирования имени файла
}

// NewJSONRepository создает новый экземпляр JSON репозитория (возвращает интерфейс)
func NewJSONRepository() repository.Repository {
	// Проверяем, есть ли уже существующие файлы с заметками
	latestTime := findLatestNoteFileTime()
	if latestTime.IsZero() {
		// Если файлов нет, используем текущее время
		latestTime = time.Now()
	}

	repo := &JSONRepository{
		notes:      make([]*model.Note, 0),
		notesIndex: make(map[string]*model.Note),
		startTime:  latestTime,
	}

	// Загружаем данные из хранилища при создании репозитория
	repo.LoadFromStorage()

	return repo
}

// findLatestNoteFileTime находит время самого последнего файла с заметками
func findLatestNoteFileTime() time.Time {
	// Проверяем, существует ли директория
	if _, err := os.Stat(StorageDir); os.IsNotExist(err) {
		return time.Time{}
	}

	// Читаем содержимое директории
	files, err := os.ReadDir(StorageDir)
	if err != nil {
		log.Printf("Ошибка при чтении директории %s: %v", StorageDir, err)
		return time.Time{}
	}

	var latestTime time.Time

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if filepath.Ext(filename) == ".json" && len(filename) > len(NoteFilePrefix) && filename[:len(NoteFilePrefix)] == NoteFilePrefix {
			// Извлекаем временную метку из имени файла
			timeStr := filename[len(NoteFilePrefix)+1 : len(filename)-5] // убираем префикс_ и .json
			if fileTime, err := time.Parse(TimeFormat, timeStr); err == nil {
				if fileTime.After(latestTime) {
					latestTime = fileTime
				}
			}
		}
	}

	return latestTime
}

// Save сохраняет сущность в соответствующий слайс и в JSON файл
func (r *JSONRepository) Save(entity repository.Entity) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch entity := entity.(type) {
	case *model.Note:
		// Добавляем заметку в слайс и индекс
		r.notes = append(r.notes, entity)
		r.notesIndex[entity.GetID()] = entity

		log.Printf("Репозиторий: сохранена заметка ID=%s", entity.GetID())

		// Сохраняем в JSON файл
		if err := r.saveToJSON(); err != nil {
			log.Printf("Ошибка при сохранении в JSON: %v", err)
		}
	default:
		log.Printf("Репозиторий: неподдерживаемый тип сущности: %T", entity)
	}
}

// GetAllNotes возвращает все сохраненные заметки
func (r *JSONRepository) GetAllNotes() []*model.Note {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notes := make([]*model.Note, len(r.notes))
	copy(notes, r.notes)
	return notes
}

// GetNotesCount возвращает количество сохраненных заметок
func (r *JSONRepository) GetNotesCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.notes)
}

// GetNewNotes возвращает заметки, добавленные после указанного индекса
func (r *JSONRepository) GetNewNotes(lastIndex int) []*model.Note {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if lastIndex >= len(r.notes) {
		return []*model.Note{}
	}

	newNotes := r.notes[lastIndex:]
	result := make([]*model.Note, len(newNotes))
	copy(result, newNotes)
	return result
}

// LoadFromStorage загружает данные из JSON файла при старте приложения
func (r *JSONRepository) LoadFromStorage() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Находим файл с данными для заметок
	noteFile, err := r.findLatestFile(NoteFilePrefix)
	if err != nil {
		// Если файл не найден, начинаем с пустого репозитория
		log.Printf("Файл с данными не найден, начнем с пустого репозитория: %v", err)
		return
	}

	// Загружаем данные из файла
	if err := r.loadFromJSONFile(noteFile); err != nil {
		fmt.Printf("Ошибка при загрузке данных из файла %s: %v\n", noteFile, err)
		return
	}

	// Обновляем startTime на основе времени файла
	filename := filepath.Base(noteFile)
	timeStr := filename[len(NoteFilePrefix)+1 : len(filename)-5] // убираем префикс_ и .json
	if fileTime, err := time.Parse(TimeFormat, timeStr); err == nil {
		r.startTime = fileTime
	}

	fmt.Printf("Загружено %d заметок из файла %s\n", len(r.notes), noteFile)
}

// saveToJSON сохраняет данные в JSON файл с временной меткой запуска приложения
func (r *JSONRepository) saveToJSON() error {
	// Создаем директорию, если она не существует
	if err := os.MkdirAll(StorageDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %w", StorageDir, err)
	}

	// Формируем имя файла с временной меткой запуска приложения
	timestamp := r.startTime.Format(TimeFormat)
	filename := fmt.Sprintf("%s_%s.json", NoteFilePrefix, timestamp)
	filePath := filepath.Join(StorageDir, filename)

	// Создаем/перезаписываем файл
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл %s: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Ошибка при закрытии файла %s: %v", filePath, closeErr)
		}
	}()

	// Кодируем данные в JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r.notes); err != nil {
		return fmt.Errorf("ошибка при кодировании в JSON: %w", err)
	}

	return nil
}

// loadFromJSONFile загружает данные из указанного JSON файла
func (r *JSONRepository) loadFromJSONFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл %s: %w", filepath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Ошибка при закрытии файла %s: %v", filepath, closeErr)
		}
	}()

	// Читаем содержимое файла
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("ошибка при чтении файла: %w", err)
	}

	// Декодируем данные из JSON
	var jsonNotes []json.RawMessage
	if err := json.Unmarshal(byteValue, &jsonNotes); err != nil {
		return fmt.Errorf("ошибка при декодировании JSON: %w", err)
	}

	// Преобразуем каждый JSON объект в Note
	notes := make([]*model.Note, len(jsonNotes))
	for i, rawNote := range jsonNotes {
		note := &model.Note{}
		if err := json.Unmarshal(rawNote, note); err != nil {
			return fmt.Errorf("ошибка при десериализации заметки: %w", err)
		}
		notes[i] = note
	}

	// Обновляем внутренние структуры
	r.notes = notes
	r.notesIndex = make(map[string]*model.Note)
	for _, note := range notes {
		r.notesIndex[note.GetID()] = note
	}

	return nil
}

// findLatestFile находит файл с указанным префиксом, используя время запуска приложения
func (r *JSONRepository) findLatestFile(prefix string) (string, error) {
	// Проверяем, существует ли директория
	if _, err := os.Stat(StorageDir); os.IsNotExist(err) {
		return "", fmt.Errorf("директория %s не существует", StorageDir)
	}

	// Формируем имя файла на основе времени запуска
	timestamp := r.startTime.Format(TimeFormat)
	filename := fmt.Sprintf("%s_%s.json", prefix, timestamp)
	filePath := filepath.Join(StorageDir, filename)

	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("файл %s не существует", filePath)
	}

	return filePath, nil
}
