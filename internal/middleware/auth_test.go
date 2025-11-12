package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Устанавливаем тестовый ключ для JWT
	jwtKey = []byte("test_secret_key_for_testing")
}

func TestGenerateJWT(t *testing.T) {
	username := "testuser"
	tokenString, err := GenerateJWT(username)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Проверяем, что токен может быть расшифрован
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	assert.NoError(t, err)
	assert.True(t, token.Valid)
	assert.Equal(t, username, claims.Username)

	// Проверяем, что токен истекает в течение 24 часов
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), time.Unix(claims.ExpiresAt.Unix(), 0), 10*time.Second)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем валидный токен
	username := "testuser"
	tokenString, _ := GenerateJWT(username)

	// Создаем запрос с валидным токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware()
	authMiddleware(c)

	// Проверяем, что запрос не был прерван
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что username был установлен в контексте
	usernameFromContext, exists := c.Get("username")
	assert.True(t, exists)
	assert.Equal(t, username, usernameFromContext)
}

func TestAuthMiddleware_NoAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем запрос без заголовка Authorization
	req, _ := http.NewRequest("GET", "/test", nil)

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware()
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Authorization header is required")
}

func TestAuthMiddleware_InvalidPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем запрос с неверным префиксом в заголовке Authorization
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidPrefix token123")

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware()
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Bearer token is required")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем запрос с невалидным токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token_string")

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware()
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем истекший токен
	expiredClaims := &Claims{
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Токен истек час назад
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredToken, _ := token.SignedString(jwtKey)

	// Создаем запрос с истекшим токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware()
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем запрос с неправильно сформированным токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer malformed_token")

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware()
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")
}
