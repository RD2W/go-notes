package auth

import (
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	"github.com/stretchr/testify/assert"
)

func createTestTokenManager() *TokenManager {
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

	return NewTokenManager(testConfig)
}

func TestTokenManager_GenerateTokens(t *testing.T) {
	tm := createTestTokenManager()

	username := "testuser"

	accessToken, refreshToken, err := tm.GenerateTokens(username)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// Проверяем, что токены можно валидировать
	accessClaims, err := tm.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, username, accessClaims.Username)

	// Проверяем, что refresh токен также действителен
	refreshClaims, err := tm.validateRefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, username, refreshClaims.Username)

	// Проверяем, что у токенов одинаковый TokenID
	assert.Equal(t, accessClaims.TokenID, refreshClaims.TokenID)
}

func TestTokenManager_ValidateAccessToken_Valid(t *testing.T) {
	tm := createTestTokenManager()

	username := "testuser"
	accessToken, _, err := tm.GenerateTokens(username)
	assert.NoError(t, err)

	claims, err := tm.ValidateAccessToken(accessToken)

	assert.NoError(t, err)
	assert.Equal(t, username, claims.Username)
	assert.NotEmpty(t, claims.TokenID)
}

func TestTokenManager_ValidateAccessToken_Invalid(t *testing.T) {
	tm := createTestTokenManager()

	claims, err := tm.ValidateAccessToken("invalid_token")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "access токен недействителен")
}

func TestTokenManager_ValidateAccessToken_Revoked(t *testing.T) {
	tm := createTestTokenManager()

	username := "testuser"
	accessToken, refreshToken, err := tm.GenerateTokens(username)
	assert.NoError(t, err)

	// Отзываем токены через logout
	err = tm.Logout(refreshToken)
	assert.NoError(t, err)

	// Проверяем, что access токен больше не валиден
	claims, err := tm.ValidateAccessToken(accessToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "access токен был отозван или недействителен")
}

func TestTokenManager_ValidateAccessToken_Expired(t *testing.T) {
	// Для тестирования истекших токенов создадим специальный TokenManager с коротким сроком действия
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test_secret_key_for_testing",
			AccessTokenTTL:  "10ms", // Очень короткое время жизни для теста
			RefreshTokenTTL: "168h",
		},
		Refresh: config.RefreshConfig{
			SecretKey: "test_refresh_secret_key_for_testing",
		},
	}

	tm := NewTokenManager(testConfig)

	username := "testuser"
	accessToken, _, err := tm.GenerateTokens(username)
	assert.NoError(t, err)

	// Немного ждем, чтобы токен истек
	time.Sleep(10 * time.Millisecond)

	claims, err := tm.ValidateAccessToken(accessToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "access токен недействителен")
}

