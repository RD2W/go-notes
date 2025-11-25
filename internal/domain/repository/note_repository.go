package repository

import (
	"github.com/rd2w/go-notes/internal/domain/model"
)

// NoteRepository интерфейс для работы с заметками
type NoteRepository interface {
	CRUDRepository[*model.Note]
	GetAllNotesByUserID(userID string) ([]*model.Note, error)
	GetListByUserID(userID string, limit, offset int) ([]*model.Note, error)
	GetAllNotes() ([]*model.Note, error)
}
