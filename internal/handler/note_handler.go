package handler

import (
	"log"
	"net/http"
	_ "strconv"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
)

// NoteHandler структура для обработки HTTP запросов, связанных с заметками
type NoteHandler struct {
	repo repository.Repository
}

// NewNoteHandler создает новый экземпляр NoteHandler
func NewNoteHandler(repo repository.Repository) *NoteHandler {
	return &NoteHandler{
		repo: repo,
	}
}

// CreateNote создает новую заметку
// @Summary Создать новую заметку
// @Description Создает новую заметку с указанными заголовком и содержимым
// @Tags notes
// @Accept json
// @Produce json
// @Param note body createNoteRequest true "Заметка"
// @Success 201 {object} model.Note
// @Failure 400 {object} map[string]string
// @Router /api/notes [post]
func (h *NoteHandler) CreateNote(c *gin.Context) {
	log.Printf("CreateNote handler вызван")
	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Ошибка при привязке JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Получен запрос на создание заметки: title='%s', content='%s'", req.Title, req.Content)
	note := model.NewNote(req.Title, req.Content)
	log.Printf("Создана новая заметка с ID: %s", note.GetID())
	h.repo.Save(note)
	log.Printf("Заметка успешно сохранена в репозиторий")
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
// @Success 200 {object} model.Note
// @Failure 404 {object} map[string]string
// @Router /api/notes/{id} [get]
func (h *NoteHandler) GetNote(c *gin.Context) {
	id := c.Param("id")
	entity := h.repo.GetByID("note", id)
	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	note, ok := entity.(*model.Note)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cast entity to note"})
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
// @Router /api/notes/{id} [put]
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	id := c.Param("id")
	entity := h.repo.GetByID("note", id)
	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	note, ok := entity.(*model.Note)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cast entity to note"})
		return
	}

	var updatedNote model.Note
	if err := c.ShouldBindJSON(&updatedNote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note.SetTitle(updatedNote.GetTitle())
	note.SetContent(updatedNote.GetContent())

	h.repo.Save(note)
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
// @Router /api/notes/{id} [delete]
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	id := c.Param("id")
	deleted := h.repo.DeleteByID("note", id)
	if !deleted {
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
// @Router /api/notes [get]
func (h *NoteHandler) GetAllNotes(c *gin.Context) {
	notes := h.repo.GetAllNotes()
	c.JSON(http.StatusOK, notes)
}
