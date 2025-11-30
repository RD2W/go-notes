package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BaseRepository содержит общую логику для всех репозиториев
type BaseRepository struct {
	db      *pgxpool.Pool
	tx      pgx.Tx // текущая транзакция (если есть)
	timeout time.Duration
}

// NewBaseRepository создает новый экземпляр базового репозитория
func NewBaseRepository(db *pgxpool.Pool, timeout time.Duration) *BaseRepository {
	if timeout <= 0 {
		timeout = 5 * time.Second // значение по умолчанию
	}
	return &BaseRepository{
		db:      db,
		timeout: timeout,
	}
}

// WithContext возвращает контекст с таймаутом
func (r *BaseRepository) WithContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.timeout)
}

// SetTransaction устанавливает транзакцию для репозитория
func (r *BaseRepository) SetTransaction(tx pgx.Tx) {
	r.tx = tx
}

// ClearTransaction убирает транзакцию
func (r *BaseRepository) ClearTransaction() {
	r.tx = nil
}

// GetDB возвращает пул подключений или транзакцию, если она установлена
func (r *BaseRepository) GetDB() interface{} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Exec выполняет SQL-запрос и возвращает результат
func (r *BaseRepository) Exec(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	var commandTag pgconn.CommandTag
	var err error

	if r.tx != nil {
		commandTag, err = r.tx.Exec(ctx, sql, args...)
	} else {
		commandTag, err = r.db.Exec(ctx, sql, args...)
	}

	if err != nil {
		return 0, err
	}
	return commandTag.RowsAffected(), nil
}

// QueryRow выполняет запрос и возвращает одну строку
func (r *BaseRepository) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if r.tx != nil {
		return r.tx.QueryRow(ctx, sql, args...)
	}
	return r.db.QueryRow(ctx, sql, args...)
}

// Query выполняет запрос и возвращает несколько строк
func (r *BaseRepository) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if r.tx != nil {
		return r.tx.Query(ctx, sql, args...)
	}
	return r.db.Query(ctx, sql, args...)
}
