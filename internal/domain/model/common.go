package model

import (
	"encoding/json"
	"time"
)

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

// JSONTimeFields вспомогательная структура для JSON сериализации
type JSONTimeFields struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MarshalJSON реализует интерфейс json.Marshaler
func (t *TimeFields) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONTimeFields{
		CreatedAt: t.createdAt,
		UpdatedAt: t.updatedAt,
	})
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler
func (t *TimeFields) UnmarshalJSON(data []byte) error {
	var jsonTimeFields JSONTimeFields
	if err := json.Unmarshal(data, &jsonTimeFields); err != nil {
		return err
	}

	t.createdAt = jsonTimeFields.CreatedAt
	t.updatedAt = jsonTimeFields.UpdatedAt

	return nil
}
