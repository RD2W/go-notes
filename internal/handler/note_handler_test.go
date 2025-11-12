package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
	"github.com/stretchr/testify/assert"
)

// MockRepository - мок-репозиторий для тестирования
type MockRepository struct {
	notes map[string]repository.Entity
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		notes: make(map[string]repository.Entity),
	}
}

func (m *MockRepository) Save(entity repository.Entity) {
	m.notes[entity.GetID()] = entity
}

func (m *MockRepository) GetByID(entityType, id string) repository.Entity {
	entity, exists := m.notes[id]
	if !exists {
		return nil
	}
	return entity
}

func (m *MockRepository) DeleteByID(entityType, id string) bool {
	_, exists := m.notes[id]
	if !exists {
		return false
	}
	delete(m.notes, id)
	return true
}

func (m *MockRepository) GetAllNotes() []*model.Note {
	notes := make([]*model.Note, 0, len(m.notes))
	for _, entity := range m.notes {
		if note, ok := entity.(*model.Note); ok {
			notes = append(notes, note)
		}
	}
	return notes
}

func (m *MockRepository) GetNotesCount() int {
	count := 0
	for _, entity := range m.notes {
		if _, ok := entity.(*model.Note); ok {
			count++
		}
	}
	return count
}

func (m *MockRepository) GetNewNotes(lastIndex int) []*model.Note {
	// В мок-репозитории возвращаем все заметки, так как у нас нет временной метки
	notes := m.GetAllNotes()
	if lastIndex >= len(notes) {
		return []*model.Note{}
	}
	return notes[lastIndex:]
}

func (m *MockRepository) GetAllByType(entityType string) []repository.Entity {
	entities := make([]repository.Entity, 0, len(m.notes))
	for _, entity := range m.notes {
		if entity.GetType() == entityType {
			entities = append(entities, entity)
		}
	}
	return entities
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCreateNote(t *testing.T) {
	repo := NewMockRepository()
	handler := NewNoteHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/notes", handler.CreateNote)

	// Подготовка тестовых данных
	requestBody := createNoteRequest{
		Title:   "Test Note",
		Content: "Test Content",
	}
	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/notes", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var responseNote model.Note
	err := json.Unmarshal(w.Body.Bytes(), &responseNote)
	assert.NoError(t, err)
	assert.Equal(t, "Test Note", responseNote.GetTitle())
	assert.Equal(t, "Test Content", responseNote.GetContent())
	assert.NotEmpty(t, responseNote.GetID())
}

func TestGetNote(t *testing.T) {
	// Создаем мок-репозиторий с тестовой заметкой
	repo := NewMockRepository()
	testNote := model.NewNote("Test Title", "Test Content")
	repo.Save(testNote)

	handler := NewNoteHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/notes/:id", handler.GetNote)

	req, _ := http.NewRequest(http.MethodGet, "/api/notes/"+testNote.GetID(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseNote model.Note
	err := json.Unmarshal(w.Body.Bytes(), &responseNote)
	assert.NoError(t, err)
	assert.Equal(t, testNote.GetID(), responseNote.GetID())
	assert.Equal(t, testNote.GetTitle(), responseNote.GetTitle())
	assert.Equal(t, testNote.GetContent(), responseNote.GetContent())
}

func TestGetNoteNotFound(t *testing.T) {
	repo := NewMockRepository()
	handler := NewNoteHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/notes/:id", handler.GetNote)

	req, _ := http.NewRequest(http.MethodGet, "/api/notes/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateNote(t *testing.T) {
	// Создаем мок-репозиторий с тестовой заметкой
	repo := NewMockRepository()
	testNote := model.NewNote("Original Title", "Original Content")
	repo.Save(testNote)

	handler := NewNoteHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/api/notes/:id", handler.UpdateNote)

	// Подготовка обновленных данных
	updatedData := map[string]interface{}{
		"title":   "Updated Title",
		"content": "Updated Content",
	}
	jsonData, _ := json.Marshal(updatedData)

	req, _ := http.NewRequest(http.MethodPut, "/api/notes/"+testNote.GetID(), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseNote model.Note
	err := json.Unmarshal(w.Body.Bytes(), &responseNote)
	assert.NoError(t, err)
	assert.Equal(t, testNote.GetID(), responseNote.GetID())
	assert.Equal(t, "Updated Title", responseNote.GetTitle())
	assert.Equal(t, "Updated Content", responseNote.GetContent())
}

func TestUpdateNoteNotFound(t *testing.T) {
	repo := NewMockRepository()
	handler := NewNoteHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/api/notes/:id", handler.UpdateNote)

	updatedData := map[string]interface{}{
		"title":   "Updated Title",
		"content": "Updated Content",
	}
	jsonData, _ := json.Marshal(updatedData)

	req, _ := http.NewRequest(http.MethodPut, "/api/notes/nonexistent", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteNote(t *testing.T) {
	// Создаем мок-репозиторий с тестовой заметкой
	repo := NewMockRepository()
	testNote := model.NewNote("Test Title", "Test Content")
	repo.Save(testNote)

	handler := NewNoteHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/notes/:id", handler.DeleteNote)

	req, _ := http.NewRequest(http.MethodDelete, "/api/notes/"+testNote.GetID(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Проверяем, что заметка действительно удалена
	assert.Nil(t, repo.GetByID("note", testNote.GetID()))
}

func TestDeleteNoteNotFound(t *testing.T) {
	repo := NewMockRepository()
	handler := NewNoteHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/notes/:id", handler.DeleteNote)

	req, _ := http.NewRequest(http.MethodDelete, "/api/notes/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetAllNotes(t *testing.T) {
	// Создаем мок-репозиторий с несколькими заметками
	repo := NewMockRepository()
	note1 := model.NewNote("Title 1", "Content 1")
	note2 := model.NewNote("Title 2", "Content 2")
	repo.Save(note1)
	repo.Save(note2)

	handler := NewNoteHandler(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/notes", handler.GetAllNotes)

	req, _ := http.NewRequest(http.MethodGet, "/api/notes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseNotes []model.Note
	err := json.Unmarshal(w.Body.Bytes(), &responseNotes)
	assert.NoError(t, err)
	assert.Len(t, responseNotes, 2)

	// Проверяем, что обе заметки присутствуют в ответе
	foundNote1 := false
	foundNote2 := false
	for _, note := range responseNotes {
		if note.GetID() == note1.GetID() {
			foundNote1 = true
			assert.Equal(t, "Title 1", note.GetTitle())
			assert.Equal(t, "Content 1", note.GetContent())
		}
		if note.GetID() == note2.GetID() {
			foundNote2 = true
			assert.Equal(t, "Title 2", note.GetTitle())
			assert.Equal(t, "Content 2", note.GetContent())
		}
	}
	assert.True(t, foundNote1)
	assert.True(t, foundNote2)
}
