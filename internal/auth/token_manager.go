package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/domain/repository"
	"github.com/rd2w/go-notes/internal/repository/redis"
)

// TokenClaims структура для хранения данных в JWT токене
type TokenClaims struct {
	Username string `json:"username"`
	TokenID  string `json:"token_id"`
	jwt.RegisteredClaims
}

// TokenManager структура для управления токенами
type TokenManager struct {
	jwtSecret         []byte
	refreshSecret     []byte
	jwtExpiration     time.Duration
	refreshExpiration time.Duration
	store             repository.TokenRepository
}

// NewTokenManager создает новый менеджер токенов
func NewTokenManager(config *config.Config) *TokenManager {
	accessTokenDuration, err := time.ParseDuration(config.JWT.AccessTokenTTL)
	if err != nil {
		log.Printf("Ошибка парсинга access_token_ttl, используется значение по умолчанию 15m: %v", err)
		accessTokenDuration = 15 * time.Minute
	}

	refreshExpiration := 7 * 24 * time.Hour // Значение по умолчанию 7 дней
	// Используем refresh_token_ttl из JWT конфигурации
	if config.JWT.RefreshTokenTTL != "" {
		if parsedRefreshDuration, parseErr := time.ParseDuration(config.JWT.RefreshTokenTTL); parseErr == nil {
			refreshExpiration = parsedRefreshDuration
		}
	}

	tokenStore, err := redis.NewRedisTokenRepository(config)
	if err != nil {
		log.Fatalf("Ошибка создания Redis хранилища токенов: %v", err)
	}

	return &TokenManager{
		jwtSecret:         []byte(config.JWT.SecretKey),
		refreshSecret:     []byte(config.Refresh.SecretKey),
		jwtExpiration:     accessTokenDuration,
		refreshExpiration: refreshExpiration,
		store:             tokenStore,
	}
}

// GenerateTokens генерирует пару access и refresh токенов
func (tm *TokenManager) GenerateTokens(username string) (string, string, error) {
	// Генерируем уникальный ID для токена, который будет использоваться для отслеживания обоих токенов (access и refresh)
	tokenID, err := tm.generateTokenID()
	if err != nil {
		return "", "", fmt.Errorf("ошибка генерации ID токена: %w", err)
	}

	// Время истечения токенов
	accessExpiresAt := time.Now().Add(tm.jwtExpiration)
	refreshExpiresAt := time.Now().Add(tm.refreshExpiration)

	// Генерируем access токен
	accessClaims := &TokenClaims{
		Username: username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
		},
	}

	accessSignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	token1, err := accessSignedToken.SignedString(tm.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("ошибка подписания access токена: %w", err)
	}

	// Генерируем refresh токен
	refreshClaims := &TokenClaims{
		Username: username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		},
	}

	refreshSignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	token2, err := refreshSignedToken.SignedString(tm.refreshSecret)
	if err != nil {
		return "", "", fmt.Errorf("ошибка подписания refresh токена: %w", err)
	}

	return token1, token2, nil
}

// RefreshTokens обновляет пару токенов по refresh токену
func (tm *TokenManager) RefreshTokens(refreshToken string) (string, string, error) {
	// Сначала проверяем валидность refresh токена
	claims, err := tm.validateRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("refresh токен недействителен: %w", err)
	}

	// Проверяем, не находится ли токен в черном списке
	isBlacklisted, err := tm.store.IsBlacklisted(claims.TokenID)
	if err != nil || isBlacklisted {
		return "", "", errors.New("refresh токен не найден или был отозван")
	}

	// Генерируем новые токены
	newAccessToken, newRefreshToken, err := tm.GenerateTokens(claims.Username)
	if err != nil {
		return "", "", fmt.Errorf("ошибка генерации новых токенов: %w", err)
	}

	// Добавляем старый refresh токен в черный список
	err = tm.store.AddToBlacklist(claims.TokenID, time.Now().Add(tm.refreshExpiration))
	if err != nil {
		log.Printf("Ошибка добавления токена в черный список: %v", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// ValidateAccessToken проверяет валидность access токена
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return tm.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("access токен недействителен: %w", err)
	}

	// Проверяем, не находится ли токен в черном списке
	isBlacklisted, err := tm.store.IsBlacklisted(claims.TokenID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке токена в хранилище: %w", err)
	}
	if isBlacklisted {
		return nil, errors.New("access токен был отозван или недействителен")
	}

	return claims, nil
}

// ValidateRefreshToken проверяет валидность refresh токена
func (tm *TokenManager) validateRefreshToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return tm.refreshSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("refresh токен недействителен: %w", err)
	}

	return claims, nil
}

// Logout отзывает токены пользователя
func (tm *TokenManager) Logout(refreshToken string) error {
	// Сначала проверяем валидность refresh токена
	claims, err := tm.validateRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("refresh токен недействителен: %w", err)
	}

	// Добавляем токен в черный список
	err = tm.store.AddToBlacklist(claims.TokenID, time.Now().Add(tm.refreshExpiration))
	if err != nil {
		return fmt.Errorf("ошибка добавления токена в черный список: %w", err)
	}

	return nil
}

// GetJWTExpiration возвращает время жизни JWT токена
func (tm *TokenManager) GetJWTExpiration() time.Duration {
	return tm.jwtExpiration
}

// GetJWTExpirationSeconds возвращает время жизни JWT токена в секундах
func (tm *TokenManager) GetJWTExpirationSeconds() int64 {
	return int64(tm.jwtExpiration.Seconds())
}

// generateTokenID генерирует уникальный ID для токена
func (tm *TokenManager) generateTokenID() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
