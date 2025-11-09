package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
)

const (
	StorageDir    = "data"
	NotesFileName = "notes.json"
	UsersFileName = "users.json"
	TimeFormat    = "2006-01-02_15-04-05"
)

// Тестовые переменные, которые можно изменить в тестах
var (
	TestStorageDir    = StorageDir
	TestNotesFileName = NotesFileName
	TestUsersFileName = UsersFileName
	TestTimeFormat    = TimeFormat
)

// JSONRepository реализация репозитория с хранением данных в JSON файлах
type JSONRepository struct {
	notes             []*model.Note
	notesIndex        map[string]*model.Note
	users             []*model.User
	usersIndex        map[string]*model.User
	entities          map[string][]repository.Entity // Хранит все сущности по типу
	entityIndex       map[string]repository.Entity   // Для быстрого поиска по ID
	mu                sync.RWMutex
	initialNotesCount int // Количество заметок, загруженных из файла при инициализации
	initialUsersCount int // Количество пользователей, загруженных из файла при инициализации
}

// NewJSONRepository создает новый экземпляр JSON репозитория (возвращает интерфейс)
func NewJSONRepository() repository.Repository {
	repo := &JSONRepository{
		notes:             make([]*model.Note, 0),
		notesIndex:        make(map[string]*model.Note),
		users:             make([]*model.User, 0),
		usersIndex:        make(map[string]*model.User),
		entities:          make(map[string][]repository.Entity),
		entityIndex:       make(map[string]repository.Entity),
		initialNotesCount: 0,
		initialUsersCount: 0,
	}

	// Загружаем данные из хранилища при создании репозитория
	repo.LoadFromStorage()

	return repo
}

// Save сохраняет сущность в соответствующий слайс и в JSON файл
func (r *JSONRepository) Save(entity repository.Entity) {
	log.Printf("Репозиторий: вызван метод Save для сущности типа %s с ID=%s", entity.GetType(), entity.GetID())
	r.mu.Lock()
	defer r.mu.Unlock()

	switch e := entity.(type) {
	case *model.Note:
		// Добавляем заметку в слайс и индекс
		r.notes = append(r.notes, e)
		r.notesIndex[e.GetID()] = e

		log.Printf("Репозиторий: сохранена заметка ID=%s", e.GetID())

		// Сохраняем в JSON файл
		if err := r.saveNotesToJSON(); err != nil {
			log.Printf("Ошибка при сохранении заметки в JSON: %v", err)
		}
	case *model.User:
		// Добавляем пользователя в слайс и индекс
		r.users = append(r.users, e)
		r.usersIndex[e.GetID()] = e

		log.Printf("Репозиторий: сохранен пользователь ID=%s", e.GetID())

		// Сохраняем в JSON файл
		if err := r.saveUsersToJSON(); err != nil {
			log.Printf("Ошибка при сохранении пользователя в JSON: %v", err)
		}
	default:
		// Для других типов сущностей
		entityType := entity.GetType()
		r.entities[entityType] = append(r.entities[entityType], entity)
		r.entityIndex[entity.GetID()] = entity

		log.Printf("Репозиторий: сохранена сущность %s ID=%s", entityType, entity.GetID())

		// Сохраняем в JSON файл
		if err := r.saveEntitiesToJSON(entityType); err != nil {
			log.Printf("Ошибка при сохранении сущности %s в JSON: %v", entityType, err)
		}
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

	// Обрабатываем отрицательные индексы как 0
	if lastIndex < 0 {
		lastIndex = 0
	}

	if lastIndex >= len(r.notes) {
		return []*model.Note{}
	}

	newNotes := r.notes[lastIndex:]
	result := make([]*model.Note, len(newNotes))
	copy(result, newNotes)
	return result
}

// LoadFromStorage загружает данные из JSON файлов при старте приложения
func (r *JSONRepository) LoadFromStorage() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Загружаем заметки
	r.loadNotesFromStorage()

	// Загружаем пользователей
	r.loadUsersFromStorage()
}

