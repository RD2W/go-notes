package repository

import (
	"time"
)

// TokenRepository интерфейс для хранилища токенов
type TokenRepository interface {
	AddToBlacklist(tokenID string, expiresAt time.Time) error
	IsBlacklisted(tokenID string) (bool, error)
}
