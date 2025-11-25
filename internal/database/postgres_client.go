package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rd2w/go-notes/internal/config"
)

// PostgresClient клиент для работы с PostgreSQL
type PostgresClient struct {
	pool *pgxpool.Pool
}

// NewPostgresClient создает новый клиент для работы с PostgreSQL
func NewPostgresClient(cfg config.PostgresConfig) (*PostgresClient, error) {
	dsn := buildConnectionString(cfg)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ошибка пинга базы данных: %w", err)
	}

	client := &PostgresClient{
		pool: pool,
	}

	log.Println("✓ Подключение к PostgreSQL успешно установлено")

	return client, nil
}

// buildConnectionString создает строку подключения к PostgreSQL из конфигурации
func buildConnectionString(cfg config.PostgresConfig) string {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	if cfg.Parameters != "" {
		connStr += " " + cfg.Parameters
	}

	return connStr
}

// GetPool возвращает пул подключений
func (c *PostgresClient) GetPool() *pgxpool.Pool {
	return c.pool
}

// Close закрывает подключение к базе данных
func (c *PostgresClient) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

// Ping проверяет подключение к базе данных
func (c *PostgresClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.pool.Ping(ctx)
}
