package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

// TestConcurrentHTTPNoteCreation tests concurrent HTTP note creation requests
func TestConcurrentHTTPNoteCreation(t *testing.T) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	note := model.NewNote("Test Title", "Test Content", "user123")
	note.SetID("1")

	mockService.On("CreateNote", "Test Title", "Test Content", "user123").Return(note, nil).Times(10)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/notes", handler.CreateNote)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			requestBody := `{"title":"Test Title","content":"Test Content"}`
			req, _ := http.NewRequest("POST", "/notes", bytes.NewBufferString(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", "user123")

			handler.CreateNote(c)
			assert.Equal(t, http.StatusCreated, w.Code)
		}(i)
	}

	wg.Wait()
	mockService.AssertExpectations(t)
}

// TestConcurrentHTTPNoteRetrieval tests concurrent HTTP note retrieval requests
func TestConcurrentHTTPNoteRetrieval(t *testing.T) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	note := model.NewNote("Test Title", "Test Content", "user123")
	note.SetID("1")

	mockService.On("GetNoteByID", "1").Return(note, nil).Times(10)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/notes/:id", handler.GetNote)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "/notes/1", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}()
	}

	wg.Wait()
	mockService.AssertExpectations(t)
}

// BenchmarkConcurrentHTTPNoteCreation benchmarks concurrent HTTP note creation
func BenchmarkConcurrentHTTPNoteCreation(b *testing.B) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	note := model.NewNote("Benchmark Title", "Benchmark Content", "benchmark_user")
	note.SetID("benchmark_note")

	mockService.On("CreateNote", "Benchmark Title", "Benchmark Content", "benchmark_user").Return(note, nil).Maybe()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/notes", handler.CreateNote)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				requestBody := `{"title":"Benchmark Title","content":"Benchmark Content"}`
				req, _ := http.NewRequest("POST", "/notes", bytes.NewBufferString(requestBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = req
				c.Set("user_id", "benchmark_user")

				handler.CreateNote(c)
			}()
		}

		wg.Wait()
	}
	mockService.AssertExpectations(b)
}

// BenchmarkConcurrentHTTPNoteRetrieval benchmarks concurrent HTTP note retrieval
func BenchmarkConcurrentHTTPNoteRetrieval(b *testing.B) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	note := model.NewNote("Benchmark Title", "Benchmark Content", "benchmark_user")
	note.SetID("benchmark_note")

	mockService.On("GetNoteByID", "benchmark_note").Return(note, nil).Maybe()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/notes/:id", handler.GetNote)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest("GET", "/notes/benchmark_note", nil)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
			}()
		}

		wg.Wait()
	}
	mockService.AssertExpectations(b)
}

// TestMixedConcurrentHTTPRequests tests mixed concurrent HTTP requests
func TestMixedConcurrentHTTPRequests(t *testing.T) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	createdNote := model.NewNote("Created Title", "Created Content", "user123")
	createdNote.SetID("created_note")

	retrievedNote := model.NewNote("Retrieved Title", "Retrieved Content", "user456")
	retrievedNote.SetID("retrieved_note")

	mockService.On("CreateNote", "Created Title", "Created Content", "user123").Return(createdNote, nil).Times(5)
	mockService.On("GetNoteByID", "retrieved_note").Return(retrievedNote, nil).Times(5)
	mockService.On("GetAllNotes").Return([]*model.Note{retrievedNote}, nil).Times(5)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/notes", handler.CreateNote)
	router.GET("/notes/:id", handler.GetNote)
	router.GET("/notes", handler.GetAllNotes)

	var wg sync.WaitGroup

	// Concurrent note creation requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			requestBody := `{"title":"Created Title","content":"Created Content"}`
			req, _ := http.NewRequest("POST", "/notes", bytes.NewBufferString(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", "user123")

			handler.CreateNote(c)
			assert.Equal(t, http.StatusCreated, w.Code)
		}()
	}

	// Concurrent note retrieval requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "/notes/retrieved_note", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}()
	}

	// Concurrent get all notes requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "/notes", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}()
	}

	wg.Wait()
	mockService.AssertExpectations(t)
}
