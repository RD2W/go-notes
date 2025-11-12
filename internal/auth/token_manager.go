package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rd2w/go-notes/internal/config"
)

// TokenClaims структура для хранения данных в JWT токене
type TokenClaims struct {
	Username string `json:"username"`
	TokenID  string `json:"token_id"`
	jwt.RegisteredClaims
}

// TokenStore интерфейс для хранения токенов
type TokenStore interface {
	Save(tokenID, username string, expiresAt time.Time) error
	Validate(tokenID, username string) (bool, error)
	Revoke(tokenID, username string) error
	Cleanup() error
}

// InMemoryTokenStore реализация хранилища токенов в памяти
type InMemoryTokenStore struct {
	tokens map[string]TokenData
	mutex  sync.RWMutex
}

// TokenData структура для хранения информации о токене
type TokenData struct {
	Username  string
	ExpiresAt time.Time
	Revoked   bool
}

// NewInMemoryTokenStore создает новое хранилище токенов в памяти
func NewInMemoryTokenStore() *InMemoryTokenStore {
	store := &InMemoryTokenStore{
		tokens: make(map[string]TokenData),
	}

	// Запускаем горутину для очистки просроченных токенов
	go store.startCleanupTicker()

	return store
}

// Save сохраняет токен в хранилище
func (s *InMemoryTokenStore) Save(tokenID, username string, expiresAt time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tokens[tokenID] = TokenData{
		Username:  username,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}

	return nil
}

// Validate проверяет валидность токена
func (s *InMemoryTokenStore) Validate(tokenID, username string) (bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tokenData, exists := s.tokens[tokenID]
	if !exists {
		return false, nil
	}

	if tokenData.Revoked {
		return false, nil
	}

	if tokenData.Username != username {
		return false, nil
	}

	if time.Now().After(tokenData.ExpiresAt) {
		return false, nil
	}

	return true, nil
}

// Revoke отменяет (отзывает) токен
func (s *InMemoryTokenStore) Revoke(tokenID, username string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	tokenData, exists := s.tokens[tokenID]
	if !exists {
		return errors.New("token not found")
	}

	if tokenData.Username != username {
		return errors.New("username does not match")
	}

	tokenData.Revoked = true
	s.tokens[tokenID] = tokenData

	return nil
}

// Cleanup удаляет просроченные и отозванные токены
func (s *InMemoryTokenStore) Cleanup() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for tokenID, tokenData := range s.tokens {
		if now.After(tokenData.ExpiresAt) || tokenData.Revoked {
			delete(s.tokens, tokenID)
		}
	}

	return nil
}

// startCleanupTicker запускает тикер для периодической очистки просроченных токенов
func (s *InMemoryTokenStore) startCleanupTicker() {
	ticker := time.NewTicker(1 * time.Hour) // Очищать раз в час
	defer ticker.Stop()

	for range ticker.C {
		if err := s.Cleanup(); err != nil {
			log.Printf("Error cleaning up tokens: %v", err)
		}
	}
}

// TokenManager структура для управления токенами
type TokenManager struct {
	jwtSecret         []byte
	refreshSecret     []byte
	jwtExpiration     time.Duration
	refreshExpiration time.Duration
	store             TokenStore
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

	return &TokenManager{
		jwtSecret:         []byte(config.JWT.SecretKey),
		refreshSecret:     []byte(config.Refresh.SecretKey),
		jwtExpiration:     accessTokenDuration,
		refreshExpiration: refreshExpiration,
		store:             NewInMemoryTokenStore(),
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

	// Сохраняем refresh токен в хранилище (используется для отслеживания и отзыва)
	err = tm.store.Save(tokenID, username, refreshExpiresAt)
	if err != nil {
		return "", "", fmt.Errorf("ошибка сохранения refresh токена: %w", err)
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

	// Проверяем, не был ли токен отозван
	isValid, err := tm.store.Validate(claims.TokenID, claims.Username)
	if err != nil || !isValid {
		return "", "", errors.New("refresh токен не найден или был отозван")
	}

	// Генерируем новые токены
	newAccessToken, newRefreshToken, err := tm.GenerateTokens(claims.Username)
	if err != nil {
		return "", "", fmt.Errorf("ошибка генерации новых токенов: %w", err)
	}

	// Отзываем старый refresh токен
	err = tm.store.Revoke(claims.TokenID, claims.Username)
	if err != nil {
		log.Printf("Ошибка отзыва старого refresh токена: %v", err)
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

	// Проверяем, не был ли токен отозван
	isValid, err := tm.store.Validate(claims.TokenID, claims.Username)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке токена в хранилище: %w", err)
	}
	if !isValid {
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

	// Отзываем токен (refresh токен, но также может затронуть и соответствующий access токен с тем же TokenID)
	err = tm.store.Revoke(claims.TokenID, claims.Username)
	if err != nil {
		return fmt.Errorf("ошибка отзыва токена: %w", err)
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
