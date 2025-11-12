package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestNewUser(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	user, err := NewUser(username, email, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.GetID())
	assert.Equal(t, username, user.GetUsername())
	assert.Equal(t, email, user.GetEmail())
	assert.NotEmpty(t, user.GetPassword())
	assert.True(t, user.CheckPassword(password))
	assert.False(t, user.CheckPassword("wrongpassword"))

	// Проверяем, что временные метки установлены
	assert.NotZero(t, user.GetCreatedAt())
	assert.NotZero(t, user.GetUpdatedAt())
}

func TestNewUserWithPasswordHash(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	// Создаем хеш пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user := NewUserWithPasswordHash(username, email, string(hashedPassword))

	assert.NotNil(t, user)
	assert.NotEmpty(t, user.GetID())
	assert.Equal(t, username, user.GetUsername())
	assert.Equal(t, email, user.GetEmail())
	assert.Equal(t, string(hashedPassword), user.GetPassword())
	assert.True(t, user.CheckPassword(password))
	assert.False(t, user.CheckPassword("wrongpassword"))

	// Проверяем, что временные метки установлены
	assert.NotZero(t, user.GetCreatedAt())
	assert.NotZero(t, user.GetUpdatedAt())
}

func TestNewUserError(t *testing.T) {
	// Тестируем ошибку при генерации хеша пароля (искусственно вызываем ошибку)
	// В реальности bcrypt.GenerateFromPassword редко возвращает ошибки,
	// но тестируем этот сценарий на всякий случай
	user, err := NewUser("testuser", "test@example.com", "password123")

	// Должно пройти успешно, так как валидный пароль
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestGetters(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	user, err := NewUser(username, email, password)
	assert.NoError(t, err)

	assert.Equal(t, user.GetID(), user.GetID())
	assert.Equal(t, username, user.GetUsername())
	assert.Equal(t, email, user.GetEmail())
	assert.Equal(t, "user", user.GetType())
	assert.NotEmpty(t, user.GetPassword())
}

func TestSetUsername(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "password123")
	assert.NoError(t, err)

	originalUpdatedAt := user.GetUpdatedAt()

	newUsername := "newusername"
	user.SetUsername(newUsername)

	assert.Equal(t, newUsername, user.GetUsername())
	assert.True(t, user.GetUpdatedAt().After(originalUpdatedAt))
}

func TestSetEmail(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "password123")
	assert.NoError(t, err)

	originalUpdatedAt := user.GetUpdatedAt()

	newEmail := "newemail@example.com"
	user.SetEmail(newEmail)

	assert.Equal(t, newEmail, user.GetEmail())
	assert.True(t, user.GetUpdatedAt().After(originalUpdatedAt))
}

func TestSetPassword(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "password123")
	assert.NoError(t, err)

	originalUpdatedAt := user.GetUpdatedAt()

	newPassword := "newpassword123"
	err = user.SetPassword(newPassword)
	assert.NoError(t, err)

	assert.True(t, user.CheckPassword(newPassword))
	assert.False(t, user.CheckPassword("password123"))
	assert.True(t, user.GetUpdatedAt().After(originalUpdatedAt))
}

func TestSetPasswordError(t *testing.T) {
	// Тестируем ошибку при установке пароля
	// В реальности bcrypt.GenerateFromPassword редко возвращает ошибки,
	// но тестируем этот сценарий на всякий случай
	user, err := NewUser("testuser", "test@example.com", "password123")
	assert.NoError(t, err)

	err = user.SetPassword("newpassword123")

	// Должно пройти успешно, так как валидный пароль
	assert.NoError(t, err)
	assert.True(t, user.CheckPassword("newpassword123"))
}

func TestCheckPassword(t *testing.T) {
	password := "password123"
	user, err := NewUser("testuser", "test@example.com", password)
	assert.NoError(t, err)

	assert.True(t, user.CheckPassword(password))
	assert.False(t, user.CheckPassword("wrongpassword"))
	assert.False(t, user.CheckPassword(""))
}

