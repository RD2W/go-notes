package grpc

import (
	"context"
	"errors"
	"testing"

	authPb "github.com/rd2w/go-notes/pkg/proto/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService - мок для сервиса аутентификации
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

func TestAuthServiceServer_Login(t *testing.T) {
	mockAuthService := new(MockAuthService)
	authServer := NewAuthServiceServer(mockAuthService)

	ctx := context.Background()
	req := &authPb.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	t.Run("successful login", func(t *testing.T) {
		expectedAccessToken := "access_token_123"
		expectedRefreshToken := "refresh_token_123"

		mockAuthService.On("Login", req.Username, req.Password).Return(expectedAccessToken, expectedRefreshToken, nil).Once()

		resp, err := authServer.Login(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedAccessToken, resp.AccessToken)
		assert.Equal(t, expectedRefreshToken, resp.RefreshToken)
		assert.Equal(t, "Bearer", resp.TokenType)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("login error", func(t *testing.T) {
		mockAuthService.On("Login", req.Username, req.Password).Return("", "", errors.New("invalid credentials")).Once()

		resp, err := authServer.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthServiceServer_Logout(t *testing.T) {
	mockAuthService := new(MockAuthService)
	authServer := NewAuthServiceServer(mockAuthService)

	ctx := context.Background()
	req := &authPb.LogoutRequest{
		RefreshToken: "refresh_token_123",
	}

	t.Run("successful logout", func(t *testing.T) {
		mockAuthService.On("Logout", req.RefreshToken).Return(nil).Once()

		resp, err := authServer.Logout(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, "successful logout", resp.Message)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("logout error", func(t *testing.T) {
		mockAuthService.On("Logout", req.RefreshToken).Return(errors.New("logout failed")).Once()

		resp, err := authServer.Logout(ctx, req)

		assert.NoError(t, err) // gRPC метод возвращает nil ошибку даже при внутренней ошибке
		assert.NotNil(t, resp)
		assert.False(t, resp.Success)
		assert.Equal(t, "failed to logout", resp.Message)

		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthServiceServer_Refresh(t *testing.T) {
	mockAuthService := new(MockAuthService)
	authServer := NewAuthServiceServer(mockAuthService)

	ctx := context.Background()
	req := &authPb.RefreshRequest{
		RefreshToken: "refresh_token_123",
	}

	t.Run("successful token refresh", func(t *testing.T) {
		expectedNewAccessToken := "new_access_token_123"
		expectedNewRefreshToken := "new_refresh_token_123"

		mockAuthService.On("RefreshTokens", req.RefreshToken).Return(expectedNewAccessToken, expectedNewRefreshToken, nil).Once()

		resp, err := authServer.Refresh(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedNewAccessToken, resp.AccessToken)
		assert.Equal(t, expectedNewRefreshToken, resp.RefreshToken)
		assert.Equal(t, "Bearer", resp.TokenType)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("token refresh error", func(t *testing.T) {
		mockAuthService.On("RefreshTokens", req.RefreshToken).Return("", "", errors.New("refresh failed")).Once()

		resp, err := authServer.Refresh(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthServiceServer_ValidateToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	authServer := NewAuthServiceServer(mockAuthService)

	ctx := context.Background()
	req := &authPb.ValidateTokenRequest{
		Token: "valid_token_123",
	}

	t.Run("valid token", func(t *testing.T) {
		expectedUsername := "testuser"

		mockAuthService.On("ValidateToken", req.Token).Return(expectedUsername, nil).Once()

		resp, err := authServer.ValidateToken(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Valid)
		assert.Equal(t, expectedUsername, resp.Username)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("invalid token", func(t *testing.T) {
		mockAuthService.On("ValidateToken", req.Token).Return("", errors.New("token is invalid")).Once()

		resp, err := authServer.ValidateToken(ctx, req)

		assert.NoError(t, err) // gRPC метод возвращает nil ошибку даже при внутренней ошибке
		assert.NotNil(t, resp)
		assert.False(t, resp.Valid)
		assert.Equal(t, "token is invalid", resp.ErrorMessage)

		mockAuthService.AssertExpectations(t)
	})
}
