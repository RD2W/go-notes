package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/rd2w/go-notes/internal/domain/model"
	notePb "github.com/rd2w/go-notes/pkg/proto/note"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNoteService - мок-сервис для тестирования
type MockNoteService struct {
	mock.Mock
}

func (m *MockNoteService) CreateNote(title, content, userId string) (*model.Note, error) {
	args := m.Called(title, content, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Note), args.Error(1)
}

func (m *MockNoteService) GetNoteByID(id string) (*model.Note, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Note), args.Error(1)
}

func (m *MockNoteService) UpdateNote(id, title, content string) (*model.Note, error) {
	args := m.Called(id, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Note), args.Error(1)
}

func (m *MockNoteService) DeleteNote(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockNoteService) GetAllNotes() ([]*model.Note, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Note), args.Error(1)
}

func (m *MockNoteService) GetAllNotesByUserID(userID string) ([]*model.Note, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Note), args.Error(1)
}

func (m *MockNoteService) GetListByUserID(userID string, limit, offset int) ([]*model.Note, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Note), args.Error(1)
}

func TestCreateNote(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.CreateNoteRequest{
		Title:   "Test Title",
		Content: "Test Content",
		UserId:  "user123",
	}

	// Создаем тестовую заметку с помощью конструктора
	expectedNote := model.NewNote("Test Title", "Test Content", "user123")
	expectedNote.SetID("note123")

	mockService.On("CreateNote", "Test Title", "Test Content", "user123").Return(expectedNote, nil)

	resp, err := server.CreateNote(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "note123", resp.Note.Id)
	assert.Equal(t, "Test Title", resp.Note.Title)
	assert.Equal(t, "Test Content", resp.Note.Content)
	assert.Equal(t, "user123", resp.Note.UserId)

	mockService.AssertExpectations(t)
}

func TestCreateNote_Error(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.CreateNoteRequest{
		Title:   "Test Title",
		Content: "Test Content",
		UserId:  "user123",
	}

	mockService.On("CreateNote", "Test Title", "Test Content", "user123").Return(nil, errors.New("creation failed"))

	resp, err := server.CreateNote(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "creation failed", err.Error())

	mockService.AssertExpectations(t)
}

func TestGetNote(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.GetRequest{
		Id: "note123",
	}

	// Создаем тестовую заметку с помощью конструктора
	expectedNote := model.NewNote("Test Title", "Test Content", "user123")
	expectedNote.SetID("note123")

	mockService.On("GetNoteByID", "note123").Return(expectedNote, nil)

	resp, err := server.GetNote(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "note123", resp.Note.Id)
	assert.Equal(t, "Test Title", resp.Note.Title)
	assert.Equal(t, "Test Content", resp.Note.Content)
	assert.Equal(t, "user123", resp.Note.UserId)

	mockService.AssertExpectations(t)
}

func TestGetNote_Error(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.GetRequest{
		Id: "note123",
	}

	mockService.On("GetNoteByID", "note123").Return(nil, errors.New("not found"))

	resp, err := server.GetNote(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "not found", err.Error())

	mockService.AssertExpectations(t)
}

func TestUpdateNote(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.UpdateNoteRequest{
		Id:      "note123",
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	// Создаем тестовую заметку с помощью конструктора
	expectedNote := model.NewNote("Updated Title", "Updated Content", "user123")
	expectedNote.SetID("note123")

	mockService.On("UpdateNote", "note123", "Updated Title", "Updated Content").Return(expectedNote, nil)

	resp, err := server.UpdateNote(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "note123", resp.Note.Id)
	assert.Equal(t, "Updated Title", resp.Note.Title)
	assert.Equal(t, "Updated Content", resp.Note.Content)
	assert.Equal(t, "user123", resp.Note.UserId)

	mockService.AssertExpectations(t)
}

func TestUpdateNote_Error(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.UpdateNoteRequest{
		Id:      "note123",
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	mockService.On("UpdateNote", "note123", "Updated Title", "Updated Content").Return(nil, errors.New("update failed"))

	resp, err := server.UpdateNote(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "update failed", err.Error())

	mockService.AssertExpectations(t)
}

func TestDeleteNote(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.GetRequest{
		Id: "note123",
	}

	mockService.On("DeleteNote", "note123").Return(nil)

	resp, err := server.DeleteNote(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "note deleted successfully", resp.Message)

	mockService.AssertExpectations(t)
}

func TestDeleteNote_Error(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.GetRequest{
		Id: "note123",
	}

	mockService.On("DeleteNote", "note123").Return(errors.New("delete failed"))

	resp, err := server.DeleteNote(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "delete failed", err.Error())

	mockService.AssertExpectations(t)
}

func TestListNotes(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.Empty{}

	// Создаем тестовые заметки с помощью конструктора
	note1 := model.NewNote("Test Title 1", "Test Content 1", "user123")
	note1.SetID("note123")
	note2 := model.NewNote("Test Title 2", "Test Content 2", "user456")
	note2.SetID("note456")

	expectedNotes := []*model.Note{note1, note2}

	mockService.On("GetAllNotes").Return(expectedNotes, nil)

	resp, err := server.ListNotes(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Notes, 2)
	assert.Equal(t, "note123", resp.Notes[0].Id)
	assert.Equal(t, "Test Title 1", resp.Notes[0].Title)
	assert.Equal(t, "Test Content 1", resp.Notes[0].Content)
	assert.Equal(t, "user123", resp.Notes[0].UserId)
	assert.Equal(t, "note456", resp.Notes[1].Id)
	assert.Equal(t, "Test Title 2", resp.Notes[1].Title)
	assert.Equal(t, "Test Content 2", resp.Notes[1].Content)
	assert.Equal(t, "user456", resp.Notes[1].UserId)

	mockService.AssertExpectations(t)
}

func TestListNotes_Error(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.Empty{}

	mockService.On("GetAllNotes").Return(nil, errors.New("list failed"))

	resp, err := server.ListNotes(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "list failed", err.Error())

	mockService.AssertExpectations(t)
}
