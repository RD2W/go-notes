package ram

import (
	"log"
	"sync"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
)

// RamRepository управляет хранением различных сущностей в памяти
type RamRepository struct {
	notes       []*model.Note
	notesIndex  map[string]*model.Note         // Для быстрого поиска по ID
	entities    map[string][]repository.Entity // Хранит все сущности по типу
	entityIndex map[string]repository.Entity   // Для быстрого поиска по ID
	mu          sync.RWMutex
}

// NewRamRepository создает новый экземпляр RAM репозитория (возвращает интерфейс)
func NewRamRepository() repository.Repository {
	return &RamRepository{
		notes:       make([]*model.Note, 0),
		notesIndex:  make(map[string]*model.Note),
		entities:    make(map[string][]repository.Entity),
		entityIndex: make(map[string]repository.Entity),
	}
}

// Save сохраняет сущность в соответствующий слайс
func (r *RamRepository) Save(entity repository.Entity) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch e := entity.(type) {
	case *model.Note:
		r.notes = append(r.notes, e)
		r.notesIndex[e.GetID()] = e
		log.Printf("Репозиторий: сохранена заметка ID=%s", e.GetID())
	case *model.User:
		// Добавляем пользователя в общий массив сущностей
		entityType := e.GetType()
		r.entities[entityType] = append(r.entities[entityType], e)
		r.entityIndex[e.GetID()] = e
		log.Printf("Репозиторий: сохранен пользователь ID=%s", e.GetID())
	default:
		// Для других типов сущностей
		entityType := entity.GetType()
		r.entities[entityType] = append(r.entities[entityType], entity)
		r.entityIndex[entity.GetID()] = entity
		log.Printf("Репозиторий: сохранена сущность %s ID=%s", entityType, entity.GetID())
	}
}

// GetAllNotes возвращает все сохраненные заметки
func (r *RamRepository) GetAllNotes() []*model.Note {
	r.mu.RLock()
	defer r.mu.RUnlock()
	notes := make([]*model.Note, len(r.notes))
	copy(notes, r.notes)
	return notes
}

// GetNotesCount возвращает количество сохраненных заметок
func (r *RamRepository) GetNotesCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.notes)
}

// GetNewNotes возвращает заметки, добавленные после указанного индекса
func (r *RamRepository) GetNewNotes(lastIndex int) []*model.Note {
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

// GetAllByType возвращает все сущности указанного типа
func (r *RamRepository) GetAllByType(entityType string) []repository.Entity {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entities, exists := r.entities[entityType]
	if !exists {
		return []repository.Entity{}
	}

	result := make([]repository.Entity, len(entities))
	copy(result, entities)
	return result
}

// GetByID возвращает сущность по типу и ID
func (r *RamRepository) GetByID(entityType, id string) repository.Entity {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if entityType == "note" {
		// Для заметок проверяем в notesIndex
		if note, exists := r.notesIndex[id]; exists {
			return note
		}
	} else {
		// Для других типов проверяем в entityIndex
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
func (r *RamRepository) DeleteByID(entityType, id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entityType == "note" {
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
			return true
		}
	} else {
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
			return true
		}
	}

	return false
}
