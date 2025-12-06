package repository

// Entity интерфейс, который должны реализовывать все сущности
type Entity interface {
	GetID() string
	GetType() string
}

// CRUDRepository общий интерфейс для CRUD операций
type CRUDRepository[T Entity] interface {
	Create(entity T) error
	GetByID(id string) (T, error)
	Update(entity T) error
	DeleteByID(id string) error
}
