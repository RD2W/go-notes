package util

import "github.com/google/uuid"

// GenerateID генерирует уникальный идентификатор
func GenerateID() string {
	return uuid.New().String()
}