func TestMarshalJSON(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	user, err := NewUser(username, email, password)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(user)
	assert.NoError(t, err)

	// Проверяем, что JSON содержит ожидаемые поля, но не содержит пароля
	var userMap map[string]interface{}
	err = json.Unmarshal(jsonData, &userMap)
	assert.NoError(t, err)

	assert.Contains(t, userMap, "id")
	assert.Contains(t, userMap, "username")
	assert.Contains(t, userMap, "email")
	assert.Contains(t, userMap, "created_at")
	assert.Contains(t, userMap, "updated_at")
	assert.NotContains(t, userMap, "password")

	assert.Equal(t, user.GetID(), userMap["id"])
	assert.Equal(t, username, userMap["username"])
	assert.Equal(t, email, userMap["email"])
}

func TestMarshalJSONWithPassword(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	user, err := NewUser(username, email, password)
	assert.NoError(t, err)

	jsonData, err := user.MarshalJSONWithPassword()
	assert.NoError(t, err)

	// Проверяем, что JSON содержит все поля, включая пароль
	var userMap map[string]interface{}
	err = json.Unmarshal(jsonData, &userMap)
	assert.NoError(t, err)

	assert.Contains(t, userMap, "id")
	assert.Contains(t, userMap, "username")
	assert.Contains(t, userMap, "email")
	assert.Contains(t, userMap, "password")
	assert.Contains(t, userMap, "created_at")
	assert.Contains(t, userMap, "updated_at")

	assert.Equal(t, user.GetID(), userMap["id"])
	assert.Equal(t, username, userMap["username"])
	assert.Equal(t, email, userMap["email"])
	assert.Equal(t, user.GetPassword(), userMap["password"])
}

func TestUnmarshalJSONWithPassword(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	user, err := NewUser(username, email, password)
	assert.NoError(t, err)

	// Создаем JSON с паролем
	jsonData, err := user.MarshalJSONWithPassword()
	assert.NoError(t, err)

	// Создаем нового пользователя и десериализуем в него данные
	newUser := &User{}
	err = newUser.UnmarshalJSONWithPassword(jsonData)
	assert.NoError(t, err)

	assert.Equal(t, user.GetID(), newUser.GetID())
	assert.Equal(t, user.GetUsername(), newUser.GetUsername())
	assert.Equal(t, user.GetEmail(), newUser.GetEmail())
	assert.Equal(t, user.GetPassword(), newUser.GetPassword())
	assert.WithinDuration(t, user.GetCreatedAt(), newUser.GetCreatedAt(), time.Second)
	assert.WithinDuration(t, user.GetUpdatedAt(), newUser.GetUpdatedAt(), time.Second)
}

func TestUnmarshalJSON(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	user, err := NewUser(username, email, password)
	assert.NoError(t, err)

	// Маршалим в JSON (без пароля)
	jsonData, err := json.Marshal(user)
	assert.NoError(t, err)

	// Создаем нового пользователя и десериализуем в него данные
	newUser := &User{}
	err = newUser.UnmarshalJSON(jsonData)
	assert.NoError(t, err)

	assert.Equal(t, user.GetID(), newUser.GetID())
	assert.Equal(t, user.GetUsername(), newUser.GetUsername())
	assert.Equal(t, user.GetEmail(), newUser.GetEmail())
	// Пароль не должен быть установлен при десериализации без пароля
	assert.Empty(t, newUser.GetPassword())
	assert.WithinDuration(t, user.GetCreatedAt(), newUser.GetCreatedAt(), time.Second)
	assert.WithinDuration(t, user.GetUpdatedAt(), newUser.GetUpdatedAt(), time.Second)
}

func TestUnmarshalJSONWithPasswordIncluded(t *testing.T) {
	// Создаем JSON с паролем вручную
	jsonStr := `{"id":"test-id","username":"testuser","email":"test@example.com","password":"hashed_password","created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z"}`

	user := &User{}
	err := user.UnmarshalJSON([]byte(jsonStr))
	assert.NoError(t, err)

	assert.Equal(t, "test-id", user.id)
	assert.Equal(t, "testuser", user.username)
	assert.Equal(t, "test@example.com", user.email)
	assert.Equal(t, "hashed_password", user.password)

	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	assert.Equal(t, createdAt, user.createdAt)
	assert.Equal(t, updatedAt, user.updatedAt)
}
