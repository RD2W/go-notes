package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNoteService - mock-объект для NoteService
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

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

func TestNoteHandler_CreateNote(t *testing.T) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	router := setupRouter()
	router.POST("/notes", handler.CreateNote)

	// Тест 1: Успешное создание заметки
	t.Run("Successful note creation", func(t *testing.T) {
		note := model.NewNote("Test Title", "Test Content", "user123")
		note.SetID("1")

		mockService.On("CreateNote", "Test Title", "Test Content", "user123").Return(note, nil).Once()

		requestBody := `{"title":"Test Title","content":"Test Content"}`
		req, _ := http.NewRequest("POST", "/notes", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		// Добавляем user_id в контекст, как это делает middleware аутентификации
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", "user123")

		handler.CreateNote(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var responseNote model.Note
		err := json.Unmarshal(w.Body.Bytes(), &responseNote)
		assert.NoError(t, err)
		assert.Equal(t, "1", responseNote.GetID())
		assert.Equal(t, "Test Title", responseNote.GetTitle())
		assert.Equal(t, "Test Content", responseNote.GetContent())
		assert.Equal(t, "user123", responseNote.GetUserID())

		mockService.AssertExpectations(t)
	})

	// Тест 2: Ошибка валидации (отсутствует заголовок)
	t.Run("Validation error - missing title", func(t *testing.T) {
		requestBody := `{"content":"Test Content"}`
		req, _ := http.NewRequest("POST", "/notes", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", "user123")

		handler.CreateNote(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Тест 3: Пользователь не аутентифицирован
	t.Run("User not authenticated", func(t *testing.T) {
		requestBody := `{"title":"Test Title","content":"Test Content"}`
		req, _ := http.NewRequest("POST", "/notes", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// Не устанавливаем user_id в контексте

		handler.CreateNote(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Тест 4: Ошибка сервиса при создании заметки
	t.Run("Service error during note creation", func(t *testing.T) {
		mockService.On("CreateNote", "Test Title", "Test Content", "user123").Return((*model.Note)(nil), errors.New("service error")).Once()

		requestBody := `{"title":"Test Title","content":"Test Content"}`
		req, _ := http.NewRequest("POST", "/notes", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", "user123")

		handler.CreateNote(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestNoteHandler_GetNote(t *testing.T) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	router := setupRouter()
	router.GET("/notes/:id", handler.GetNote)

	// Тест 1: Успешное получение заметки
	t.Run("Successful note retrieval", func(t *testing.T) {
		note := model.NewNote("Test Title", "Test Content", "user123")
		note.SetID("1")

		mockService.On("GetNoteByID", "1").Return(note, nil).Once()

		req, _ := http.NewRequest("GET", "/notes/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseNote model.Note
		err := json.Unmarshal(w.Body.Bytes(), &responseNote)
		assert.NoError(t, err)
		assert.Equal(t, "1", responseNote.GetID())
		assert.Equal(t, "Test Title", responseNote.GetTitle())
		assert.Equal(t, "Test Content", responseNote.GetContent())

		mockService.AssertExpectations(t)
	})

	// Тест 2: Заметка не найдена
	t.Run("Note not found", func(t *testing.T) {
		mockService.On("GetNoteByID", "nonexistent").Return((*model.Note)(nil), errors.New("note not found")).Once()

		req, _ := http.NewRequest("GET", "/notes/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestNoteHandler_UpdateNote(t *testing.T) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	router := setupRouter()
	router.PUT("/notes/:id", handler.UpdateNote)

	// Тест 1: Успешное обновление заметки
	t.Run("Successful note update", func(t *testing.T) {
		updatedNote := model.NewNote("Updated Title", "Updated Content", "user123")
		updatedNote.SetID("1")

		mockService.On("UpdateNote", "1", "Updated Title", "Updated Content").Return(updatedNote, nil).Once()

		requestBody := `{"id":"1","title":"Updated Title","content":"Updated Content","user_id":"user123"}`
		req, _ := http.NewRequest("PUT", "/notes/1", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		handler.UpdateNote(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseNote model.Note
		err := json.Unmarshal(w.Body.Bytes(), &responseNote)
		assert.NoError(t, err)
		assert.Equal(t, "1", responseNote.GetID())
		assert.Equal(t, "Updated Title", responseNote.GetTitle())
		assert.Equal(t, "Updated Content", responseNote.GetContent())

		mockService.AssertExpectations(t)
	})

	// Тест 2: Заметка не найдена при обновлении
	t.Run("Note not found during update", func(t *testing.T) {
		mockService.On("UpdateNote", "nonexistent", "Updated Title", "Updated Content").Return((*model.Note)(nil), errors.New("note not found")).Once()

		requestBody := `{"id":"nonexistent","title":"Updated Title","content":"Updated Content","user_id":"user123"}`
		req, _ := http.NewRequest("PUT", "/notes/nonexistent", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}

		handler.UpdateNote(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	// Тест 3: Ошибка валидации при обновлении
	t.Run("Validation error during update", func(t *testing.T) {
		// В этом тесте мы проверяем случай, когда валидация проходит, но обновление не происходит
		// потому что поля пустые, и mock не ожидает вызова UpdateNote с пустыми строками
		mockService.On("UpdateNote", "1", "", "").Return((*model.Note)(nil), errors.New("validation error")).Once()

		requestBody := `{"title":"","content":""}`
		req, _ := http.NewRequest("PUT", "/notes/1", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		handler.UpdateNote(c)

		assert.Equal(t, http.StatusNotFound, w.Code) // В note_handler.go при ошибке UpdateNote возвращается 404
		mockService.AssertExpectations(t)
	})
}

func TestNoteHandler_DeleteNote(t *testing.T) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	router := setupRouter()
	router.DELETE("/notes/:id", handler.DeleteNote)

	// Тест 1: Успешное удаление заметки
	t.Run("Successful note deletion", func(t *testing.T) {
		mockService.On("DeleteNote", "1").Return(nil).Once()

		req, _ := http.NewRequest("DELETE", "/notes/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		mockService.AssertExpectations(t)
	})

	// Тест 2: Заметка не найдена при удалении
	t.Run("Note not found during deletion", func(t *testing.T) {
		mockService.On("DeleteNote", "nonexistent").Return(errors.New("note not found")).Once()

		req, _ := http.NewRequest("DELETE", "/notes/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestNoteHandler_GetAllNotes(t *testing.T) {
	mockService := new(MockNoteService)
	handler := NewNoteHandler(mockService)

	router := setupRouter()
	router.GET("/notes", handler.GetAllNotes)

	// Тест 1: Успешное получение всех заметок
	t.Run("Successful retrieval of all notes", func(t *testing.T) {
		note1 := model.NewNote("Test Title 1", "Test Content 1", "user123")
		note1.SetID("1")
		note2 := model.NewNote("Test Title 2", "Test Content 2", "user123")
		note2.SetID("2")

		notes := []*model.Note{note1, note2}

		mockService.On("GetAllNotes").Return(notes, nil).Once()

		req, _ := http.NewRequest("GET", "/notes", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseNotes []*model.Note
		err := json.Unmarshal(w.Body.Bytes(), &responseNotes)
		assert.NoError(t, err)
		assert.Len(t, responseNotes, 2)
		assert.Equal(t, "1", responseNotes[0].GetID())
		assert.Equal(t, "2", responseNotes[1].GetID())

		mockService.AssertExpectations(t)
	})

	// Тест 2: Ошибка при получении всех заметок
	t.Run("Error during retrieval of all notes", func(t *testing.T) {
		mockService.On("GetAllNotes").Return(([]*model.Note)(nil), errors.New("database error")).Once()

		req, _ := http.NewRequest("GET", "/notes", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}