// loadNotesFromStorage загружает заметки из JSON файла
func (r *JSONRepository) loadNotesFromStorage() {
	// Находим файл с данными для заметок
	noteFile := filepath.Join(TestStorageDir, TestNotesFileName)

	// Проверяем, существует ли файл
	if _, err := os.Stat(noteFile); os.IsNotExist(err) {
		// Если файл не найден, начинаем с пустого репозитория
		log.Printf("Файл с заметками %s не найден, начнем с пустого репозитория", noteFile)
		return
	}

	// Загружаем данные из файла
	if err := r.loadNotesFromJSONFile(noteFile); err != nil {
		fmt.Printf("Ошибка при загрузке заметок из файла %s: %v\n", noteFile, err)
		return
	}

	fmt.Printf("Загружено %d заметок из файла %s\n", len(r.notes), noteFile)

	// Сохраняем количество загруженных заметок как начальное
	r.initialNotesCount = len(r.notes)
}

// loadUsersFromStorage загружает пользователей из JSON файла
func (r *JSONRepository) loadUsersFromStorage() {
	// Находим файл с данными для пользователей
	userFile := filepath.Join(TestStorageDir, TestUsersFileName)

	// Проверяем, существует ли файл
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		// Если файл не найден, начинаем с пустого репозитория
		log.Printf("Файл с пользователями %s не найден, начнем с пустого репозитория", userFile)
		return
	}

	// Загружаем данные из файла
	if err := r.loadUsersFromJSONFile(userFile); err != nil {
		fmt.Printf("Ошибка при загрузке пользователей из файла %s: %v\n", userFile, err)
		return
	}

	fmt.Printf("Загружено %d пользователей из файла %s\n", len(r.users), userFile)

	// Сохраняем количество загруженных пользователей как начальное
	r.initialUsersCount = len(r.users)
}

// saveNotesToJSON сохраняет заметки в JSON файл
func (r *JSONRepository) saveNotesToJSON() error {
	// Создаем директорию, если она не существует
	if err := os.MkdirAll(TestStorageDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %w", TestStorageDir, err)
	}

	// Формируем имя файла
	filename := TestNotesFileName
	filePath := filepath.Join(TestStorageDir, filename)

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
		return fmt.Errorf("ошибка при кодировании заметок в JSON: %w", err)
	}

	return nil
}

// saveUsersToJSON сохраняет пользователей в JSON файл
func (r *JSONRepository) saveUsersToJSON() error {
	// Создаем директорию, если она не существует
	if err := os.MkdirAll(TestStorageDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %w", TestStorageDir, err)
	}

	// Формируем имя файла
	filename := TestUsersFileName
	filePath := filepath.Join(TestStorageDir, filename)

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

	// Кодируем данные в JSON, включая пароли
	usersWithPasswords := make([]json.RawMessage, len(r.users))
	for i, user := range r.users {
		userData, err := user.MarshalJSONWithPassword()
		if err != nil {
			return fmt.Errorf("ошибка при кодировании пользователя %s: %w", user.GetID(), err)
		}
		usersWithPasswords[i] = userData
	}

	// Кодируем данные в JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(usersWithPasswords); err != nil {
		return fmt.Errorf("ошибка при кодировании пользователей в JSON: %w", err)
	}

	return nil
}

// saveEntitiesToJSON сохраняет сущности указанного типа в JSON файл
func (r *JSONRepository) saveEntitiesToJSON(entityType string) error {
	// Создаем директорию, если она не существует
	if err := os.MkdirAll(TestStorageDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию %s: %w", TestStorageDir, err)
	}

	// Формируем имя файла
	filename := fmt.Sprintf("%ss.json", entityType) // например, "items.json"
	filePath := filepath.Join(TestStorageDir, filename)

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
	if err := encoder.Encode(r.entities[entityType]); err != nil {
		return fmt.Errorf("ошибка при кодировании сущностей %s в JSON: %w", entityType, err)
	}

	return nil
}

