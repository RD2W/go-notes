package file

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/domain/repository"
)

// FileRepository реализация интерфейса репозитория с сохранением в файл
type FileRepository struct {
	mu       sync.RWMutex
	notes    map[string]*model.Note
	users    map[string]*model.User
	entity   map[string]map[string]repository.Entity
	filePath string
}

// NewFileRepository создает новый экземпляр репозитория с файловым хранилищем
func NewFileRepository(filePath string) (repository.Repository, error) {
	repo := &FileRepository{
		notes:    make(map[string]*model.Note),
		users:    make(map[string]*model.User),
		entity:   make(map[string]map[string]repository.Entity),
		filePath: filePath,
	}

	// Загружаем данные из файла при инициализации
	err := repo.loadFromFile()
	if err != nil {
		// Если файл не существует, создаем его
		if os.IsNotExist(err) {
			err = repo.saveToFile()
			if err != nil {
				return nil, fmt.Errorf("ошибка создания файла: %v", err)
			}
		} else {
			return nil, fmt.Errorf("ошибка загрузки из файла: %v", err)
		}
	}

	return repo, nil
}

// loadFromFile загружает данные из файла
func (r *FileRepository) loadFromFile() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var fileData struct {
		Notes  map[string]*model.Note            `json:"notes"`
		Users  map[string]*model.User            `json:"users"`
		Entity map[string]map[string]interface{} `json:"entity"`
	}

	err = json.Unmarshal(data, &fileData)
	if err != nil {
		return err
	}

	r.notes = fileData.Notes
	r.users = fileData.Users

	// Преобразуем entity из интерфейсов обратно в сущности
	r.entity = make(map[string]map[string]repository.Entity)
	for entityType, entities := range fileData.Entity {
		r.entity[entityType] = make(map[string]repository.Entity)
		for id, entityData := range entities {
			// Преобразуем данные обратно в соответствующие сущности
			entity := r.convertToEntity(entityType, entityData)
			if entity != nil {
				r.entity[entityType][id] = entity
			}
		}
	}

	return nil
}

// saveToFile сохраняет данные в файл
func (r *FileRepository) saveToFile() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fileData := struct {
		Notes  map[string]*model.Note            `json:"notes"`
		Users  map[string]*model.User            `json:"users"`
		Entity map[string]map[string]interface{} `json:"entity"`
	}{
		Notes:  r.notes,
		Users:  r.users,
		Entity: make(map[string]map[string]interface{}),
	}

	for entityType, entities := range r.entity {
		fileData.Entity[entityType] = make(map[string]interface{})
		for id, entity := range entities {
			fileData.Entity[entityType][id] = entity
		}
	}

	data, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.filePath, data, 0644)
}

// convertToEntity преобразует интерфейс обратно в сущность
func (r *FileRepository) convertToEntity(entityType string, data interface{}) repository.Entity {
	switch entityType {
	case "note":
		if noteData, ok := data.(*model.Note); ok {
			return noteData
		}
	case "user":
		if userData, ok := data.(*model.User); ok {
			return userData
		}
	}
	return nil
}

// Save сохраняет сущность
func (r *FileRepository) Save(entity repository.Entity) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entityType := entity.GetType()
	if r.entity[entityType] == nil {
		r.entity[entityType] = make(map[string]repository.Entity)
	}
	r.entity[entityType][entity.GetID()] = entity

	// Обновляем специфичные мапы для удобства
	switch entity.GetType() {
	case "note":
		if note, ok := entity.(*model.Note); ok {
			r.notes[entity.GetID()] = note
		}
	case "user":
		if user, ok := entity.(*model.User); ok {
			r.users[entity.GetID()] = user
		}
	}

	// Сохраняем в файл
	r.saveToFile()
}

// GetAllNotes возвращает все заметки
func (r *FileRepository) GetAllNotes() []*model.Note {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notes := make([]*model.Note, 0, len(r.notes))
	for _, note := range r.notes {
		notes = append(notes, note)
	}
	return notes
}

// GetNotesCount возвращает количество заметок
func (r *FileRepository) GetNotesCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.notes)
}

// GetNewNotes возвращает новые заметки начиная с указанного индекса
func (r *FileRepository) GetNewNotes(lastIndex int) []*model.Note {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notes := r.GetAllNotes()
	if lastIndex >= len(notes) {
		return []*model.Note{}
	}
	return notes[lastIndex:]
}

// GetAllByType возвращает все сущности указанного типа
func (r *FileRepository) GetAllByType(entityType string) []repository.Entity {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entities := make([]repository.Entity, 0)
	if entityMap, exists := r.entity[entityType]; exists {
		for _, entity := range entityMap {
			entities = append(entities, entity)
		}
	}
	return entities
}

// GetByID возвращает сущность по ID
func (r *FileRepository) GetByID(entityType string, id string) repository.Entity {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if entityMap, exists := r.entity[entityType]; exists {
		if entity, exists := entityMap[id]; exists {
			return entity
		}
	}
	return nil
}

// DeleteByID удаляет сущность по ID
func (r *FileRepository) DeleteByID(entityType string, id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entityMap, exists := r.entity[entityType]; exists {
		if _, exists := entityMap[id]; exists {
			delete(entityMap, id)

			// Удаляем из специфичных мапов
			switch entityType {
			case "note":
				delete(r.notes, id)
			case "user":
				delete(r.users, id)
			}

			// Сохраняем в файл
			r.saveToFile()
			return true
		}
	}
	return false
}
