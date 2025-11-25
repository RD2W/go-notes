package user

import (
	"errors"
	"testing"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository - мок-реализация репозитория пользователей
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetAllUsers() ([]*model.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func TestNewUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	assert.NotNil(t, service)
}

func TestCreateUser(t *testing.T) {
	t.Run("успешное создание пользователя", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		username := "testuser"
		email := "test@example.com"
		password := "password123"

		mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil).Once()

		result, err := service.CreateUser(username, email, password)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, username, result.GetUsername())
		assert.Equal(t, email, result.GetEmail())
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при пустом имени пользователя", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		result, err := service.CreateUser("", "test@example.com", "password123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "все поля обязательны для заполнения", err.Error())
	})

	t.Run("ошибка при пустом email", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		result, err := service.CreateUser("testuser", "", "password123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "все поля обязательны для заполнения", err.Error())
	})

	t.Run("ошибка при пустом пароле", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		result, err := service.CreateUser("testuser", "test@example.com", "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "все поля обязательны для заполнения", err.Error())
	})

	t.Run("ошибка при создании пользователя в репозитории", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		username := "testuser"
		email := "test@example.com"
		password := "password123"

		mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(errors.New("ошибка репозитория")).Once()

		result, err := service.CreateUser(username, email, password)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "ошибка репозитория", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("успешное получение пользователя по ID", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		id := "123"
		user, err := model.NewUser("testuser", "test@example.com", "password123")
		assert.NoError(t, err)
		user.SetID(id)

		mockRepo.On("GetByID", id).Return(user, nil).Once()

		result, err := service.GetUserByID(id)

		assert.NoError(t, err)
		assert.Equal(t, user, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении пользователя по несуществующему ID", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		id := "123"

		mockRepo.On("GetByID", id).Return(nil, errors.New("пользователь не найден")).Once()

		result, err := service.GetUserByID(id)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "пользователь не найден", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("успешное обновление пользователя", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		id := "123"
		username := "updateduser"
		email := "updated@example.com"

		existingUser, err := model.NewUser("testuser", "test@example.com", "password123")
		assert.NoError(t, err)
		existingUser.SetID(id)

		updatedUser := existingUser
		updatedUser.SetUsername(username)
		updatedUser.SetEmail(email)

		mockRepo.On("GetByID", id).Return(existingUser, nil).Once()
		mockRepo.On("Update", updatedUser).Return(nil).Once()

		result, err := service.UpdateUser(id, username, email)

		assert.NoError(t, err)
		assert.Equal(t, updatedUser, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("обновление только имени пользователя", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		id := "123"
		username := "updateduser"
		email := ""

		existingUser, err := model.NewUser("testuser", "test@example.com", "password123")
		assert.NoError(t, err)
		existingUser.SetID(id)

		updatedUser := existingUser
		updatedUser.SetUsername(username)

		mockRepo.On("GetByID", id).Return(existingUser, nil).Once()
		mockRepo.On("Update", updatedUser).Return(nil).Once()

		result, err := service.UpdateUser(id, username, email)

		assert.NoError(t, err)
		assert.Equal(t, updatedUser, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("обновление только email", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		id := "123"
		username := ""
		email := "updated@example.com"

		existingUser, err := model.NewUser("testuser", "test@example.com", "password123")
		assert.NoError(t, err)
		existingUser.SetID(id)

		updatedUser := existingUser
		updatedUser.SetEmail(email)

		mockRepo.On("GetByID", id).Return(existingUser, nil).Once()
		mockRepo.On("Update", updatedUser).Return(nil).Once()

		result, err := service.UpdateUser(id, username, email)

		assert.NoError(t, err)
		assert.Equal(t, updatedUser, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при обновлении несуществующего пользователя", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		id := "123"
		username := "updateduser"
		email := "updated@example.com"

		mockRepo.On("GetByID", id).Return(nil, errors.New("пользователь не найден")).Once()

		result, err := service.UpdateUser(id, username, email)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "пользователь не найден", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при обновлении пользователя в репозитории", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		id := "123"
		username := "updateduser"
		email := "updated@example.com"

		existingUser, err := model.NewUser("testuser", "test@example.com", "password123")
		assert.NoError(t, err)
		existingUser.SetID(id)

		mockRepo.On("GetByID", id).Return(existingUser, nil).Once()
		mockRepo.On("Update", mock.AnythingOfType("*model.User")).Return(errors.New("ошибка репозитория")).Once()

		result, err := service.UpdateUser(id, username, email)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "ошибка репозитория", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("успешное удаление пользователя", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		id := "123"

		mockRepo.On("DeleteByID", id).Return(nil).Once()

		err := service.DeleteUser(id)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при удалении несуществующего пользователя", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		id := "123"

		mockRepo.On("DeleteByID", id).Return(errors.New("пользователь не найден")).Once()

		err := service.DeleteUser(id)

		assert.Error(t, err)
		assert.Equal(t, "пользователь не найден", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllUsers(t *testing.T) {
	t.Run("успешное получение всех пользователей", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		user1, err := model.NewUser("user1", "user1@example.com", "password123")
		assert.NoError(t, err)
		user1.SetID("1")

		user2, err := model.NewUser("user2", "user2@example.com", "password123")
		assert.NoError(t, err)
		user2.SetID("2")

		users := []*model.User{user1, user2}

		mockRepo.On("GetAllUsers").Return(users, nil).Once()

		result, err := service.GetAllUsers()

		assert.NoError(t, err)
		assert.Equal(t, users, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении всех пользователей", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		mockRepo.On("GetAllUsers").Return(nil, errors.New("ошибка репозитория")).Once()

		result, err := service.GetAllUsers()

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "ошибка репозитория", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserByUsername(t *testing.T) {
	t.Run("успешное получение пользователя по имени", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		username := "testuser"
		user, err := model.NewUser(username, "test@example.com", "password123")
		assert.NoError(t, err)
		user.SetID("123")

		mockRepo.On("GetUserByUsername", username).Return(user, nil).Once()

		result, err := service.GetUserByUsername(username)

		assert.NoError(t, err)
		assert.Equal(t, user, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении пользователя по несуществующему имени", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		username := "nonexistent"

		mockRepo.On("GetUserByUsername", username).Return(nil, errors.New("пользователь не найден")).Once()

		result, err := service.GetUserByUsername(username)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "пользователь не найден", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Run("успешное получение пользователя по email", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		email := "test@example.com"
		username := "testuser"
		user, err := model.NewUser(username, email, "password123")
		assert.NoError(t, err)
		user.SetID("123")

		mockRepo.On("GetUserByEmail", email).Return(user, nil).Once()

		result, err := service.GetUserByEmail(email)

		assert.NoError(t, err)
		assert.Equal(t, user, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении пользователя по несуществующему email", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		email := "nonexistent@example.com"

		mockRepo.On("GetUserByEmail", email).Return(nil, errors.New("пользователь не найден")).Once()

		result, err := service.GetUserByEmail(email)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "пользователь не найден", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
