package user

import (
	"testing"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// BenchmarkUserServiceCreateUser benchmarks the CreateUser method
func BenchmarkUserServiceCreateUser(b *testing.B) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateUser("test@example.com", "John Doe", "password123")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkUserServiceGetUserByID benchmarks the GetUserByID method
func BenchmarkUserServiceGetUserByID(b *testing.B) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	expectedUser := model.NewUserWithPasswordHash("John Doe", "test@example.com", "hashedPassword123")
	expectedUser.SetID("user123")

	mockRepo.On("GetByID", "user123").Return(expectedUser, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetUserByID("user123")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkUserServiceGetUserByEmail benchmarks the GetUserByEmail method
func BenchmarkUserServiceGetUserByEmail(b *testing.B) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	expectedUser := model.NewUserWithPasswordHash("John Doe", "test@example.com", "hashedPassword123")
	expectedUser.SetID("user123")

	mockRepo.On("GetUserByEmail", "test@example.com").Return(expectedUser, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetUserByEmail("test@example.com")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkUserServiceUpdateUser benchmarks the UpdateUser method
func BenchmarkUserServiceUpdateUser(b *testing.B) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	existingUser := model.NewUserWithPasswordHash("John Doe", "test@example.com", "hashedPassword123")
	existingUser.SetID("user123")

	mockRepo.On("GetByID", "user123").Return(existingUser, nil)
	mockRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.UpdateUser("user123", "updated@example.com", "John Updated")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkUserServiceDeleteUser benchmarks the DeleteUser method
func BenchmarkUserServiceDeleteUser(b *testing.B) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("DeleteByID", "user123").Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.DeleteUser("user123")
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkUserServiceGetAllUsers benchmarks the GetAllUsers method
func BenchmarkUserServiceGetAllUsers(b *testing.B) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	users := []*model.User{
		model.NewUserWithPasswordHash("User One", "test1@example.com", "hashedPassword123"),
		model.NewUserWithPasswordHash("User Two", "test2@example.com", "hashedPassword123"),
	}

	mockRepo.On("GetAllUsers").Return(users, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetAllUsers()
		assert.NoError(b, err)
	}
	mockRepo.AssertExpectations(b)
}
