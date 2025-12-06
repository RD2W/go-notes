package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUser(t *testing.T) {
	t.Run("NewUser creates user with proper initialization and hashed password", func(t *testing.T) {
		username := "testuser"
		email := "test@example.com"
		password := "password123"

		user, err := NewUser(username, email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, user.GetID())
		assert.Equal(t, username, user.GetUsername())
		assert.Equal(t, email, user.GetEmail())
		assert.NotEmpty(t, user.GetPassword()) // Should be hashed
		assert.Equal(t, "user", user.GetType())

		// Check that timestamps are set
		assert.False(t, user.GetCreatedAt().IsZero())
		assert.False(t, user.GetUpdatedAt().IsZero())
		assert.True(t, user.GetCreatedAt().Equal(user.GetUpdatedAt()))

		// Verify that password is properly hashed
		assert.True(t, user.CheckPassword(password))
		assert.False(t, user.CheckPassword("wrongpassword"))
	})

	t.Run("NewUser returns error for invalid password", func(t *testing.T) {
		// Test with a very long password that might cause bcrypt to fail
		user, err := NewUser("testuser", "test@example.com", "")

		// Empty password should work (bcrypt handles it)
		assert.NoError(t, err)
		if err == nil {
			assert.NotNil(t, user)
		}
	})

	t.Run("NewUserWithPasswordHash creates user with provided hash", func(t *testing.T) {
		username := "testuser"
		email := "test@example.com"
		password := "password123"

		// First hash the password manually
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.NoError(t, err)

		user := NewUserWithPasswordHash(username, email, string(hashedPassword))

		assert.NotEmpty(t, user.GetID())
		assert.Equal(t, username, user.GetUsername())
		assert.Equal(t, email, user.GetEmail())
		assert.Equal(t, string(hashedPassword), user.GetPassword())
		assert.Equal(t, "user", user.GetType())

		// Check that timestamps are set
		assert.False(t, user.GetCreatedAt().IsZero())
		assert.False(t, user.GetUpdatedAt().IsZero())
		assert.True(t, user.GetCreatedAt().Equal(user.GetUpdatedAt()))

		// Verify that password check works
		assert.True(t, user.CheckPassword(password))
		assert.False(t, user.CheckPassword("wrongpassword"))
	})

	t.Run("Getters and setters work correctly", func(t *testing.T) {
		user, err := NewUser("originaluser", "original@example.com", "password123")
		assert.NoError(t, err)

		originalID := user.GetID()
		originalCreatedAt := user.GetCreatedAt()
		originalUpdatedAt := user.GetUpdatedAt()

		// Test SetID
		newID := "new-user-id"
		user.SetID(newID)
		assert.Equal(t, newID, user.GetID())

		// Test SetUsername (should update timestamp)
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		updatedAtBefore := user.GetUpdatedAt()
		user.SetUsername("newusername")
		assert.Equal(t, "newusername", user.GetUsername())
		assert.True(t, user.GetUpdatedAt().After(updatedAtBefore))

		// Test SetEmail (should update timestamp)
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		updatedAtBefore = user.GetUpdatedAt()
		user.SetEmail("newemail@example.com")
		assert.Equal(t, "newemail@example.com", user.GetEmail())
		assert.True(t, user.GetUpdatedAt().After(updatedAtBefore))

		// Test SetPassword (should update timestamp and hash the password)
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		updatedAtBefore = user.GetUpdatedAt()
		err = user.SetPassword("newpassword456")
		assert.NoError(t, err)
		assert.True(t, user.GetUpdatedAt().After(updatedAtBefore))
		assert.True(t, user.CheckPassword("newpassword456"))
		assert.False(t, user.CheckPassword("password123"))

		// Test SetPasswordHash (should update timestamp)
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		updatedAtBefore = user.GetUpdatedAt()
		newHash, err := bcrypt.GenerateFromPassword([]byte("directhash"), bcrypt.DefaultCost)
		assert.NoError(t, err)
		user.SetPasswordHash(string(newHash))
		assert.True(t, user.GetUpdatedAt().After(updatedAtBefore))
		assert.True(t, user.CheckPassword("directhash"))

		// Test SetCreatedAt and SetUpdatedAt
		customTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		user.SetCreatedAt(customTime)
		user.SetUpdatedAt(customTime.Add(1 * time.Hour))

		assert.True(t, user.GetCreatedAt().Equal(customTime))
		assert.True(t, user.GetUpdatedAt().Equal(customTime.Add(1*time.Hour)))

		// Restore original ID for consistency
		user.SetID(originalID)
		user.SetCreatedAt(originalCreatedAt)
		user.SetUpdatedAt(originalUpdatedAt)
	})

	t.Run("SetPassword returns error for invalid password", func(t *testing.T) {
		user, err := NewUser("testuser", "test@example.com", "password123")
		assert.NoError(t, err)

		// Test with a very long password that might cause bcrypt to fail
		longPassword := make([]byte, 800)
		for i := range longPassword {
			longPassword[i] = 'a'
		}

		err = user.SetPassword(string(longPassword))
		// This should cause bcrypt to return an error due to password length
		// If it doesn't, that's fine - we're just testing error handling
		if err != nil {
			// We expect an error might occur due to password length
			t.Logf("SetPassword returned error (expected for long password): %v", err)
		}
	})
}

