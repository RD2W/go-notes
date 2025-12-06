package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionManager интерфейс для управления транзакциями
type TransactionManager interface {
	// WithinTransaction выполняет функцию в транзакции
	WithinTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error
}

// PostgresTransactionManager реализация TransactionManager для PostgreSQL
type PostgresTransactionManager struct {
	db *pgxpool.Pool
}

// NewPostgresTransactionManager создает новый менеджер транзакций
func NewPostgresTransactionManager(db *pgxpool.Pool) *PostgresTransactionManager {
	return &PostgresTransactionManager{
		db: db,
	}
}

// WithinTransaction выполняет функцию в транзакции
func (tm *PostgresTransactionManager) WithinTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			rbErr := tx.Rollback(ctx)
			if rbErr != nil {
				fmt.Printf("Error rolling back transaction after panic: %v\n", rbErr)
			}
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		rbErr := tx.Rollback(ctx)
		if rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
