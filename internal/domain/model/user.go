package model

import (
	"encoding/json"
	"time"

	"github.com/rd2w/go-notes/internal/util"
	"golang.org/x/crypto/bcrypt"
)

// User представляет сущность пользователя
type User struct {
	TimeFields
	id       string
	username string
	email    string
	password string // хешированный пароль
}

// NewUser создает нового пользователя с инициализацией временных меток
func NewUser(username, email, password string) (*User, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		id:       util.GenerateID(),
		username: username,
		email:    email,
		password: string(hashedPassword),
	}
	user.initializeTimestamps()
	return user, nil
}

// NewUserWithPasswordHash создает нового пользователя с уже хешированным паролем (для загрузки из хранилища)
func NewUserWithPasswordHash(username, email, passwordHash string) *User {
	user := &User{
		id:       util.GenerateID(),
		username: username,
		email:    email,
		password: passwordHash,
	}
	user.initializeTimestamps()
	return user
}

// GetID возвращает идентификатор пользователя (реализация интерфейса Entity)
func (u *User) GetID() string {
	return u.id
}

// GetType возвращает тип сущности (реализация интерфейса Entity)
func (u *User) GetType() string {
	return "user"
}

// GetUsername возвращает имя пользователя
func (u *User) GetUsername() string {
	return u.username
}

// GetEmail возвращает email пользователя
func (u *User) GetEmail() string {
	return u.email
}

// GetPassword возвращает хеш пароля пользователя
func (u *User) GetPassword() string {
	return u.password
}

// SetUsername устанавливает новое имя пользователя и обновляет временную метку
func (u *User) SetUsername(newUsername string) {
	u.username = newUsername
	u.updateTimestamp()
}

// SetEmail устанавливает новый email и обновляет временную метку
func (u *User) SetEmail(newEmail string) {
	u.email = newEmail
	u.updateTimestamp()
}

// SetPassword устанавливает новый пароль (хешируется автоматически) и обновляет временную метку
func (u *User) SetPassword(newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.password = string(hashedPassword)
	u.updateTimestamp()
	return nil
}

// SetPasswordHash устанавливает хеш пароля напрямую (для загрузки из базы данных)
func (u *User) SetPasswordHash(passwordHash string) {
	u.password = passwordHash
	u.updateTimestamp()
}

// CheckPassword проверяет, соответствует ли переданный пароль хешу
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.password), []byte(password))
	return err == nil
}

// SetID устанавливает идентификатор пользователя
func (u *User) SetID(id string) {
	u.id = id
}

// SetCreatedAt устанавливает время создания
func (u *User) SetCreatedAt(createdAt time.Time) {
	u.createdAt = createdAt
}

// SetUpdatedAt устанавливает время обновления
func (u *User) SetUpdatedAt(updatedAt time.Time) {
	u.updatedAt = updatedAt
}

// JSONUser вспомогательная структура для JSON сериализации
type JSONUser struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// JSONUserWithPassword вспомогательная структура для JSON сериализации с паролем (для внутреннего хранения)
type JSONUserWithPassword struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MarshalJSON реализует интерфейс json.Marshaler
// При сериализации в JSON пароль не включается в целях безопасности
func (u *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONUser{
		ID:        u.id,
		Username:  u.username,
		Email:     u.email,
		CreatedAt: u.createdAt,
		UpdatedAt: u.updatedAt,
	})
}

// MarshalJSONWithPassword реализует сериализацию с включением пароля (для внутреннего хранения)
func (u *User) MarshalJSONWithPassword() ([]byte, error) {
	return json.Marshal(JSONUserWithPassword{
		ID:        u.id,
		Username:  u.username,
		Email:     u.email,
		Password:  u.password,
		CreatedAt: u.createdAt,
		UpdatedAt: u.updatedAt,
	})
}

// UnmarshalJSONWithPassword реализует десериализацию с извлечением пароля (для внутреннего хранения)
func (u *User) UnmarshalJSONWithPassword(data []byte) error {
	var jsonUser JSONUserWithPassword
	if err := json.Unmarshal(data, &jsonUser); err != nil {
		return err
	}

	u.id = jsonUser.ID
	u.username = jsonUser.Username
	u.email = jsonUser.Email
	u.password = jsonUser.Password
	u.createdAt = jsonUser.CreatedAt
	u.updatedAt = jsonUser.UpdatedAt

	return nil
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler
func (u *User) UnmarshalJSON(data []byte) error {
	// Сначала десериализуем в промежуточную структуру без пароля
	var jsonUser JSONUser
	if err := json.Unmarshal(data, &jsonUser); err != nil {
		return err
	}

	// Теперь попробуем десериализовать в структуру с паролем
	var fullData map[string]interface{}
	if err := json.Unmarshal(data, &fullData); err != nil {
		return err
	}

	u.id = jsonUser.ID
	u.username = jsonUser.Username
	u.email = jsonUser.Email
	u.createdAt = jsonUser.CreatedAt
	u.updatedAt = jsonUser.UpdatedAt

	// Восстанавливаем пароль, если он присутствует в данных
	if password, exists := fullData["password"]; exists {
		if passwordStr, ok := password.(string); ok {
			u.password = passwordStr
		}
	}

	return nil
}
