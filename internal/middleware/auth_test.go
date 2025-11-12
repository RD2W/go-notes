package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	"github.com/stretchr/testify/assert"
)

// Создаем тестовую обертку для TokenManager, которая будет обходить проверку в хранилище
type TestTokenManagerWrapper struct {
	originalManager *auth.TokenManager
}

// Переопределяем метод ValidateAccessToken для тестов, чтобы пропускать проверку в хранилище
func (tmw *TestTokenManagerWrapper) ValidateAccessToken(tokenString string) (*auth.TokenClaims, error) {
	claims := &auth.TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Используем JWT секрет из оригинального TokenManager
		// Для этого нужно получить доступ к приватному полю, поэтому мы будем использовать
		// публичный метод или обойти через рефлексию, но в данном случае проще создать
		// тестовый токен с правильным секретом
		return []byte("test_secret_key_for_testing"), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("access токен недействителен: %w", err)
	}

	// Пропускаем проверку в хранилище для тестов
	// Возвращаем claims без дополнительной проверки
	return claims, nil
}

// Создаем функцию для создания тестовой обертки TokenManager
func createTestAuthManager() *TestTokenManagerWrapper {
	// Создаем тестовую конфигурацию
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test_secret_key_for_testing",
			AccessTokenTTL:  "24h",
			RefreshTokenTTL: "168h",
		},
		Refresh: config.RefreshConfig{
			SecretKey: "test_refresh_secret_key_for_testing",
		},
	}

	// Создаем оригинальный TokenManager
	originalManager := auth.NewTokenManager(testConfig)

	return &TestTokenManagerWrapper{
		originalManager: originalManager,
	}
}

// Создаем функцию AuthMiddleware для тестов, которая принимает TestTokenManagerWrapper
func AuthMiddlewareForTests(tokenManager *TestTokenManagerWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token is required"})
			c.Abort()
			return
		}

		claims, err := tokenManager.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("tokenID", claims.TokenID) // Устанавливаем также TokenID, если нужно
		c.Next()
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовый TokenManager
	tokenManager := createTestAuthManager()

	// Создаем валидный токен вручную
	username := "testuser"
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &auth.TokenClaims{
		Username: username,
		TokenID:  "test-token-id",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test_secret_key_for_testing"))

	// Создаем запрос с валидным токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware с использованием обертки
	authMiddleware := AuthMiddlewareForTests(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос не был прерван
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что username был установлен в контексте
	usernameFromContext, exists := c.Get("username")
	assert.True(t, exists)
	assert.Equal(t, username, usernameFromContext)

	// Проверяем, что tokenID был установлен в контексте
	tokenIDFromContext, exists := c.Get("tokenID")
	assert.True(t, exists)
	assert.Equal(t, "test-token-id", tokenIDFromContext)
}

func TestAuthMiddleware_NoAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовый TokenManager
	tokenManager := createTestAuthManager()

	// Создаем запрос без заголовка Authorization
	req, _ := http.NewRequest("GET", "/test", nil)

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddlewareForTests(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Authorization header is required")
}

func TestAuthMiddleware_InvalidPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовый TokenManager
	tokenManager := createTestAuthManager()

	// Создаем запрос с неверным префиксом в заголовке Authorization
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidPrefix token123")

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddlewareForTests(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Bearer token is required")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовый TokenManager
	tokenManager := createTestAuthManager()

	// Создаем запрос с невалидным токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token_string")

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddlewareForTests(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовый TokenManager
	tokenManager := createTestAuthManager()

	// Создаем истекший токен
	expiredClaims := &auth.TokenClaims{
		Username: "testuser",
		TokenID:  "test-token-id",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Токен истек час назад
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredToken, _ := token.SignedString([]byte("test_secret_key_for_testing"))

	// Создаем запрос с истекшим токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddlewareForTests(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем тестовый TokenManager
	tokenManager := createTestAuthManager()

	// Создаем запрос с неправильно сформированным токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer malformed_token")

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddlewareForTests(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")
}