func TestTokenManager_RefreshTokens(t *testing.T) {
	tm := createTestTokenManager()

	username := "testuser"
	_, refreshToken, err := tm.GenerateTokens(username)
	assert.NoError(t, err)

	newAccessToken, newRefreshToken, err := tm.RefreshTokens(refreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)

	// Проверяем, что новые токены валидны
	newClaims, err := tm.ValidateAccessToken(newAccessToken)
	assert.NoError(t, err)
	assert.Equal(t, username, newClaims.Username)

	// Проверяем, что старый refresh токен больше не действителен для обновления
	// Обратите внимание, что сам JWT токен остается валидным по подписи, но в системе он отозван
	// Для проверки отозванных refresh токенов нужно использовать дополнительную логику
	// или попытаться обновить токены с помощью старого refresh токена (должно вернуть ошибку)

	// Проверяем, что старый refresh токен не может быть использован для обновления
	_, _, err = tm.RefreshTokens(refreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh токен не найден или был отозван")
}

func TestTokenManager_RefreshTokens_InvalidRefreshToken(t *testing.T) {
	tm := createTestTokenManager()

	newAccessToken, newRefreshToken, err := tm.RefreshTokens("invalid_refresh_token")

	assert.Error(t, err)
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
	assert.Contains(t, err.Error(), "refresh токен недействителен")
}

func TestTokenManager_RefreshTokens_RevokedRefreshToken(t *testing.T) {
	tm := createTestTokenManager()

	username := "testuser"
	_, refreshToken, err := tm.GenerateTokens(username)
	assert.NoError(t, err)

	// Отзываем refresh токен
	err = tm.Logout(refreshToken)
	assert.NoError(t, err)

	// Пытаемся обновить токены с помощью отозванного refresh токена
	newAccessToken, newRefreshToken, err := tm.RefreshTokens(refreshToken)

	assert.Error(t, err)
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
	assert.Contains(t, err.Error(), "refresh токен не найден или был отозван")
}

func TestTokenManager_Logout(t *testing.T) {
	tm := createTestTokenManager()

	username := "testuser"
	accessToken, refreshToken, err := tm.GenerateTokens(username)
	assert.NoError(t, err)

	// Проверяем, что токены валидны до logout
	_, err = tm.ValidateAccessToken(accessToken)
	assert.NoError(t, err)

	// Выполняем logout
	err = tm.Logout(refreshToken)
	assert.NoError(t, err)

	// Проверяем, что access токен больше не валиден
	_, err = tm.ValidateAccessToken(accessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access токен был отозван или недействителен")
}

func TestTokenManager_Logout_InvalidToken(t *testing.T) {
	tm := createTestTokenManager()

	err := tm.Logout("invalid_token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh токен недействителен")
}

func TestInMemoryTokenStore_AddToBlacklistAndIsBlacklisted(t *testing.T) {
	store := NewInMemoryTokenStore()

	tokenID := "test-token-id"
	expiresAt := time.Now().Add(1 * time.Hour)

	// Добавляем токен в черный список
	err := store.AddToBlacklist(tokenID, expiresAt)
	assert.NoError(t, err)

	// Проверяем, что токен в черном списке
	isBlacklisted, err := store.IsBlacklisted(tokenID)
	assert.NoError(t, err)
	assert.True(t, isBlacklisted)

	// Проверяем несуществующий токен
	isBlacklisted, err = store.IsBlacklisted("other-token-id")
	assert.NoError(t, err)
	assert.False(t, isBlacklisted)
}

func TestInMemoryTokenStore_AddToBlacklist(t *testing.T) {
	store := NewInMemoryTokenStore()

	tokenID := "test-token-id"
	expiresAt := time.Now().Add(1 * time.Hour)

	// Добавляем токен в черный список
	err := store.AddToBlacklist(tokenID, expiresAt)
	assert.NoError(t, err)

	// Проверяем, что токен в черном списке
	isBlacklisted, err := store.IsBlacklisted(tokenID)
	assert.NoError(t, err)
	assert.True(t, isBlacklisted)
}

func TestInMemoryTokenStore_Cleanup(t *testing.T) {
	store := NewInMemoryTokenStore()

	// Добавляем просроченный токен в черный список
	expiredTokenID := "expired-token-id"
	expiredAt := time.Now().Add(-1 * time.Hour) // Токен просрочен

	err := store.AddToBlacklist(expiredTokenID, expiredAt)
	assert.NoError(t, err)

	// Добавляем валидный токен в черный список
	validTokenID := "valid-token-id"
	validAt := time.Now().Add(1 * time.Hour) // Токен валиден

	err = store.AddToBlacklist(validTokenID, validAt)
	assert.NoError(t, err)

	// Выполняем очистку
	err = store.Cleanup()
	assert.NoError(t, err)

	// Проверяем, что просроченный токен удален из черного списка
	isBlacklisted, err := store.IsBlacklisted(expiredTokenID)
	assert.NoError(t, err)
	assert.False(t, isBlacklisted)

	// Проверяем, что валидный токен остался в черном списке
	isBlacklisted, err = store.IsBlacklisted(validTokenID)
	assert.NoError(t, err)
	assert.True(t, isBlacklisted)
}

func TestInMemoryTokenStore_IsBlacklistedNonExistentToken(t *testing.T) {
	store := NewInMemoryTokenStore()

	isBlacklisted, err := store.IsBlacklisted("non-existent-token")

	assert.NoError(t, err)
	assert.False(t, isBlacklisted)
}

func TestTokenManager_GetJWTExpiration(t *testing.T) {
	expectedDuration := 2 * time.Hour
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test_secret_key_for_testing",
			AccessTokenTTL:  "2h", // 2 часа
			RefreshTokenTTL: "168h",
		},
		Refresh: config.RefreshConfig{
			SecretKey: "test_refresh_secret_key_for_testing",
		},
	}

	tm := NewTokenManager(testConfig)

	duration := tm.GetJWTExpiration()

	assert.Equal(t, expectedDuration, duration)

	seconds := tm.GetJWTExpirationSeconds()
	assert.Equal(t, int64(expectedDuration.Seconds()), seconds)
}
