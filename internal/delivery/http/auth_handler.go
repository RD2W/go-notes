package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/domain/service"
)

// AuthHandler структура для обработки HTTP запросов, связанных с аутентификацией
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
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
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
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
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.Logout(req.RefreshToken)
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
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, refreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
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
// @Router /auth/validate [post]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req validateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"valid": false, "error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, validateResponse{
		Valid: true,
	})
}

// refreshResponse структура для ответа обновления токенов
type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// validateRequest структура для запроса проверки токена
type validateRequest struct {
	Token string `json:"token" binding:"required"`
}

// validateResponse структура для ответа проверки токена
type validateResponse struct {
	Valid    bool   `json:"valid"`
	Username string `json:"username,omitempty"`
	Expires  int64  `json:"expires,omitempty"`
}
