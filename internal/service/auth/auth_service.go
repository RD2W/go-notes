package auth

import (
	"errors"

	"github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/domain/service"
)

// authService реализация бизнес-логики для аутентификации
type authService struct {
	tokenManager *auth.TokenManager
	userService  service.UserService
}

// NewAuthService создает новый экземпляр сервиса аутентификации
func NewAuthService(tokenManager *auth.TokenManager, userService service.UserService) service.AuthService {
	return &authService{
		tokenManager: tokenManager,
		userService:  userService,
	}
}

// Login реализует аутентификацию пользователя
func (s *authService) Login(username, password string) (string, string, error) {
	user, err := s.userService.GetUserByUsername(username)
	if err != nil {
		return "", "", errors.New("пользователь не найден")
	}

	if !user.CheckPassword(password) {
		return "", "", errors.New("неверный пароль")
	}

	accessToken, refreshToken, err := s.tokenManager.GenerateTokens(user.GetUsername())
	if err != nil {
		return "", "", errors.New("ошибка генерации токенов")
	}

	return accessToken, refreshToken, nil
}

// Logout реализует выход пользователя
func (s *authService) Logout(refreshToken string) error {
	err := s.tokenManager.Logout(refreshToken)
	if err != nil {
		return errors.New("ошибка при выходе")
	}
	return nil
}

// RefreshTokens обновляет токены
func (s *authService) RefreshTokens(refreshToken string) (string, string, error) {
	newAccessToken, newRefreshToken, err := s.tokenManager.RefreshTokens(refreshToken)
	if err != nil {
		return "", "", errors.New("ошибка обновления токенов")
	}
	return newAccessToken, newRefreshToken, nil
}

// ValidateToken проверяет валидность токена
func (s *authService) ValidateToken(token string) (string, error) {
	claims, err := s.tokenManager.ValidateAccessToken(token)
	if err != nil {
		return "", errors.New("токен недействителен")
	}
	return claims.Username, nil
}
