package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

// BenchmarkNoteHandlerCreateNote benchmarks the HTTP CreateNote handler
func BenchmarkNoteHandlerCreateNote(b *testing.B) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	note := model.NewNote("Test Title", "Test Content", "user123")
	note.SetID("1")

	mockService.On("CreateNote", "Test Title", "Test Content", "user123").Return(note, nil).Maybe()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/notes", handler.CreateNote)

	requestBody := `{"title":"Test Title","content":"Test Content"}`
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", "user123")

		handler.CreateNote(c)
		assert.Equal(b, http.StatusCreated, w.Code)
	}
	mockService.AssertExpectations(b)
}

// BenchmarkNoteHandlerGetNote benchmarks the HTTP GetNote handler
func BenchmarkNoteHandlerGetNote(b *testing.B) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	note := model.NewNote("Test Title", "Test Content", "user123")
	note.SetID("1")

	mockService.On("GetNoteByID", "1").Return(note, nil).Maybe()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/notes/:id", handler.GetNote)

	req, _ := http.NewRequest("GET", "/notes/1", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(b, http.StatusOK, w.Code)
	}
	mockService.AssertExpectations(b)
}

// BenchmarkNoteHandlerUpdateNote benchmarks the HTTP UpdateNote handler
func BenchmarkNoteHandlerUpdateNote(b *testing.B) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	updatedNote := model.NewNote("Updated Title", "Updated Content", "user123")
	updatedNote.SetID("1")

	mockService.On("UpdateNote", "1", "Updated Title", "Updated Content").Return(updatedNote, nil).Maybe()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/notes/:id", handler.UpdateNote)

	requestBody := `{"id":"1","title":"Updated Title","content":"Updated Content","user_id":"user123"}`
	req, _ := http.NewRequest("PUT", "/notes/1", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		handler.UpdateNote(c)
		assert.Equal(b, http.StatusOK, w.Code)
	}
	mockService.AssertExpectations(b)
}

// BenchmarkNoteHandlerDeleteNote benchmarks the HTTP DeleteNote handler
func BenchmarkNoteHandlerDeleteNote(b *testing.B) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	mockService.On("DeleteNote", "1").Return(nil).Maybe()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/notes/:id", handler.DeleteNote)

	req, _ := http.NewRequest("DELETE", "/notes/1", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(b, http.StatusNoContent, w.Code)
	}
	mockService.AssertExpectations(b)
}

// BenchmarkNoteHandlerGetAllNotes benchmarks the HTTP GetAllNotes handler
func BenchmarkNoteHandlerGetAllNotes(b *testing.B) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	note1 := model.NewNote("Test Title 1", "Test Content 1", "user123")
	note1.SetID("1")
	note2 := model.NewNote("Test Title 2", "Test Content 2", "user123")
	note2.SetID("2")

	notes := []*model.Note{note1, note2}

	mockService.On("GetAllNotes").Return(notes, nil).Maybe()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/notes", handler.GetAllNotes)

	req, _ := http.NewRequest("GET", "/notes", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(b, http.StatusOK, w.Code)
	}
	mockService.AssertExpectations(b)
}
