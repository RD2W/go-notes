package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTokenRepository - mock для TokenRepository
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) AddToBlacklist(tokenID string, expiresAt time.Time) error {
	args := m.Called(tokenID, expiresAt)
	return args.Error(0)
}

func (m *MockTokenRepository) IsBlacklisted(tokenID string) (bool, error) {
	args := m.Called(tokenID)
	return args.Bool(0), args.Error(1)
}

// Helper functions for creating test tokens
func createValidToken(secretKey, username, tokenID string, duration time.Duration) string {
	claims := &auth.TokenClaims{
		Username: username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secretKey))
	return tokenString
}

func createExpiredToken(secretKey, username, tokenID string) string {
	claims := &auth.TokenClaims{
		Username: username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Токен истек час назад
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secretKey))
	return tokenString
}

func createTokenWithInvalidSignature(username, tokenID string) string {
	claims := &auth.TokenClaims{
		Username: username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("different_secret_key")) // Используем другой секрет для создания токена
	return tokenString
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем mock репозитория токенов
	mockRepo := new(MockTokenRepository)
	mockRepo.On("IsBlacklisted", "test-token-id").Return(false, nil)

	// Создаем TokenManager с mock репозиторием
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
	tokenManager := auth.NewTokenManagerWithStore(testConfig, mockRepo)

	// Создаем валидный токен
	tokenString := createValidToken("test_secret_key_for_testing", "testuser", "test-token-id", 1*time.Hour)

	// Создаем запрос с валидным токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос не был прерван
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что username был установлен в контексте
	usernameFromContext, exists := c.Get("username")
	assert.True(t, exists)
	assert.Equal(t, "testuser", usernameFromContext)

	// Проверяем, что tokenID был установлен в контексте
	tokenIDFromContext, exists := c.Get("tokenID")
	assert.True(t, exists)
	assert.Equal(t, "test-token-id", tokenIDFromContext)

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestAuthMiddleware_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		authHeader         string
		tokenString        string
		setupMock          func() *MockTokenRepository
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:               "No Auth Header",
			authHeader:         "",
			tokenString:        "",
			setupMock:          func() *MockTokenRepository { return new(MockTokenRepository) },
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "Authorization header is required",
		},
		{
			name:               "Invalid Prefix",
			authHeader:         "InvalidPrefix token123",
			tokenString:        "",
			setupMock:          func() *MockTokenRepository { return new(MockTokenRepository) },
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "Bearer token is required",
		},
		{
			name:        "Valid Token",
			authHeader:  "Bearer ",
			tokenString: createValidToken("test_secret_key_for_testing", "testuser", "test-token-id", 1*time.Hour),
			setupMock: func() *MockTokenRepository {
				mockRepo := new(MockTokenRepository)
				mockRepo.On("IsBlacklisted", "test-token-id").Return(false, nil)
				return mockRepo
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "",
		},
		{
			name:        "Expired Token",
			authHeader:  "Bearer ",
			tokenString: createExpiredToken("test_secret_key_for_testing", "testuser", "test-token-id"),
			setupMock: func() *MockTokenRepository {
				mockRepo := new(MockTokenRepository)
				// Для истекшего токена вызов IsBlacklisted не происходит, т.к. валидация прерывается раньше
				return mockRepo
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "Invalid token",
		},
		{
			name:        "Invalid Signature",
			authHeader:  "Bearer ",
			tokenString: createTokenWithInvalidSignature("testuser", "test-token-id"),
			setupMock: func() *MockTokenRepository {
				mockRepo := new(MockTokenRepository)
				// Для токена с неверной подписью вызов IsBlacklisted не происходит
				return mockRepo
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "Invalid token",
		},
		{
			name:        "Token in Blacklist",
			authHeader:  "Bearer ",
			tokenString: createValidToken("test_secret_key_for_testing", "testuser", "blacklisted-token-id", 1*time.Hour),
			setupMock: func() *MockTokenRepository {
				mockRepo := new(MockTokenRepository)
				mockRepo.On("IsBlacklisted", "blacklisted-token-id").Return(true, nil)
				return mockRepo
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "Invalid token",
		},
		{
			name:        "Token Store Error",
			authHeader:  "Bearer ",
			tokenString: createValidToken("test_secret_key_for_testing", "testuser", "error-token-id", 1*time.Hour),
			setupMock: func() *MockTokenRepository {
				mockRepo := new(MockTokenRepository)
				mockRepo.On("IsBlacklisted", "error-token-id").Return(false, assert.AnError)
				return mockRepo
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "Invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.setupMock()

			// Создаем TokenManager с mock репозиторием
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
			tokenManager := auth.NewTokenManagerWithStore(testConfig, mockRepo)

			// Создаем запрос
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				if tt.tokenString != "" {
					req.Header.Set("Authorization", tt.authHeader+tt.tokenString)
				} else {
					req.Header.Set("Authorization", tt.authHeader)
				}
			}

			// Создаем gin контекст
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Применяем middleware
			authMiddleware := AuthMiddleware(tokenManager)
			authMiddleware(c)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatusCode, w.Code)

			// Проверяем тело ответа, если ожидается ошибка
			if tt.expectedResponse != "" {
				assert.Contains(t, w.Body.String(), tt.expectedResponse)
			}

			// Проверяем, что mock был вызван (если были установлены ожидания)
			if tt.name != "No Auth Header" && tt.name != "Invalid Prefix" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestAuthMiddleware_BlacklistedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем mock репозитория токенов
	mockRepo := new(MockTokenRepository)
	mockRepo.On("IsBlacklisted", "blacklisted-token-id").Return(true, nil)

	// Создаем TokenManager с mock репозиторием
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
	tokenManager := auth.NewTokenManagerWithStore(testConfig, mockRepo)

	// Создаем токен, который будет в черном списке
	tokenString := createValidToken("test_secret_key_for_testing", "testuser", "blacklisted-token-id", 1*time.Hour)

	// Создаем запрос с токеном из черного списка
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestAuthMiddleware_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем mock репозитория токенов, который возвращает ошибку
	mockRepo := new(MockTokenRepository)
	mockRepo.On("IsBlacklisted", "error-token-id").Return(false, assert.AnError)

	// Создаем TokenManager с mock репозиторием
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
	tokenManager := auth.NewTokenManagerWithStore(testConfig, mockRepo)

	// Создаем валидный токен
	tokenString := createValidToken("test_secret_key_for_testing", "testuser", "error-token-id", 1*time.Hour)

	// Создаем запрос с токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401 из-за ошибки в хранилище
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestAuthMiddleware_EmptyTokenString(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем mock репозитория токенов
	mockRepo := new(MockTokenRepository)

	// Создаем TokenManager с mock репозиторием
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
	tokenManager := auth.NewTokenManagerWithStore(testConfig, mockRepo)

	// Создаем запрос с пустым токеном (после Bearer идет пустая строка)
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")

	// Проверяем, что mock не был вызван, так как токен пустой
	mockRepo.AssertNotCalled(t, "IsBlacklisted")
}

func TestAuthMiddleware_WhitespaceOnlyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем mock репозитория токенов
	mockRepo := new(MockTokenRepository)

	// Создаем TokenManager с mock репозиторием
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
	tokenManager := auth.NewTokenManagerWithStore(testConfig, mockRepo)

	// Создаем запрос с токеном, состоящим только из пробелов
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer \t ")

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Invalid token")

	// Проверяем, что mock не был вызван, так как токен пустой
	mockRepo.AssertNotCalled(t, "IsBlacklisted")
}

func TestAuthMiddleware_CaseInsensitiveBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем mock репозитория токенов
	mockRepo := new(MockTokenRepository)

	// Создаем TokenManager с mock репозиторием
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
	tokenManager := auth.NewTokenManagerWithStore(testConfig, mockRepo)

	// Создаем запрос с токеном в нижнем регистре (должно быть отклонено)
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "bearer "+createValidToken("test_secret_key_for_testing", "testuser", "test-token-id", 1*time.Hour))

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос был прерван с ошибкой 401 (так как "bearer" не равно "Bearer")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Проверяем, что в теле ответа содержится ошибка
	assert.Contains(t, w.Body.String(), "Bearer token is required")

	// Проверяем, что mock не был вызван
	mockRepo.AssertNotCalled(t, "IsBlacklisted")
}

func TestAuthMiddleware_MultipleAuthHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Создаем mock репозитория токенов
	mockRepo := new(MockTokenRepository)
	mockRepo.On("IsBlacklisted", "test-token-id").Return(false, nil)

	// Создаем TokenManager с mock репозиторием
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
	tokenManager := auth.NewTokenManagerWithStore(testConfig, mockRepo)

	// Создаем валидный токен
	tokenString := createValidToken("test_secret_key_for_testing", "testuser", "test-token-id", 1*time.Hour)

	// Создаем запрос с несколькими заголовками Authorization
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header["Authorization"] = []string{
		"Bearer " + tokenString,
		"Bearer some-other-token",
	}

	// Создаем gin контекст
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Применяем middleware
	authMiddleware := AuthMiddleware(tokenManager)
	authMiddleware(c)

	// Проверяем, что запрос не был прерван
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что username был установлен в контексте (берется первый заголовок)
	usernameFromContext, exists := c.Get("username")
	assert.True(t, exists)
	assert.Equal(t, "testuser", usernameFromContext)

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}
