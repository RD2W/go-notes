package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNote(t *testing.T) {
	t.Run("NewNote creates note with proper initialization", func(t *testing.T) {
		title := "Test Title"
		content := "Test Content"
		userId := "user123"

		note := NewNote(title, content, userId)

		assert.NotEmpty(t, note.GetID())
		assert.Equal(t, title, note.GetTitle())
		assert.Equal(t, content, note.GetContent())
		assert.Equal(t, userId, note.GetUserID())
		assert.Equal(t, "note", note.GetType())

		// Check that timestamps are set
		assert.False(t, note.GetCreatedAt().IsZero())
		assert.False(t, note.GetUpdatedAt().IsZero())
		assert.True(t, note.GetCreatedAt().Equal(note.GetUpdatedAt()))
	})

	t.Run("Getters and setters work correctly", func(t *testing.T) {
		note := NewNote("Original Title", "Original Content", "user123")
		originalID := note.GetID()
		originalCreatedAt := note.GetCreatedAt()

		// Test SetID
		newID := "new-id-123"
		note.SetID(newID)
		assert.Equal(t, newID, note.GetID())

		// Test SetUserID
		newUserID := "new-user-456"
		note.SetUserID(newUserID)
		assert.Equal(t, newUserID, note.GetUserID())

		// Test SetTitle (should update timestamp)
		originalUpdatedAt := note.GetUpdatedAt()
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		note.SetTitle("New Title")
		assert.Equal(t, "New Title", note.GetTitle())
		assert.True(t, note.GetUpdatedAt().After(originalUpdatedAt))

		// Test SetContent (should update timestamp)
		updatedAtAfterSetTitle := note.GetUpdatedAt()
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		note.SetContent("New Content")
		assert.Equal(t, "New Content", note.GetContent())
		assert.True(t, note.GetUpdatedAt().After(updatedAtAfterSetTitle))

		// Restore original ID for consistency
		note.SetID(originalID)
		note.SetCreatedAt(originalCreatedAt)
	})

	t.Run("SetCreatedAt and SetUpdatedAt work correctly", func(t *testing.T) {
		note := NewNote("Title", "Content", "user123")
		customTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

		note.SetCreatedAt(customTime)
		note.SetUpdatedAt(customTime.Add(1 * time.Hour))

		assert.True(t, note.GetCreatedAt().Equal(customTime))
		assert.True(t, note.GetUpdatedAt().Equal(customTime.Add(1*time.Hour)))
	})
}

func TestNoteJSONSerialization(t *testing.T) {
	t.Run("MarshalJSON works correctly", func(t *testing.T) {
		note := NewNote("Test Title", "Test Content", "user123")
		customTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		note.SetCreatedAt(customTime)
		note.SetUpdatedAt(customTime.Add(1 * time.Hour))
		note.SetID("test-id-123")

		data, err := json.Marshal(note)
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)

		assert.Equal(t, "test-id-123", result["id"])
		assert.Equal(t, "Test Title", result["title"])
		assert.Equal(t, "Test Content", result["content"])
		assert.Equal(t, "user123", result["user_id"])
		assert.Equal(t, "2023-01-01T12:00:00Z", result["created_at"])
		assert.Equal(t, "2023-01-01T13:00:00Z", result["updated_at"])
	})

	t.Run("UnmarshalJSON works correctly", func(t *testing.T) {
		jsonData := `{
			"id": "test-id-123",
			"title": "Test Title",
			"content": "Test Content",
			"user_id": "user123",
			"created_at": "2023-01-01T12:00:00Z",
			"updated_at": "2023-01-01T13:00:00Z"
		}`

		var note Note
		err := json.Unmarshal([]byte(jsonData), &note)
		assert.NoError(t, err)

		assert.Equal(t, "test-id-123", note.GetID())
		assert.Equal(t, "Test Title", note.GetTitle())
		assert.Equal(t, "Test Content", note.GetContent())
		assert.Equal(t, "user123", note.GetUserID())

		expectedCreatedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		expectedUpdatedAt := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)

		assert.True(t, note.GetCreatedAt().Equal(expectedCreatedAt))
		assert.True(t, note.GetUpdatedAt().Equal(expectedUpdatedAt))
	})

	t.Run("Round trip serialization/deserialization", func(t *testing.T) {
		original := NewNote("Test Title", "Test Content", "user123")
		customTime := time.Date(2023, 5, 15, 10, 30, 45, 123456789, time.UTC)
		original.SetCreatedAt(customTime)
		original.SetUpdatedAt(customTime.Add(1 * time.Hour))
		original.SetID("test-id-123")
		original.SetUserID("user456")

		// Serialize
		data, err := json.Marshal(original)
		assert.NoError(t, err)

		// Deserialize
		var deserialized Note
		err = json.Unmarshal(data, &deserialized)
		assert.NoError(t, err)

		// Compare
		assert.Equal(t, original.GetID(), deserialized.GetID())
		assert.Equal(t, original.GetTitle(), deserialized.GetTitle())
		assert.Equal(t, original.GetContent(), deserialized.GetContent())
		assert.Equal(t, original.GetUserID(), deserialized.GetUserID())
		assert.True(t, original.GetCreatedAt().Equal(deserialized.GetCreatedAt()))
		assert.True(t, original.GetUpdatedAt().Equal(deserialized.GetUpdatedAt()))
	})
}
