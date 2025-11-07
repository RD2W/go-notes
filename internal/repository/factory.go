package repository

import (
	"sync"
)

// StorageType тип хранилища
type StorageType string

const (
	RAM  StorageType = "memory"
	JSON StorageType = "json"
)

// Creator функция для создания репозитория
type Creator func() Repository

var (
	creators     = make(map[StorageType]Creator)
	creatorsLock sync.RWMutex
	defaultType  = JSON
)

// Register регистрирует создателя репозитория для определенного типа
func Register(repoType StorageType, creator Creator) {
	creatorsLock.Lock()
	defer creatorsLock.Unlock()
	creators[repoType] = creator
}

// NewRepository создает новый экземпляр репозитория
func NewRepository() Repository {
	return NewRepositoryByType(defaultType)
}

// NewRepositoryByType создает репозиторий указанного типа
func NewRepositoryByType(repoType StorageType) Repository {
	creatorsLock.RLock()
	defer creatorsLock.RUnlock()

	creator, exists := creators[repoType]
	if !exists {
		creator = creators[RAM]
	}
	return creator()
}
