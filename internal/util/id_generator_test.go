package util

import (
	"regexp"
	"testing"
)

// TestGenerateID тестирует функцию GenerateID
func TestGenerateID(t *testing.T) {
	// Тестируем, что генерируемый ID не пустой
	id := GenerateID()
	if id == "" {
		t.Error("Generated ID should not be empty")
	}

	// Проверяем, что ID имеет правильный формат UUID
	uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
	if !uuidRegex.MatchString(id) {
		t.Errorf("Generated ID '%s' does not match UUID format", id)
	}
}

// TestGenerateIDUniqueness тестирует уникальность генерируемых ID
func TestGenerateIDUniqueness(t *testing.T) {
	ids := make(map[string]bool)
	count := 1000 // Количество генераций для тестирования уникальности

	for i := 0; i < count; i++ {
		id := GenerateID()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}

	// Проверяем, что мы получили ожидаемое количество уникальных ID
	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, but got %d", count, len(ids))
	}
}
