package repository

import (
	"log"
	"sync"

	"github.com/rd2w/go-notes/internal/model"
)

// Entity интерфейс, который должны реализовывать все сущности
type Entity interface {
	GetID() string
	GetType() string
}

// Repository управляет хранением различных сущностей
type Repository struct {
	notes      []*model.Note
	notesIndex map[string]*model.Note // для быстрого поиска по ID
	mu         sync.RWMutex
}

// NewRepository создает новый экземпляр репозитория
func NewRepository() *Repository {
	return &Repository{
		notes:      make([]*model.Note, 0),
		notesIndex: make(map[string]*model.Note),
	}
}

// Save сохраняет сущность в соответствующий слайс
func (r *Repository) Save(entity Entity) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch entity := entity.(type) {
	case *model.Note:
		r.notes = append(r.notes, entity)
		r.notesIndex[entity.GetID()] = entity
		log.Printf("Репозиторий: сохранена заметка ID=%s", entity.GetID())
	default:
		log.Printf("Репозиторий: неподдерживаемый тип сущности: %T", entity)
	}
}

// GetAllNotes возвращает все сохраненные заметки
func (r *Repository) GetAllNotes() []*model.Note {
	r.mu.RLock()
	defer r.mu.RUnlock()
	notes := make([]*model.Note, len(r.notes))
	copy(notes, r.notes)
	return notes
}

// GetNotesCount возвращает количество сохраненных заметок
func (r *Repository) GetNotesCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.notes)
}

// GetNewNotes возвращает заметки, добавленные после указанного индекса
func (r *Repository) GetNewNotes(lastIndex int) []*model.Note {
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
