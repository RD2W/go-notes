package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService - mock для сервиса аутентификации
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(username, password string) (string, string, error) {
	args := m.Called(username, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) Logout(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) RefreshTokens(refreshToken string) (string, string, error) {
	args := m.Called(refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) ValidateToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := new(MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)

	t.Run("successful login", func(t *testing.T) {
		// Подготовка
		expectedAccessToken := "access_token"
		expectedRefreshToken := "refresh_token"
		username := "testuser"
		password := "password123"

		mockAuthService.On("Login", username, password).Return(expectedAccessToken, expectedRefreshToken, nil).Once()

		loginReq := loginRequest{
			Username: username,
			Password: password,
		}
		jsonData, _ := json.Marshal(loginReq)

		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.Login(c)

		// Проверка
		assert.Equal(t, http.StatusOK, w.Code)

		var response loginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedAccessToken, response.AccessToken)
		assert.Equal(t, expectedRefreshToken, response.RefreshToken)
		assert.Equal(t, "Bearer", response.TokenType)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		// Подготовка
		jsonData := []byte(`{"username": "testuser"}`) // Отсутствует password

		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.Login(c)

		// Проверка
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		// Подготовка
		username := "testuser"
		password := "wrongpassword"

		mockAuthService.On("Login", username, password).Return("", "", assert.AnError).Once()

		loginReq := loginRequest{
			Username: username,
			Password: password,
		}
		jsonData, _ := json.Marshal(loginReq)

		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.Login(c)

		// Проверка
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := new(MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)

	t.Run("successful logout", func(t *testing.T) {
		// Подготовка
		refreshToken := "refresh_token_123"

		mockAuthService.On("Logout", refreshToken).Return(nil).Once()

		logoutReq := logoutRequest{
			RefreshToken: refreshToken,
		}
		jsonData, _ := json.Marshal(logoutReq)

		req, _ := http.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.Logout(c)

		// Проверка
		assert.Equal(t, http.StatusOK, w.Code)

		var response logoutResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Successfully logged out", response.Message)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		// Подготовка
		jsonData := []byte(`{}`) // Отсутствует refresh_token

		req, _ := http.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.Logout(c)

		// Проверка
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		// Подготовка
		refreshToken := "invalid_refresh_token"

		mockAuthService.On("Logout", refreshToken).Return(assert.AnError).Once()

		logoutReq := logoutRequest{
			RefreshToken: refreshToken,
		}
		jsonData, _ := json.Marshal(logoutReq)

		req, _ := http.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.Logout(c)

		// Проверка
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := new(MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)

	t.Run("successful token refresh", func(t *testing.T) {
		// Подготовка
		refreshToken := "old_refresh_token"
		newAccessToken := "new_access_token"
		newRefreshToken := "new_refresh_token"

		mockAuthService.On("RefreshTokens", refreshToken).Return(newAccessToken, newRefreshToken, nil).Once()

		refreshReq := refreshRequest{
			RefreshToken: refreshToken,
		}
		jsonData, _ := json.Marshal(refreshReq)

		req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.Refresh(c)

		// Проверка
		assert.Equal(t, http.StatusOK, w.Code)

		var response refreshResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, newAccessToken, response.AccessToken)
		assert.Equal(t, newRefreshToken, response.RefreshToken)
		assert.Equal(t, "Bearer", response.TokenType)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		// Подготовка
		jsonData := []byte(`{}`) // Отсутствует refresh_token

		req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.Refresh(c)

		// Проверка
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		// Подготовка
		refreshToken := "invalid_refresh_token"

		mockAuthService.On("RefreshTokens", refreshToken).Return("", "", assert.AnError).Once()

		refreshReq := refreshRequest{
			RefreshToken: refreshToken,
		}
		jsonData, _ := json.Marshal(refreshReq)

		req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.Refresh(c)

		// Проверка
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthHandler_ValidateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := new(MockAuthService)
	authHandler := NewAuthHandler(mockAuthService)

	t.Run("valid token", func(t *testing.T) {
		// Подготовка
		token := "valid_token"
		claims := "testuser"

		mockAuthService.On("ValidateToken", token).Return(claims, nil).Once()

		validateReq := validateRequest{
			Token: token,
		}
		jsonData, _ := json.Marshal(validateReq)

		req, _ := http.NewRequest(http.MethodPost, "/auth/validate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.ValidateToken(c)

		// Проверка
		assert.Equal(t, http.StatusOK, w.Code)

		var response validateResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Valid)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		// Подготовка
		jsonData := []byte(`{}`) // Отсутствует token

		req, _ := http.NewRequest(http.MethodPost, "/auth/validate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.ValidateToken(c)

		// Проверка
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		// Подготовка
		token := "invalid_token"

		mockAuthService.On("ValidateToken", token).Return("", assert.AnError).Once()

		validateReq := validateRequest{
			Token: token,
		}
		jsonData, _ := json.Marshal(validateReq)

		req, _ := http.NewRequest(http.MethodPost, "/auth/validate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Выполнение
		authHandler.ValidateToken(c)

		// Проверка
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		mockAuthService.AssertExpectations(t)
	})
}
