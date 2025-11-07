package ram

import (
	"log"
	"sync"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
)

// RamRepository управляет хранением различных сущностей в памяти
type RamRepository struct {
	notes      []*model.Note
	notesIndex map[string]*model.Note // Для быстрого поиска по ID
	mu         sync.RWMutex
}

// NewRamRepository создает новый экземпляр RAM репозитория (возвращает интерфейс)
func NewRamRepository() repository.Repository {
	return &RamRepository{
		notes:      make([]*model.Note, 0),
		notesIndex: make(map[string]*model.Note),
	}
}

// Save сохраняет сущность в соответствующий слайс
func (r *RamRepository) Save(entity repository.Entity) {
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
