package repository

import (
	"github.com/rd2w/go-notes/internal/domain/model"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	CRUDRepository[*model.User]
	GetAllUsers() ([]*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
}
