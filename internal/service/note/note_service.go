package note

import (
	"errors"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/domain/repository"
	"github.com/rd2w/go-notes/internal/domain/service"
)

// noteService реализация бизнес-логики для заметок
type noteService struct {
	repo repository.Repository
}

// NewNoteService создает новый экземпляр сервиса заметок
func NewNoteService(repo repository.Repository) service.NoteService {
	return &noteService{
		repo: repo,
	}
}

// CreateNote создает новую заметку
func (s *noteService) CreateNote(title, content string) (*model.Note, error) {
	if title == "" {
		return nil, errors.New("заголовок не может быть пустым")
	}

	note := model.NewNote(title, content)
	s.repo.Save(note)
	return note, nil
}

// GetNoteByID возвращает заметку по ID
func (s *noteService) GetNoteByID(id string) (*model.Note, error) {
	entity := s.repo.GetByID("note", id)
	if entity == nil {
		return nil, errors.New("заметка не найдена")
	}

	note, ok := entity.(*model.Note)
	if !ok {
		return nil, errors.New("ошибка преобразования сущности")
	}

	return note, nil
}

// UpdateNote обновляет заметку
func (s *noteService) UpdateNote(id, title, content string) (*model.Note, error) {
	entity := s.repo.GetByID("note", id)
	if entity == nil {
		return nil, errors.New("заметка не найдена")
	}

	note, ok := entity.(*model.Note)
	if !ok {
		return nil, errors.New("ошибка преобразования сущности")
	}

	if title != "" {
		note.SetTitle(title)
	}
	if content != "" {
		note.SetContent(content)
	}

	s.repo.Save(note)
	return note, nil
}

// DeleteNote удаляет заметку
func (s *noteService) DeleteNote(id string) error {
	deleted := s.repo.DeleteByID("note", id)
	if !deleted {
		return errors.New("заметка не найдена")
	}
	return nil
}

// GetAllNotes возвращает все заметки
func (s *noteService) GetAllNotes() ([]*model.Note, error) {
	notes := s.repo.GetAllNotes()
	return notes, nil
}
