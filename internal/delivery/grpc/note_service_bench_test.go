package grpc

import (
	"context"
	"testing"

	"github.com/rd2w/go-notes/internal/domain/model"
	notePb "github.com/rd2w/go-notes/pkg/proto/note"
	"github.com/stretchr/testify/assert"
)

// BenchmarkNoteServiceServerCreateNote benchmarks the gRPC CreateNote method
func BenchmarkNoteServiceServerCreateNote(b *testing.B) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.CreateNoteRequest{
		Title:   "Test Title",
		Content: "Test Content",
		UserId:  "user123",
	}

	expectedNote := model.NewNote("Test Title", "Test Content", "user123")
	expectedNote.SetID("note123")

	mockService.On("CreateNote", "Test Title", "Test Content", "user123").Return(expectedNote, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.CreateNote(ctx, req)
		assert.NoError(b, err)
	}
	mockService.AssertExpectations(b)
}

// BenchmarkNoteServiceServerGetNote benchmarks the gRPC GetNote method
func BenchmarkNoteServiceServerGetNote(b *testing.B) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.GetRequest{
		Id: "note123",
	}

	expectedNote := model.NewNote("Test Title", "Test Content", "user123")
	expectedNote.SetID("note123")

	mockService.On("GetNoteByID", "note123").Return(expectedNote, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.GetNote(ctx, req)
		assert.NoError(b, err)
	}
	mockService.AssertExpectations(b)
}

// BenchmarkNoteServiceServerUpdateNote benchmarks the gRPC UpdateNote method
func BenchmarkNoteServiceServerUpdateNote(b *testing.B) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.UpdateNoteRequest{
		Id:      "note123",
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	expectedNote := model.NewNote("Updated Title", "Updated Content", "user123")
	expectedNote.SetID("note123")

	mockService.On("UpdateNote", "note123", "Updated Title", "Updated Content").Return(expectedNote, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.UpdateNote(ctx, req)
		assert.NoError(b, err)
	}
	mockService.AssertExpectations(b)
}

// BenchmarkNoteServiceServerDeleteNote benchmarks the gRPC DeleteNote method
func BenchmarkNoteServiceServerDeleteNote(b *testing.B) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.GetRequest{
		Id: "note123",
	}

	mockService.On("DeleteNote", "note123").Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.DeleteNote(ctx, req)
		assert.NoError(b, err)
	}
	mockService.AssertExpectations(b)
}

// BenchmarkNoteServiceServerListNotes benchmarks the gRPC ListNotes method
func BenchmarkNoteServiceServerListNotes(b *testing.B) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.Empty{}

	note1 := model.NewNote("Test Title 1", "Test Content 1", "user123")
	note1.SetID("note123")
	note2 := model.NewNote("Test Title 2", "Test Content 2", "user456")
	note2.SetID("note456")

	expectedNotes := []*model.Note{note1, note2}

	mockService.On("GetAllNotes").Return(expectedNotes, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.ListNotes(ctx, req)
		assert.NoError(b, err)
	}
	mockService.AssertExpectations(b)
}
