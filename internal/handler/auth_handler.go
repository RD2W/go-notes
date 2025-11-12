package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
)

// AuthHandler структура для обработки HTTP запросов, связанных с аутентификацией
type AuthHandler struct {
	repo         repository.Repository
	tokenManager *auth.TokenManager
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(repo repository.Repository, tokenManager *auth.TokenManager) *AuthHandler {
	return &AuthHandler{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

// Login обрабатывает аутентификацию пользователя и возвращает пару токенов
// @Summary Аутентификация пользователя
// @Description Аутентифицирует пользователя и возвращает access и refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body loginRequest true "Учетные данные"
// @Success 200 {object} loginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Находим пользователя по имени
	entities := h.repo.GetAllByType("user")
	var user *model.User
	for _, entity := range entities {
		u, ok := entity.(*model.User)
		if ok && u.GetUsername() == req.Username {
			user = u
			break
		}
	}

	if user == nil || !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	accessToken, refreshToken, err := h.tokenManager.GenerateTokens(user.GetUsername())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.tokenManager.GetJWTExpirationSeconds()), // использовать фактическое время жизни токена
	})
}

// Logout обрабатывает выход пользователя и отзыв refresh токена
// @Summary Выход пользователя
// @Description Выходит пользователя и отзывает refresh токен
// @Tags auth
// @Accept json
// @Produce json
// @Param logoutRequest body logoutRequest true "Данные для выхода"
// @Success 200 {object} logoutResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.tokenManager.Logout(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, logoutResponse{
		Success: true,
		Message: "Successfully logged out",
	})
}

// Refresh обновляет пару токенов по refresh токену
// @Summary Обновление токенов
// @Description Обновляет access и refresh токены по старому refresh токену
// @Tags auth
// @Accept json
// @Produce json
// @Param refreshRequest body refreshRequest true "Запрос на обновление токенов"
// @Success 200 {object} refreshResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccessToken, newRefreshToken, err := h.tokenManager.RefreshTokens(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Вычисляем время жизни токена из токен-менеджера
	expiresIn := int(h.tokenManager.GetJWTExpirationSeconds())

	c.JSON(http.StatusOK, refreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	})
}

// ValidateToken проверяет валидность токена
// @Summary Проверка токена
// @Description Проверяет валидность предоставленного токена
// @Tags auth
// @Accept json
// @Produce json
// @Param validateRequest body validateRequest true "Запрос на проверку токена"
// @Success 200 {object} validateResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/validate [post]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req validateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.tokenManager.ValidateAccessToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"valid": false, "error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, validateResponse{
		Valid:    true,
		Username: claims.Username,
		Expires:  claims.ExpiresAt.Unix(),
	})
}

// loginRequest структура для запроса аутентификации
type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// loginResponse структура для ответа аутентификации
type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // Время жизни токена в секундах
}

// logoutRequest структура для запроса выхода
type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// logoutResponse структура для ответа выхода
type logoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// refreshRequest структура для запроса обновления токенов
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// refreshResponse структура для ответа обновления токенов
type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // Время жизни токена в секундах
}

// validateRequest структура для запроса проверки токена
type validateRequest struct {
	Token string `json:"token" binding:"required"`
}

// validateResponse структура для ответа проверки токена
type validateResponse struct {
	Valid    bool   `json:"valid"`
	Username string `json:"username"`
	Expires  int64  `json:"expires"`
}
