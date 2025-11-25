package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeFields(t *testing.T) {
	t.Run("initializeTimestamps sets both timestamps to current time", func(t *testing.T) {
		timeFields := &TimeFields{}

		// Capture time before and after initialization
		before := time.Now()
		timeFields.initializeTimestamps()
		after := time.Now()

		// Check that both timestamps are set
		createdAt := timeFields.GetCreatedAt()
		updatedAt := timeFields.GetUpdatedAt()

		assert.True(t, createdAt.After(before) || createdAt.Equal(before))
		assert.True(t, createdAt.Before(after) || createdAt.Equal(after))
		assert.True(t, updatedAt.After(before) || updatedAt.Equal(before))
		assert.True(t, updatedAt.Before(after) || updatedAt.Equal(after))
		assert.True(t, createdAt.Equal(updatedAt))
	})

	t.Run("updateTimestamp updates only updatedAt", func(t *testing.T) {
		timeFields := &TimeFields{}
		timeFields.initializeTimestamps()

		originalCreatedAt := timeFields.GetCreatedAt()
		originalUpdatedAt := timeFields.GetUpdatedAt()

		// Wait a moment before updating
		time.Sleep(1 * time.Millisecond)
		timeFields.updateTimestamp()

		newCreatedAt := timeFields.GetCreatedAt()
		newUpdatedAt := timeFields.GetUpdatedAt()

		// CreatedAt should remain unchanged
		assert.True(t, newCreatedAt.Equal(originalCreatedAt))
		// UpdatedAt should be newer
		assert.True(t, newUpdatedAt.After(originalUpdatedAt))
	})

	t.Run("TimeFields can be initialized and accessed", func(t *testing.T) {
		// Note: TimeFields doesn't have SetCreatedAt/SetUpdatedAt methods
		// This test verifies that the internal fields can be accessed via exported methods
		timeFields := &TimeFields{}
		timeFields.initializeTimestamps()

		createdAt := timeFields.GetCreatedAt()
		updatedAt := timeFields.GetUpdatedAt()

		assert.True(t, !createdAt.IsZero())
		assert.True(t, !updatedAt.IsZero())
	})
}

func TestTimeFieldsJSONSerialization(t *testing.T) {
	t.Run("MarshalJSON works correctly", func(t *testing.T) {
		timeFields := &TimeFields{}
		// Initialize timestamps and verify JSON serialization works
		timeFields.initializeTimestamps()

		data, err := json.Marshal(timeFields)
		assert.NoError(t, err)

		// Verify that the JSON contains the expected timestamp fields
		assert.Contains(t, string(data), "created_at")
		assert.Contains(t, string(data), "updated_at")
	})

	t.Run("UnmarshalJSON works correctly", func(t *testing.T) {
		jsonData := `{"created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T13:00:00Z"}`

		var timeFields TimeFields
		err := json.Unmarshal([]byte(jsonData), &timeFields)
		assert.NoError(t, err)

		expectedCreatedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		expectedUpdatedAt := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)

		assert.True(t, timeFields.GetCreatedAt().Equal(expectedCreatedAt))
		assert.True(t, timeFields.GetUpdatedAt().Equal(expectedUpdatedAt))
	})

	t.Run("Round trip serialization/deserialization", func(t *testing.T) {
		// Start with JSON data
		jsonData := `{"created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T13:00:00Z"}`

		// Unmarshal to TimeFields
		var timeFields TimeFields
		err := json.Unmarshal([]byte(jsonData), &timeFields)
		assert.NoError(t, err)

		originalCreatedAt := timeFields.GetCreatedAt()
		originalUpdatedAt := timeFields.GetUpdatedAt()

		// Verify the values were properly set
		expectedCreatedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		expectedUpdatedAt := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)
		assert.True(t, originalCreatedAt.Equal(expectedCreatedAt))
		assert.True(t, originalUpdatedAt.Equal(expectedUpdatedAt))

		// Serialize back to JSON - use pointer to ensure MarshalJSON is called
		data, err := json.Marshal(&timeFields)
		assert.NoError(t, err)

		t.Logf("Marshaled JSON: %s", string(data))

		// Unmarshal again
		var timeFields2 TimeFields
		err = json.Unmarshal(data, &timeFields2)
		assert.NoError(t, err)

		t.Logf("TimeFields2 CreatedAt: %v, UpdatedAt: %v", timeFields2.GetCreatedAt(), timeFields2.GetUpdatedAt())

		// Compare - allow for small time differences due to serialization precision
		assert.WithinDuration(t, originalCreatedAt, timeFields2.GetCreatedAt(), time.Nanosecond)
		assert.WithinDuration(t, originalUpdatedAt, timeFields2.GetUpdatedAt(), time.Nanosecond)
	})
}

func TestDebugTimeFields(t *testing.T) {
	jsonData := `{"created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T13:00:00Z"}`

	var timeFields TimeFields
	err := json.Unmarshal([]byte(jsonData), &timeFields)
	assert.NoError(t, err)

	createdAt := timeFields.GetCreatedAt()
	updatedAt := timeFields.GetUpdatedAt()

	t.Logf("CreatedAt: %v", createdAt)
	t.Logf("UpdatedAt: %v", updatedAt)
	t.Logf("Are they zero? Created: %t, Updated: %t", createdAt.IsZero(), updatedAt.IsZero())

	// Test that the JSON unmarshaling actually calls the UnmarshalJSON method
	expectedCreatedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	expectedUpdatedAt := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)

	assert.True(t, createdAt.Equal(expectedCreatedAt), "CreatedAt should match expected")
	assert.True(t, updatedAt.Equal(expectedUpdatedAt), "UpdatedAt should match expected")
}
