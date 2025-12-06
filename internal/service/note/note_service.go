package note

import (
	"errors"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/domain/repository"
	"github.com/rd2w/go-notes/internal/domain/service"
)

// noteService реализация бизнес-логики для заметок
type noteService struct {
	noteRepo repository.NoteRepository
}

// NewNoteService создает новый экземпляр сервиса заметок
func NewNoteService(noteRepo repository.NoteRepository) service.NoteService {
	return &noteService{
		noteRepo: noteRepo,
	}
}

// CreateNote создает новую заметку
func (s *noteService) CreateNote(title, content, userId string) (*model.Note, error) {
	if title == "" {
		return nil, errors.New("заголовок не может быть пустым")
	}

	note := model.NewNote(title, content, userId)
	err := s.noteRepo.Create(note)
	if err != nil {
		return nil, err
	}
	return note, nil
}

// GetNoteByID возвращает заметку по ID
func (s *noteService) GetNoteByID(id string) (*model.Note, error) {
	note, err := s.noteRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("заметка не найдена")
	}

	return note, nil
}

// UpdateNote обновляет заметку
func (s *noteService) UpdateNote(id, title, content string) (*model.Note, error) {
	note, err := s.noteRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("заметка не найдена")
	}

	if title != "" {
		note.SetTitle(title)
	}
	if content != "" {
		note.SetContent(content)
	}

	err = s.noteRepo.Update(note)
	if err != nil {
		return nil, err
	}
	return note, nil
}

// DeleteNote удаляет заметку
func (s *noteService) DeleteNote(id string) error {
	err := s.noteRepo.DeleteByID(id)
	if err != nil {
		return errors.New("заметка не найдена")
	}
	return nil
}

// GetAllNotes возвращает все заметки
func (s *noteService) GetAllNotes() ([]*model.Note, error) {
	notes, err := s.noteRepo.GetAllNotes()
	if err != nil {
		return nil, err
	}
	return notes, nil
}

// GetAllNotesByUserID возвращает все заметки пользователя
func (s *noteService) GetAllNotesByUserID(userID string) ([]*model.Note, error) {
	notes, err := s.noteRepo.GetAllNotesByUserID(userID)
	if err != nil {
		return nil, err
	}
	return notes, nil
}

// GetListByUserID возвращает список заметок пользователя с пагинацией
func (s *noteService) GetListByUserID(userID string, limit, offset int) ([]*model.Note, error) {
	notes, err := s.noteRepo.GetListByUserID(userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return notes, nil
}
