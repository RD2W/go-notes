package note

import (
	"errors"
	"testing"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNoteRepository - мок-репозиторий для тестирования
type MockNoteRepository struct {
	mock.Mock
}

func (m *MockNoteRepository) Create(note *model.Note) error {
	args := m.Called(note)
	return args.Error(0)
}

func (m *MockNoteRepository) GetByID(id string) (*model.Note, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Note), args.Error(1)
}

func (m *MockNoteRepository) Update(note *model.Note) error {
	args := m.Called(note)
	return args.Error(0)
}

func (m *MockNoteRepository) DeleteByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockNoteRepository) GetAllNotes() ([]*model.Note, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Note), args.Error(1)
}

func (m *MockNoteRepository) GetAllNotesByUserID(userID string) ([]*model.Note, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Note), args.Error(1)
}

func (m *MockNoteRepository) GetListByUserID(userID string, limit, offset int) ([]*model.Note, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Note), args.Error(1)
}

func TestCreateNote(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("*model.Note")).Return(nil)

		result, err := service.CreateNote("Test Title", "Test Content", "user123")

		assert.NoError(t, err)
		assert.Equal(t, "Test Title", result.GetTitle())
		assert.Equal(t, "Test Content", result.GetContent())
		assert.Equal(t, "user123", result.GetUserID())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty Title", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		result, err := service.CreateNote("", "Test Content", "user123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "заголовок не может быть пустым", err.Error())
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("*model.Note")).Return(errors.New("database error"))

		result, err := service.CreateNote("Test Title", "Test Content", "user123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "database error", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestGetNoteByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		expectedNote := model.NewNote("Test Title", "Test Content", "user123")
		expectedNote.SetID("note123")

		mockRepo.On("GetByID", "note123").Return(expectedNote, nil)

		result, err := service.GetNoteByID("note123")

		assert.NoError(t, err)
		assert.Equal(t, expectedNote, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Note Not Found", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		mockRepo.On("GetByID", "nonexistent").Return((*model.Note)(nil), errors.New("not found"))

		result, err := service.GetNoteByID("nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "заметка не найдена", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateNote(t *testing.T) {
	t.Run("Success - Update Title and Content", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		existingNote := model.NewNote("Old Title", "Old Content", "user123")
		existingNote.SetID("note123")

		mockRepo.On("GetByID", "note123").Return(existingNote, nil)
		mockRepo.On("Update", mock.AnythingOfType("*model.Note")).Return(nil)

		result, err := service.UpdateNote("note123", "New Title", "New Content")

		assert.NoError(t, err)
		assert.Equal(t, "New Title", result.GetTitle())
		assert.Equal(t, "New Content", result.GetContent())
		assert.Equal(t, "user123", result.GetUserID())
		assert.Equal(t, "note123", result.GetID())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Update Only Title", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		existingNote := model.NewNote("Old Title", "Old Content", "user123")
		existingNote.SetID("note123")

		mockRepo.On("GetByID", "note123").Return(existingNote, nil)
		mockRepo.On("Update", mock.AnythingOfType("*model.Note")).Return(nil)

		result, err := service.UpdateNote("note123", "New Title", "")

		assert.NoError(t, err)
		assert.Equal(t, "New Title", result.GetTitle())
		assert.Equal(t, "Old Content", result.GetContent())
		assert.Equal(t, "user123", result.GetUserID())
		assert.Equal(t, "note123", result.GetID())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Update Only Content", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		existingNote := model.NewNote("Old Title", "Old Content", "user123")
		existingNote.SetID("note123")

		mockRepo.On("GetByID", "note123").Return(existingNote, nil)
		mockRepo.On("Update", mock.AnythingOfType("*model.Note")).Return(nil)

		result, err := service.UpdateNote("note123", "", "New Content")

		assert.NoError(t, err)
		assert.Equal(t, "Old Title", result.GetTitle())
		assert.Equal(t, "New Content", result.GetContent())
		assert.Equal(t, "user123", result.GetUserID())
		assert.Equal(t, "note123", result.GetID())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Note Not Found", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		mockRepo.On("GetByID", "nonexistent").Return((*model.Note)(nil), errors.New("not found"))

		result, err := service.UpdateNote("nonexistent", "New Title", "New Content")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "заметка не найдена", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Update Error", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		existingNote := model.NewNote("Old Title", "Old Content", "user123")
		existingNote.SetID("note123")

		mockRepo.On("GetByID", "note123").Return(existingNote, nil)
		mockRepo.On("Update", mock.AnythingOfType("*model.Note")).Return(errors.New("database error"))

		result, err := service.UpdateNote("note123", "New Title", "New Content")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "database error", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteNote(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		mockRepo.On("DeleteByID", "note123").Return(nil)

		err := service.DeleteNote("note123")

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Note Not Found", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		mockRepo.On("DeleteByID", "nonexistent").Return(errors.New("not found"))

		err := service.DeleteNote("nonexistent")

		assert.Error(t, err)
		assert.Equal(t, "заметка не найдена", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllNotes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		notes := []*model.Note{
			model.NewNote("Title 1", "Content 1", "user1"),
			model.NewNote("Title 2", "Content 2", "user2"),
		}

		mockRepo.On("GetAllNotes").Return(notes, nil)

		result, err := service.GetAllNotes()

		assert.NoError(t, err)
		assert.Equal(t, notes, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		mockRepo.On("GetAllNotes").Return(([]*model.Note)(nil), errors.New("database error"))

		result, err := service.GetAllNotes()

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "database error", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllNotesByUserID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		notes := []*model.Note{
			model.NewNote("Title 1", "Content 1", "user123"),
			model.NewNote("Title 2", "Content 2", "user123"),
		}

		mockRepo.On("GetAllNotesByUserID", "user123").Return(notes, nil)

		result, err := service.GetAllNotesByUserID("user123")

		assert.NoError(t, err)
		assert.Equal(t, notes, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		mockRepo.On("GetAllNotesByUserID", "user123").Return(([]*model.Note)(nil), errors.New("database error"))

		result, err := service.GetAllNotesByUserID("user123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "database error", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestGetListByUserID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		notes := []*model.Note{
			model.NewNote("Title 1", "Content 1", "user123"),
			model.NewNote("Title 2", "Content 2", "user123"),
		}

		mockRepo.On("GetListByUserID", "user123", 10, 0).Return(notes, nil)

		result, err := service.GetListByUserID("user123", 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, notes, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(MockNoteRepository)
		service := NewNoteService(mockRepo)

		mockRepo.On("GetListByUserID", "user123", 10, 0).Return(([]*model.Note)(nil), errors.New("database error"))

		result, err := service.GetListByUserID("user123", 10, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "database error", err.Error())

		mockRepo.AssertExpectations(t)
	})
}
