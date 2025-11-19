package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
)

// UserHandler структура для обработки HTTP запросов, связанных с пользователями
type UserHandler struct {
	repo repository.Repository
}

// createUserRequest структура для запроса создания пользователя
type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// NewUserHandler создает новый экземпляр UserHandler
func NewUserHandler(repo repository.Repository) *UserHandler {
	return &UserHandler{
		repo: repo,
	}
}

// CreateUser создает нового пользователя
// @Summary Создать нового пользователя
// @Description Создает нового пользователя с указанными данными
// @Tags users
// @Accept json
// @Produce json
// @Param user body createUserRequest true "Пользователь"
// @Success 201 {object} model.User
// @Failure 400 {object} map[string]string
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := model.NewUser(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	h.repo.Save(user)
	c.JSON(http.StatusCreated, user)
}

// GetUser возвращает пользователя по ID
// @Summary Получить пользователя по ID
// @Description Возвращает пользователя по указанному ID
// @Tags users
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {object} model.User
// @Failure 404 {object} map[string]string
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	entity := h.repo.GetByID("user", id)
	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user, ok := entity.(*model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cast entity to user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser обновляет пользователя
// @Summary Обновить пользователя
// @Description Обновляет пользователя с указанным ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param user body model.User true "Обновленный пользователь"
// @Success 200 {object} model.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	entity := h.repo.GetByID("user", id)
	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user, ok := entity.(*model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cast entity to user"})
		return
	}

	var updatedUser model.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.SetUsername(updatedUser.GetUsername())
	user.SetEmail(updatedUser.GetEmail())

	h.repo.Save(user)
	c.JSON(http.StatusOK, user)
}

// DeleteUser удаляет пользователя
// @Summary Удалить пользователя
// @Description Удаляет пользователя с указанным ID
// @Tags users
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 204 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	deleted := h.repo.DeleteByID("user", id)
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "User deleted successfully"})
}

// GetAllUsers возвращает всех пользователей
// @Summary Получить всех пользователей
// @Description Возвращает список всех пользователей
// @Tags users
// @Produce json
// @Success 200 {array} model.User
// @Router /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	entities := h.repo.GetAllByType("user")
	users := make([]*model.User, 0)

	for _, entity := range entities {
		user, ok := entity.(*model.User)
		if !ok {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}
