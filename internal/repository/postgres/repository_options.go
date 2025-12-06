package postgres

import "time"

// RepositoryOption функция для настройки репозитория
type RepositoryOption func(*BaseRepository)

// WithTimeout устанавливает таймаут для операций
func WithTimeout(timeout time.Duration) RepositoryOption {
	return func(r *BaseRepository) {
		if timeout > 0 {
			r.timeout = timeout
		}
	}
}
