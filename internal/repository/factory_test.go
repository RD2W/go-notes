package repository

import (
	"sync"
	"testing"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/stretchr/testify/assert"
)

// MockRepository - тестовая реализация репозитория для проверки фабрики
type MockRepository struct {
	_ int // unused field, added to satisfy linter
}

func (m *MockRepository) Save(entity Entity) {
	// пустая реализация для удовлетворения интерфейса
}

func (m *MockRepository) GetAllNotes() []*model.Note {
	// пустая реализация для удовлетворения интерфейса
	return nil
}

func (m *MockRepository) GetNotesCount() int {
	// пустая реализация для удовлетворения интерфейса
	return 0
}

func (m *MockRepository) GetNewNotes(lastIndex int) []*model.Note {
	// пустая реализация для удовлетворения интерфейса
	return nil
}

func (m *MockRepository) GetAllByType(entityType string) []Entity {
	// пустая реализация для удовлетворения интерфейса
	return nil
}

func (m *MockRepository) GetByID(entityType, id string) Entity {
	// пустая реализация для удовлетворения интерфейса
	return nil
}

func (m *MockRepository) DeleteByID(entityType, id string) bool {
	// пустая реализация для удовлетворения интерфейса
	return false
}

func TestRegister(t *testing.T) {
	// Сохраняем оригинальные значения для восстановления
	originalCreators := make(map[StorageType]Creator)
	for k, v := range creators {
		originalCreators[k] = v
	}
	originalDefaultType := defaultType

	// Восстанавливаем оригинальные значения после теста
	defer func() {
		creatorsLock.Lock()
		defer creatorsLock.Unlock()
		creators = make(map[StorageType]Creator)
		for k, v := range originalCreators {
			creators[k] = v
		}
		defaultType = originalDefaultType
	}()

	// Регистрируем новый тип репозитория
	mockType := StorageType("mock")
	mockCreator := func() Repository {
		return &MockRepository{}
	}

	Register(mockType, mockCreator)

	// Проверяем, что создатель был зарегистрирован
	creatorsLock.RLock()
	creator, exists := creators[mockType]
	creatorsLock.RUnlock()

	assert.True(t, exists)
	assert.NotNil(t, creator)

	// Проверяем, что создатель создает экземпляр правильно
	repo := creator()
	assert.IsType(t, &MockRepository{}, repo)
}

func TestNewRepository(t *testing.T) {
	// Сохраняем оригинальные значения для восстановления
	originalCreators := make(map[StorageType]Creator)
	for k, v := range creators {
		originalCreators[k] = v
	}
	originalDefaultType := defaultType

	// Восстанавливаем оригинальные значения после теста
	defer func() {
		creatorsLock.Lock()
		defer creatorsLock.Unlock()
		creators = make(map[StorageType]Creator)
		for k, v := range originalCreators {
			creators[k] = v
		}
		defaultType = originalDefaultType
	}()

	// Регистрируем создателя для типа по умолчанию (JSON)
	jsonCreator := func() Repository {
		return &MockRepository{}
	}
	Register(JSON, jsonCreator)

	// Создаем репозиторий с помощью NewRepository (должен использовать тип по умолчанию)
	repo := NewRepository()

	assert.IsType(t, &MockRepository{}, repo)
}

func TestNewRepositoryByType(t *testing.T) {
	// Сохраняем оригинальные значения для восстановления
	originalCreators := make(map[StorageType]Creator)
	for k, v := range creators {
		originalCreators[k] = v
	}
	originalDefaultType := defaultType

	// Восстанавливаем оригинальные значения после теста
	defer func() {
		creatorsLock.Lock()
		defer creatorsLock.Unlock()
		creators = make(map[StorageType]Creator)
		for k, v := range originalCreators {
			creators[k] = v
		}
		defaultType = originalDefaultType
	}()

	// Регистрируем создателей для разных типов
	jsonCreator := func() Repository {
		return &MockRepository{}
	}
	Register(JSON, jsonCreator)

	ramCreator := func() Repository {
		return &MockRepository{}
	}
	Register(RAM, ramCreator)

	// Тестируем создание репозитория по типу JSON
	repoJSON := NewRepositoryByType(JSON)
	assert.IsType(t, &MockRepository{}, repoJSON)

	// Тестируем создание репозитория по типу RAM
	repoRAM := NewRepositoryByType(RAM)
	assert.IsType(t, &MockRepository{}, repoRAM)
}

