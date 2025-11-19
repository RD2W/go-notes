package user

import (
	"errors"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/domain/repository"
	"github.com/rd2w/go-notes/internal/domain/service"
)

// userService реализация бизнес-логики для пользователей
type userService struct {
	repo repository.Repository
}

// NewUserService создает новый экземпляр сервиса пользователей
func NewUserService(repo repository.Repository) service.UserService {
	return &userService{
		repo: repo,
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

	s.repo.Save(user)
	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (s *userService) GetUserByID(id string) (*model.User, error) {
	entity := s.repo.GetByID("user", id)
	if entity == nil {
		return nil, errors.New("пользователь не найден")
	}

	user, ok := entity.(*model.User)
	if !ok {
		return nil, errors.New("ошибка преобразования сущности")
	}

	return user, nil
}

// UpdateUser обновляет пользователя
func (s *userService) UpdateUser(id, username, email string) (*model.User, error) {
	entity := s.repo.GetByID("user", id)
	if entity == nil {
		return nil, errors.New("пользователь не найден")
	}

	user, ok := entity.(*model.User)
	if !ok {
		return nil, errors.New("ошибка преобразования сущности")
	}

	if username != "" {
		user.SetUsername(username)
	}
	if email != "" {
		user.SetEmail(email)
	}

	s.repo.Save(user)
	return user, nil
}

// DeleteUser удаляет пользователя
func (s *userService) DeleteUser(id string) error {
	deleted := s.repo.DeleteByID("user", id)
	if !deleted {
		return errors.New("пользователь не найден")
	}
	return nil
}

// GetAllUsers возвращает всех пользователей
func (s *userService) GetAllUsers() ([]*model.User, error) {
	entities := s.repo.GetAllByType("user")
	users := make([]*model.User, 0)

	for _, entity := range entities {
		user, ok := entity.(*model.User)
		if !ok {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserByUsername возвращает пользователя по имени
func (s *userService) GetUserByUsername(username string) (*model.User, error) {
	entities := s.repo.GetAllByType("user")
	for _, entity := range entities {
		user, ok := entity.(*model.User)
		if ok && user.GetUsername() == username {
			return user, nil
		}
	}

	return nil, errors.New("пользователь не найден")
}
