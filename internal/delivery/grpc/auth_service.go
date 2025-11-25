package grpc

import (
	"context"
	"log"

	"github.com/rd2w/go-notes/internal/domain/service"
	authPb "github.com/rd2w/go-notes/pkg/proto/auth"
)

// AuthServiceServer реализует gRPC-сервер для сервиса аутентификации
type AuthServiceServer struct {
	authPb.UnimplementedAuthServiceServer
	authService service.AuthService
}

// NewAuthServiceServer создает новый экземпляр gRPC-сервера для аутентификации
func NewAuthServiceServer(authService service.AuthService) *AuthServiceServer {
	return &AuthServiceServer{
		authService: authService,
	}
}

// Login реализует метод аутентификации пользователя и получения токенов
func (s *AuthServiceServer) Login(ctx context.Context, req *authPb.LoginRequest) (*authPb.LoginResponse, error) {
	accessToken, refreshToken, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		log.Printf("Error during login: %v", err)
		return nil, err
	}

	return &authPb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

// Logout реализует метод выхода пользователя и отзыва токена
func (s *AuthServiceServer) Logout(ctx context.Context, req *authPb.LogoutRequest) (*authPb.LogoutResponse, error) {
	err := s.authService.Logout(req.RefreshToken)
	if err != nil {
		log.Printf("Error during logout: %v", err)
		return &authPb.LogoutResponse{
			Success: false,
			Message: "failed to logout",
		}, nil
	}

	return &authPb.LogoutResponse{
		Success: true,
		Message: "successful logout",
	}, nil
}

// Refresh реализует метод обновления токена
func (s *AuthServiceServer) Refresh(ctx context.Context, req *authPb.RefreshRequest) (*authPb.RefreshResponse, error) {
	newAccessToken, newRefreshToken, err := s.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		log.Printf("Error during token refresh: %v", err)
		return nil, err
	}

	return &authPb.RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken реализует метод проверки валидности токена
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *authPb.ValidateTokenRequest) (*authPb.ValidateTokenResponse, error) {
	username, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		return &authPb.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "token is invalid",
		}, nil
	}

	return &authPb.ValidateTokenResponse{
		Valid:    true,
		Username: username,
	}, nil
}
