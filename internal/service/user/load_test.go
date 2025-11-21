package user

import (
	"sync"
	"testing"

	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestConcurrentUserCreation tests concurrent user creation
func TestConcurrentUserCreation(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Ожидаем, что Create будет вызван 10 раз
	mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil).Times(10)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := service.CreateUser(
				"user"+string(rune(id))+"@example.com",
				"User "+string(rune(id)),
				"password123",
			)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
	mockRepo.AssertExpectations(t)
}

// TestConcurrentUserRetrieval tests concurrent user retrieval
func TestConcurrentUserRetrieval(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Создаем тестового пользователя
	user := model.NewUserWithPasswordHash("Test User", "test@example.com", "hashedPassword123")
	user.SetID("user123")

	// Ожидаем, что GetByID будет вызван 10 раз
	mockRepo.On("GetByID", "user123").Return(user, nil).Times(10)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.GetUserByID("user123")
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	mockRepo.AssertExpectations(t)
}

// TestMixedConcurrentUserOperations tests mixed concurrent operations
func TestMixedConcurrentUserOperations(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Подготовим ожидания для разных операций
	user := model.NewUserWithPasswordHash("Test User", "test@example.com", "hashedPassword123")
	user.SetID("user123")

	mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil).Times(5)
	mockRepo.On("GetByID", "user123").Return(user, nil).Times(5)
	mockRepo.On("GetAllUsers").Return([]*model.User{user}, nil).Times(5)

	var wg sync.WaitGroup

	// Создание пользователей
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := service.CreateUser(
				"user"+string(rune(id))+"@example.com",
				"User "+string(rune(id)),
				"password123",
			)
			assert.NoError(t, err)
		}(i)
	}

	// Получение пользователей
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.GetUserByID("user123")
			assert.NoError(t, err)
		}()
	}

	// Получение всех пользователей
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.GetAllUsers()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	mockRepo.AssertExpectations(t)
}

// BenchmarkConcurrentUserCreation benchmarks concurrent user creation
func BenchmarkConcurrentUserCreation(b *testing.B) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, err := service.CreateUser(
					"benchmark_user"+string(rune(id+i*10))+"@example.com",
					"Benchmark User "+string(rune(id+i*10)),
					"password123",
				)
				if err != nil {
					b.Logf("Error creating user: %v", err)
				}
			}(j)
		}

		wg.Wait()
	}
	mockRepo.AssertExpectations(b)
}

// BenchmarkConcurrentUserRetrieval benchmarks concurrent user retrieval
func BenchmarkConcurrentUserRetrieval(b *testing.B) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	user := model.NewUserWithPasswordHash("Benchmark User", "benchmark@example.com", "hashedPassword123")
	user.SetID("benchmark_user_123")

	mockRepo.On("GetByID", "benchmark_user_123").Return(user, nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := service.GetUserByID("benchmark_user_123")
				if err != nil {
					b.Logf("Error retrieving user: %v", err)
				}
			}()
		}

		wg.Wait()
	}
	mockRepo.AssertExpectations(b)
}
