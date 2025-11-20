package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/domain/service"
)

// NoteHandler структура для обработки HTTP запросов, связанных с заметками
type NoteHandler struct {
	noteService service.NoteService
}

// NewNoteHandler создает новый экземпляр NoteHandler
func NewNoteHandler(noteService service.NoteService) *NoteHandler {
	return &NoteHandler{
		noteService: noteService,
	}
}

// CreateNote создает новую заметку
// @Summary Создать новую заметку
// @Description Создает новую заметку с указанными заголовком и содержимым
// @Tags notes
// @Accept json
// @Produce json
// @Param note body createNoteRequest true "Заметка"
// @Success 201 {object} Note
// @Failure 400 {object} map[string]string
// @Router /notes [post]
func (h *NoteHandler) CreateNote(c *gin.Context) {
	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID пользователя из контекста (предполагается, что он был добавлен middleware аутентификации)
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не аутентифицирован"})
		return
	}

	userIdStr, ok := userId.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения ID пользователя"})
		return
	}

	note, err := h.noteService.CreateNote(req.Title, req.Content, userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// createNoteRequest структура для запроса создания заметки
type createNoteRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// GetNote возвращает заметку по ID
// @Summary Получить заметку по ID
// @Description Возвращает заметку по указанному ID
// @Tags notes
// @Produce json
// @Param id path string true "ID заметки"
// @Success 200 {object} Note
// @Failure 404 {object} map[string]string
// @Router /notes/{id} [get]
func (h *NoteHandler) GetNote(c *gin.Context) {
	id := c.Param("id")
	note, err := h.noteService.GetNoteByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// UpdateNote обновляет заметку
// @Summary Обновить заметку
// @Description Обновляет заметку с указанным ID
// @Tags notes
// @Accept json
// @Produce json
// @Param id path string true "ID заметки"
// @Param note body model.Note true "Обновленная заметка"
// @Success 200 {object} model.Note
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /notes/{id} [put]
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	id := c.Param("id")
	var updatedNote model.Note
	if err := c.ShouldBindJSON(&updatedNote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.noteService.UpdateNote(id, updatedNote.GetTitle(), updatedNote.GetContent())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote удаляет заметку
// @Summary Удалить заметку
// @Description Удаляет заметку с указанным ID
// @Tags notes
// @Produce json
// @Param id path string true "ID заметки"
// @Success 204 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /notes/{id} [delete]
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	id := c.Param("id")
	err := h.noteService.DeleteNote(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Note deleted successfully"})
}

// GetAllNotes возвращает все заметки
// @Summary Получить все заметки
// @Description Возвращает список всех заметок
// @Tags notes
// @Produce json
// @Success 200 {array} model.Note
// @Router /notes [get]
func (h *NoteHandler) GetAllNotes(c *gin.Context) {
	notes, err := h.noteService.GetAllNotes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notes"})
		return
	}
	c.JSON(http.StatusOK, notes)
}
