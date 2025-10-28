package model

import "time"

// TimeFields содержит общие временные метки для сущностей
type TimeFields struct {
	createdAt time.Time
	updatedAt time.Time
}

// GetCreatedAt возвращает время создания
func (t *TimeFields) GetCreatedAt() time.Time {
	return t.createdAt
}

// GetUpdatedAt возвращает время обновления
func (t *TimeFields) GetUpdatedAt() time.Time {
	return t.updatedAt
}

// updateTimestamp обновляет время изменения
func (t *TimeFields) updateTimestamp() {
	t.updatedAt = time.Now()
}

// initializeTimestamps инициализирует временные метки
func (t *TimeFields) initializeTimestamps() {
	now := time.Now()
	t.createdAt = now
	t.updatedAt = now
}
