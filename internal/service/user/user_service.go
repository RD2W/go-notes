package user

import (
	"errors"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/domain/repository"
	"github.com/rd2w/go-notes/internal/domain/service"
)

// userService реализация бизнес-логики для пользователей
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService создает новый экземпляр сервиса пользователей
func NewUserService(userRepo repository.UserRepository) service.UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser создает нового пользователя
func (s *userService) CreateUser(username, email, password string) (*model.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, errors.New("все поля обязательны для заполнения")
	}

	user, err := model.NewUser(username, email, password)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (s *userService) GetUserByID(id string) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	return user, nil
}

// UpdateUser обновляет пользователя
func (s *userService) UpdateUser(id, username, email string) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if username != "" {
		user.SetUsername(username)
	}
	if email != "" {
		user.SetEmail(email)
	}

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser удаляет пользователя
func (s *userService) DeleteUser(id string) error {
	err := s.userRepo.DeleteByID(id)
	if err != nil {
		return errors.New("пользователь не найден")
	}
	return nil
}

// GetAllUsers возвращает всех пользователей
func (s *userService) GetAllUsers() ([]*model.User, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserByUsername возвращает пользователя по имени
func (s *userService) GetUserByUsername(username string) (*model.User, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	return user, nil
}

// GetUserByEmail возвращает пользователя по email
func (s *userService) GetUserByEmail(email string) (*model.User, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	return user, nil
}
