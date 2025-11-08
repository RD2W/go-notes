package repository

import (
	"github.com/rd2w/go-notes/internal/model"
)

// Entity интерфейс, который должны реализовывать все сущности
type Entity interface {
	GetID() string
	GetType() string
}

// Repository интерфейс для репозитория
type Repository interface {
	Save(entity Entity)
	GetAllNotes() []*model.Note
	GetNotesCount() int
	GetNewNotes(lastIndex int) []*model.Note
}
