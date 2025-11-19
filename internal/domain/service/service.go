package service

import (
	"github.com/rd2w/go-notes/internal/domain/model"
)

// NoteService интерфейс для бизнес-логики заметок
type NoteService interface {
	CreateNote(title, content string) (*model.Note, error)
	GetNoteByID(id string) (*model.Note, error)
	UpdateNote(id, title, content string) (*model.Note, error)
	DeleteNote(id string) error
	GetAllNotes() ([]*model.Note, error)
}

// UserService интерфейс для бизнес-логики пользователей
type UserService interface {
	CreateUser(username, email, password string) (*model.User, error)
	GetUserByID(id string) (*model.User, error)
	UpdateUser(id, username, email string) (*model.User, error)
	DeleteUser(id string) error
	GetAllUsers() ([]*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
}

// AuthService интерфейс для бизнес-логики аутентификации
type AuthService interface {
	Login(username, password string) (accessToken, refreshToken string, err error)
	Logout(refreshToken string) error
	RefreshTokens(refreshToken string) (newAccessToken, newRefreshToken string, err error)
	ValidateToken(token string) (username string, err error)
}