// loadNotesFromJSONFile загружает заметки из указанного JSON файла
func (r *JSONRepository) loadNotesFromJSONFile(filepath string) error {
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

// loadUsersFromJSONFile загружает пользователей из указанного JSON файла
func (r *JSONRepository) loadUsersFromJSONFile(filepath string) error {
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
	var jsonUsers []json.RawMessage
	if err := json.Unmarshal(byteValue, &jsonUsers); err != nil {
		return fmt.Errorf("ошибка при декодировании JSON: %w", err)
	}

	// Преобразуем каждый JSON объект в User
	users := make([]*model.User, len(jsonUsers))
	for i, rawUser := range jsonUsers {
		user := &model.User{}
		if err := json.Unmarshal(rawUser, user); err != nil {
			return fmt.Errorf("ошибка при десериализации пользователя: %w", err)
		}
		users[i] = user
	}

	// Обновляем внутренние структуры
	r.users = users
	r.usersIndex = make(map[string]*model.User)
	for _, user := range users {
		r.usersIndex[user.GetID()] = user
	}

	return nil
}

// GetAllByType возвращает все сущности указанного типа
func (r *JSONRepository) GetAllByType(entityType string) []repository.Entity {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch entityType {
	case "note":
		entities := make([]repository.Entity, len(r.notes))
		for i, note := range r.notes {
			entities[i] = note
		}
		return entities
	case "user":
		entities := make([]repository.Entity, len(r.users))
		for i, user := range r.users {
			entities[i] = user
		}
		return entities
	default:
		entities, exists := r.entities[entityType]
		if !exists {
			return []repository.Entity{}
		}

		result := make([]repository.Entity, len(entities))
		copy(result, entities)
		return result
	}
}

// GetByID возвращает сущность по типу и ID
func (r *JSONRepository) GetByID(entityType, id string) repository.Entity {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch entityType {
	case "note":
		if note, exists := r.notesIndex[id]; exists {
			return note
		}
	case "user":
		if user, exists := r.usersIndex[id]; exists {
			return user
		}
	default:
		if entity, exists := r.entityIndex[id]; exists {
			// Проверяем, что тип сущности совпадает
			if entity.GetType() == entityType {
				return entity
			}
		}
	}

	return nil
}

// DeleteByID удаляет сущность по типу и ID
func (r *JSONRepository) DeleteByID(entityType, id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch entityType {
	case "note":
		// Удаляем заметку
		if note, exists := r.notesIndex[id]; exists {
			// Удаляем из слайса
			for i, n := range r.notes {
				if n.GetID() == note.GetID() {
					r.notes = append(r.notes[:i], r.notes[i+1:]...)
					break
				}
			}
			// Удаляем из индекса
			delete(r.notesIndex, id)

			// Сохраняем изменения в файл
			if err := r.saveNotesToJSON(); err != nil {
				log.Printf("Ошибка при сохранении заметок после удаления: %v", err)
			}
			return true
		}
	case "user":
		// Удаляем пользователя
		if user, exists := r.usersIndex[id]; exists {
			// Удаляем из слайса
			for i, u := range r.users {
				if u.GetID() == user.GetID() {
					r.users = append(r.users[:i], r.users[i+1:]...)
					break
				}
			}
			// Удаляем из индекса
			delete(r.usersIndex, id)

			// Сохраняем изменения в файл
			if err := r.saveUsersToJSON(); err != nil {
				log.Printf("Ошибка при сохранении пользователей после удаления: %v", err)
			}
			return true
		}
	default:
		// Удаляем другую сущность
		if entity, exists := r.entityIndex[id]; exists && entity.GetType() == entityType {
			// Удаляем из слайса соответствующего типа
			entities := r.entities[entityType]
			for i, e := range entities {
				if e.GetID() == id {
					r.entities[entityType] = append(entities[:i], entities[i+1:]...)
					break
				}
			}
			// Удаляем из общего индекса
			delete(r.entityIndex, id)

			// Сохраняем изменения в файл
			if err := r.saveEntitiesToJSON(entityType); err != nil {
				log.Printf("Ошибка при сохранении сущностей %s после удаления: %v", entityType, err)
			}
			return true
		}
	}

	return false
}
