package repository

import (
	"fmt"
	"log"

	"github.com/rd2w/go-notes/internal/model"
)

// Repository управляет хранением различных сущностей
type Repository struct {
	notes []*model.Note
	// В будущем нужно добавить другие слайсы для других сущностей
}

// NewRepository создает новый экземпляр репозитория
func NewRepository() *Repository {
	return &Repository{
		notes: make([]*model.Note, 0),
	}
}

// Save принимает интерфейс Entity и сохраняет в соответствующий слайс
func (r *Repository) Save(entity model.Entity) error {
	// Проверяем тип сущности и сохраняем в соответствующий слайс
	switch entity := entity.(type) {
	case *model.Note:
		r.notes = append(r.notes, entity)
		log.Printf("Заметка сохранена: ID=%s, Title=%s", entity.GetID(), entity.GetTitle())
	default:
		return fmt.Errorf("неподдерживаемый тип сущности: %T", entity)
	}
	return nil
}

// GetAllNotes возвращает все сохраненные заметки
func (r *Repository) GetAllNotes() []*model.Note {
	return r.notes
}

// GetNotesCount возвращает количество сохраненных заметок
func (r *Repository) GetNotesCount() int {
	return len(r.notes)
}