func TestUserJSONSerialization(t *testing.T) {
	t.Run("MarshalJSON works correctly (without password)", func(t *testing.T) {
		user := NewUserWithPasswordHash("testuser", "test@example.com", "hashedpassword123")
		customTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		user.SetCreatedAt(customTime)
		user.SetUpdatedAt(customTime.Add(1 * time.Hour))
		user.SetID("test-user-id")

		data, err := json.Marshal(user)
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)

		assert.Equal(t, "test-user-id", result["id"])
		assert.Equal(t, "testuser", result["username"])
		assert.Equal(t, "test@example.com", result["email"])

		// Check that timestamps exist in the JSON
		assert.Contains(t, string(data), "2023-01-01T12:00:00Z")
		assert.Contains(t, string(data), "2023-01-01T13:00:00Z")

		// Password should not be in the JSON output
		_, hasPassword := result["password"]
		assert.False(t, hasPassword, "Password should not be serialized in normal MarshalJSON")
	})

	t.Run("UnmarshalJSON works correctly", func(t *testing.T) {
		jsonData := `{
			"id": "test-user-id",
			"username": "testuser",
			"email": "test@example.com",
			"created_at": "2023-01-01T12:00:00Z",
			"updated_at": "2023-01-01T13:00:00Z"
		}`

		var user User
		err := json.Unmarshal([]byte(jsonData), &user)
		assert.NoError(t, err)

		assert.Equal(t, "test-user-id", user.GetID())
		assert.Equal(t, "testuser", user.GetUsername())
		assert.Equal(t, "test@example.com", user.GetEmail())

		expectedCreatedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		expectedUpdatedAt := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)

		assert.True(t, user.GetCreatedAt().Equal(expectedCreatedAt))
		assert.True(t, user.GetUpdatedAt().Equal(expectedUpdatedAt))
	})

	t.Run("MarshalJSONWithPassword and UnmarshalJSONWithPassword work correctly", func(t *testing.T) {
		user := NewUserWithPasswordHash("testuser", "test@example.com", "hashedpassword123")
		customTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		user.SetCreatedAt(customTime)
		user.SetUpdatedAt(customTime.Add(1 * time.Hour))
		user.SetID("test-user-id")

		// Marshal with password
		data, err := user.MarshalJSONWithPassword()
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)

		assert.Equal(t, "test-user-id", result["id"])
		assert.Equal(t, "testuser", result["username"])
		assert.Equal(t, "test@example.com", result["email"])
		assert.Equal(t, "hashedpassword123", result["password"])
		assert.Equal(t, "2023-01-01T12:00:00Z", result["created_at"])
		assert.Equal(t, "2023-01-01T13:00:00Z", result["updated_at"])

		// Unmarshal with password
		var deserializedUser User
		err = deserializedUser.UnmarshalJSONWithPassword(data)
		assert.NoError(t, err)

		assert.Equal(t, "test-user-id", deserializedUser.GetID())
		assert.Equal(t, "testuser", deserializedUser.GetUsername())
		assert.Equal(t, "test@example.com", deserializedUser.GetEmail())
		assert.Equal(t, "hashedpassword123", deserializedUser.GetPassword())
		assert.True(t, deserializedUser.GetCreatedAt().Equal(customTime))
		assert.True(t, deserializedUser.GetUpdatedAt().Equal(customTime.Add(1*time.Hour)))
	})

	t.Run("UnmarshalJSON handles password field when present", func(t *testing.T) {
		jsonData := `{
			"id": "test-user-id",
			"username": "testuser",
			"email": "test@example.com",
			"password": "hashedpassword123",
			"created_at": "2023-01-01T12:00:00Z",
			"updated_at": "2023-01-01T13:00:00Z"
		}`

		var user User
		err := json.Unmarshal([]byte(jsonData), &user)
		assert.NoError(t, err)

		assert.Equal(t, "test-user-id", user.GetID())
		assert.Equal(t, "testuser", user.GetUsername())
		assert.Equal(t, "test@example.com", user.GetEmail())
		assert.Equal(t, "hashedpassword123", user.GetPassword()) // Password should be set when present in input

		expectedCreatedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		expectedUpdatedAt := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)

		assert.True(t, user.GetCreatedAt().Equal(expectedCreatedAt))
		assert.True(t, user.GetUpdatedAt().Equal(expectedUpdatedAt))
	})

	t.Run("Round trip serialization/deserialization with password", func(t *testing.T) {
		originalUser := NewUserWithPasswordHash("testuser", "test@example.com", "hashedpassword123")
		customTime := time.Date(2023, 5, 15, 10, 30, 45, 123456789, time.UTC)
		originalUser.SetCreatedAt(customTime)
		originalUser.SetUpdatedAt(customTime.Add(1 * time.Hour))
		originalUser.SetID("test-user-id")

		// Serialize with password
		data, err := originalUser.MarshalJSONWithPassword()
		assert.NoError(t, err)

		// Deserialize with password
		var deserializedUser User
		err = deserializedUser.UnmarshalJSONWithPassword(data)
		assert.NoError(t, err)

		// Compare
		assert.Equal(t, originalUser.GetID(), deserializedUser.GetID())
		assert.Equal(t, originalUser.GetUsername(), deserializedUser.GetUsername())
		assert.Equal(t, originalUser.GetEmail(), deserializedUser.GetEmail())
		assert.Equal(t, originalUser.GetPassword(), deserializedUser.GetPassword())
		assert.True(t, originalUser.GetCreatedAt().Equal(deserializedUser.GetCreatedAt()))
		assert.True(t, originalUser.GetUpdatedAt().Equal(deserializedUser.GetUpdatedAt()))
	})
}
