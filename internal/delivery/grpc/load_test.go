package grpc

import (
	"context"
	"sync"
	"testing"

	"github.com/rd2w/go-notes/internal/domain/model"
	notePb "github.com/rd2w/go-notes/pkg/proto/note"
	"github.com/stretchr/testify/assert"
)

// TestConcurrentGRPCNoteCreation tests concurrent gRPC note creation requests
func TestConcurrentGRPCNoteCreation(t *testing.T) {
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

	mockService.On("CreateNote", "Test Title", "Test Content", "user123").Return(expectedNote, nil).Times(10)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := server.CreateNote(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "note123", resp.Note.Id)
		}()
	}

	wg.Wait()
	mockService.AssertExpectations(t)
}

// TestConcurrentGRPCNoteRetrieval tests concurrent gRPC note retrieval requests
func TestConcurrentGRPCNoteRetrieval(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.GetRequest{
		Id: "note123",
	}

	// Создаем тестовую заметку с помощью конструктора
	expectedNote := model.NewNote("Test Title", "Test Content", "user123")
	expectedNote.SetID("note123")

	mockService.On("GetNoteByID", "note123").Return(expectedNote, nil).Times(10)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := server.GetNote(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "note123", resp.Note.Id)
		}()
	}

	wg.Wait()
	mockService.AssertExpectations(t)
}

// BenchmarkConcurrentGRPCNoteCreation benchmarks concurrent gRPC note creation
func BenchmarkConcurrentGRPCNoteCreation(b *testing.B) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.CreateNoteRequest{
		Title:   "Benchmark Title",
		Content: "Benchmark Content",
		UserId:  "benchmark_user",
	}

	expectedNote := model.NewNote("Benchmark Title", "Benchmark Content", "benchmark_user")
	expectedNote.SetID("benchmark_note")

	mockService.On("CreateNote", "Benchmark Title", "Benchmark Content", "benchmark_user").Return(expectedNote, nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := server.CreateNote(ctx, req)
				if err != nil {
					b.Logf("Error creating note via gRPC: %v", err)
				}
			}()
		}

		wg.Wait()
	}
	mockService.AssertExpectations(b)
}

// BenchmarkConcurrentGRPCNoteRetrieval benchmarks concurrent gRPC note retrieval
func BenchmarkConcurrentGRPCNoteRetrieval(b *testing.B) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()
	req := &notePb.GetRequest{
		Id: "benchmark_note",
	}

	expectedNote := model.NewNote("Benchmark Title", "Benchmark Content", "benchmark_user")
	expectedNote.SetID("benchmark_note")

	mockService.On("GetNoteByID", "benchmark_note").Return(expectedNote, nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := server.GetNote(ctx, req)
				if err != nil {
					b.Logf("Error retrieving note via gRPC: %v", err)
				}
			}()
		}

		wg.Wait()
	}
	mockService.AssertExpectations(b)
}

// TestMixedConcurrentGRPCRequests tests mixed concurrent gRPC requests
func TestMixedConcurrentGRPCRequests(t *testing.T) {
	mockService := new(MockNoteService)
	server := NewNoteServiceServer(mockService)

	ctx := context.Background()

	// Подготовим тестовые объекты
	createdNote := model.NewNote("Created Title", "Created Content", "user123")
	createdNote.SetID("created_note")

	retrievedNote := model.NewNote("Retrieved Title", "Retrieved Content", "user456")
	retrievedNote.SetID("retrieved_note")

	// Настроим ожидания
	mockService.On("CreateNote", "Created Title", "Created Content", "user123").Return(createdNote, nil).Times(5)
	mockService.On("GetNoteByID", "retrieved_note").Return(retrievedNote, nil).Times(5)
	mockService.On("GetAllNotes").Return([]*model.Note{retrievedNote}, nil).Times(5)

	var wg sync.WaitGroup

	// Concurrent note creation requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := &notePb.CreateNoteRequest{
				Title:   "Created Title",
				Content: "Created Content",
				UserId:  "user123",
			}
			resp, err := server.CreateNote(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "created_note", resp.Note.Id)
		}()
	}

	// Concurrent note retrieval requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := &notePb.GetRequest{
				Id: "retrieved_note",
			}
			resp, err := server.GetNote(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "retrieved_note", resp.Note.Id)
		}()
	}

	// Concurrent get all notes requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := &notePb.Empty{}
			resp, err := server.ListNotes(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, resp.Notes, 1)
		}()
	}

	wg.Wait()
	mockService.AssertExpectations(t)
}
