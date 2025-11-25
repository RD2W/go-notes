package note

import (
	"testing"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// BenchmarkNoteServiceCreateNote benchmarks the CreateNote method
func BenchmarkNoteServiceCreateNote(b *testing.B) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	mockRepo.On("Create", mock.AnythingOfType("*model.Note")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateNote("Test Title", "Test Content", "user123")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkNoteServiceGetNoteByID benchmarks the GetNoteByID method
func BenchmarkNoteServiceGetNoteByID(b *testing.B) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	expectedNote := model.NewNote("Test Title", "Test Content", "user123")
	expectedNote.SetID("note123")

	mockRepo.On("GetByID", "note123").Return(expectedNote, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetNoteByID("note123")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkNoteServiceUpdateNote benchmarks the UpdateNote method
func BenchmarkNoteServiceUpdateNote(b *testing.B) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	existingNote := model.NewNote("Old Title", "Old Content", "user123")
	existingNote.SetID("note123")

	mockRepo.On("GetByID", "note123").Return(existingNote, nil)
	mockRepo.On("Update", mock.AnythingOfType("*model.Note")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.UpdateNote("note123", "New Title", "New Content")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkNoteServiceDeleteNote benchmarks the DeleteNote method
func BenchmarkNoteServiceDeleteNote(b *testing.B) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	mockRepo.On("DeleteByID", "note123").Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.DeleteNote("note123")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkNoteServiceGetAllNotes benchmarks the GetAllNotes method
func BenchmarkNoteServiceGetAllNotes(b *testing.B) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	notes := []*model.Note{
		model.NewNote("Title 1", "Content 1", "user1"),
		model.NewNote("Title 2", "Content 2", "user2"),
	}

	mockRepo.On("GetAllNotes").Return(notes, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetAllNotes()
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkNoteServiceGetAllNotesByUserID benchmarks the GetAllNotesByUserID method
func BenchmarkNoteServiceGetAllNotesByUserID(b *testing.B) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	notes := []*model.Note{
		model.NewNote("Title 1", "Content 1", "user123"),
		model.NewNote("Title 2", "Content 2", "user123"),
	}

	mockRepo.On("GetAllNotesByUserID", "user123").Return(notes, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetAllNotesByUserID("user123")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkNoteServiceGetListByUserID benchmarks the GetListByUserID method
func BenchmarkNoteServiceGetListByUserID(b *testing.B) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	notes := []*model.Note{
		model.NewNote("Title 1", "Content 1", "user123"),
		model.NewNote("Title 2", "Content 2", "user123"),
	}

	mockRepo.On("GetListByUserID", "user123", 10, 0).Return(notes, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetListByUserID("user123", 10, 0)
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}
