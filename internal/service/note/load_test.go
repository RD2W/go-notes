package note

import (
	"sync"
	"testing"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestConcurrentNoteCreation tests concurrent note creation
func TestConcurrentNoteCreation(t *testing.T) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	// Ожидаем, что Create будет вызван 10 раз
	mockRepo.On("Create", mock.AnythingOfType("*model.Note")).Return(nil).Times(10)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := service.CreateNote(
				"Title "+string(rune(id)),
				"Content "+string(rune(id)),
				"user"+string(rune(id)),
			)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
	mockRepo.AssertExpectations(t)
}

// TestConcurrentNoteRetrieval tests concurrent note retrieval
func TestConcurrentNoteRetrieval(t *testing.T) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	// Создаем тестовую заметку
	expectedNote := model.NewNote("Test Title", "Test Content", "user123")
	expectedNote.SetID("note123")

	// Ожидаем, что GetByID будет вызван 10 раз
	mockRepo.On("GetByID", "note123").Return(expectedNote, nil).Times(10)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.GetNoteByID("note123")
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	mockRepo.AssertExpectations(t)
}

// TestMixedConcurrentOperations tests mixed concurrent operations
func TestMixedConcurrentOperations(t *testing.T) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	// Подготовим ожидания для разных операций
	expectedNote := model.NewNote("Test Title", "Test Content", "user123")
	expectedNote.SetID("note123")

	mockRepo.On("Create", mock.AnythingOfType("*model.Note")).Return(nil).Times(5)
	mockRepo.On("GetByID", "note123").Return(expectedNote, nil).Times(5)
	mockRepo.On("GetAllNotes").Return([]*model.Note{expectedNote}, nil).Times(5)

	var wg sync.WaitGroup

	// Создание заметок
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := service.CreateNote(
				"Title "+string(rune(id)),
				"Content "+string(rune(id)),
				"user"+string(rune(id)),
			)
			assert.NoError(t, err)
		}(i)
	}

	// Получение заметок
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.GetNoteByID("note123")
			assert.NoError(t, err)
		}()
	}

	// Получение всех заметок
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.GetAllNotes()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	mockRepo.AssertExpectations(t)
}

// BenchmarkConcurrentNoteCreation benchmarks concurrent note creation
func BenchmarkConcurrentNoteCreation(b *testing.B) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	mockRepo.On("Create", mock.AnythingOfType("*model.Note")).Return(nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, err := service.CreateNote(
					"Benchmark Title "+string(rune(id+i*10)),
					"Benchmark Content "+string(rune(id+i*10)),
					"benchmark_user"+string(rune(id)),
				)
				if err != nil {
					b.Logf("Error creating note: %v", err)
				}
			}(j)
		}

		wg.Wait()
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkConcurrentNoteRetrieval benchmarks concurrent note retrieval
func BenchmarkConcurrentNoteRetrieval(b *testing.B) {
	mockRepo := new(MockNoteRepository)
	service := NewNoteService(mockRepo)

	expectedNote := model.NewNote("Benchmark Title", "Benchmark Content", "benchmark_user")
	expectedNote.SetID("benchmark_note_123")

	mockRepo.On("GetByID", "benchmark_note_123").Return(expectedNote, nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := service.GetNoteByID("benchmark_note_123")
				if err != nil {
					b.Logf("Error retrieving note: %v", err)
				}
			}()
		}

		wg.Wait()
	}
	mockRepo.AssertExpectations(b)
}