func TestNewRepositoryByTypeWithNonExistentType(t *testing.T) {
	// Сохраняем оригинальные значения для восстановления
	originalCreators := make(map[StorageType]Creator)
	for k, v := range creators {
		originalCreators[k] = v
	}
	originalDefaultType := defaultType

	// Восстанавливаем оригинальные значения после теста
	defer func() {
		creatorsLock.Lock()
		defer creatorsLock.Unlock()
		creators = make(map[StorageType]Creator)
		for k, v := range originalCreators {
			creators[k] = v
		}
		defaultType = originalDefaultType
	}()

	// Регистрируем только RAM создатель
	ramCreator := func() Repository {
		return &MockRepository{}
	}
	Register(RAM, ramCreator)

	// Создаем репозиторий с несуществующим типом, ожидаем, что будет использован RAM как fallback
	nonExistentType := StorageType("nonexistent")
	repo := NewRepositoryByType(nonExistentType)

	assert.IsType(t, &MockRepository{}, repo)
}

func TestFactoryWithMultipleGoroutines(t *testing.T) {
	// Сохраняем оригинальные значения для восстановления
	originalCreators := make(map[StorageType]Creator)
	for k, v := range creators {
		originalCreators[k] = v
	}
	originalDefaultType := defaultType

	// Восстанавливаем оригинальные значения после теста
	defer func() {
		creatorsLock.Lock()
		defer creatorsLock.Unlock()
		creators = make(map[StorageType]Creator)
		for k, v := range originalCreators {
			creators[k] = v
		}
		defaultType = originalDefaultType
	}()

	// Регистрируем создателей
	jsonCreator := func() Repository {
		return &MockRepository{}
	}
	Register(JSON, jsonCreator)

	ramCreator := func() Repository {
		return &MockRepository{}
	}
	Register(RAM, ramCreator)

	// Тестируем параллельный доступ к фабрике
	var wg sync.WaitGroup
	const numGoroutines = 10

	// Создаем репозитории в разных горутинах
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var repo Repository
			if i%2 == 0 {
				repo = NewRepositoryByType(JSON)
			} else {
				repo = NewRepositoryByType(RAM)
			}
			assert.IsType(t, &MockRepository{}, repo)
		}(i)
	}

	wg.Wait()
}

func TestDefaultTypeFallback(t *testing.T) {
	// Сохраняем оригинальные значения для восстановления
	originalCreators := make(map[StorageType]Creator)
	for k, v := range creators {
		originalCreators[k] = v
	}
	originalDefaultType := defaultType

	// Восстанавливаем оригинальные значения после теста
	defer func() {
		creatorsLock.Lock()
		defer creatorsLock.Unlock()
		creators = make(map[StorageType]Creator)
		for k, v := range originalCreators {
			creators[k] = v
		}
		defaultType = originalDefaultType
	}()

	// Устанавливаем RAM как тип по умолчанию
	defaultType = RAM

	// Регистрируем создателей
	ramCreator := func() Repository {
		return &MockRepository{}
	}
	Register(RAM, ramCreator)

	// Регистрируем JSON, но не будем использовать его напрямую
	jsonCreator := func() Repository {
		return &MockRepository{}
	}
	Register(JSON, jsonCreator)

	// Создаем репозиторий с помощью NewRepository (должен использовать тип по умолчанию - RAM)
	repo := NewRepository()
	assert.IsType(t, &MockRepository{}, repo)

	// Создаем репозиторий с несуществующим типом (должен использовать RAM как fallback)
	nonExistentType := StorageType("nonexistent")
	repo = NewRepositoryByType(nonExistentType)
	assert.IsType(t, &MockRepository{}, repo)
}
