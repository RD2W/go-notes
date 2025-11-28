package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/domain/repository"
	errorsPkg "github.com/rd2w/go-notes/internal/repository/postgres/errors"
)

// PostgresUserRepository реализация интерфейсов репозитория для пользователей с использованием PostgreSQL
type PostgresUserRepository struct {
	*BaseRepository
}

// NewPostgresUserRepository создает новый экземпляр репозитория пользователей с PostgreSQL
func NewPostgresUserRepository(db *pgxpool.Pool) (repository.UserRepository, error) {
	baseRepo := NewBaseRepository(db, 5*time.Second)

	return &PostgresUserRepository{
		BaseRepository: baseRepo,
	}, nil
}

// NewPostgresUserRepositoryWithOpts создает новый экземпляр репозитория пользователей с PostgreSQL с опциями
func NewPostgresUserRepositoryWithOpts(db *pgxpool.Pool, opts ...RepositoryOption) (repository.UserRepository, error) {
	baseRepo := NewBaseRepository(db, 5*time.Second)
	for _, opt := range opts {
		opt(baseRepo)
	}

	return &PostgresUserRepository{
		BaseRepository: baseRepo,
	}, nil
}

// Create добавляет нового пользователя в базу данных
func (r *PostgresUserRepository) Create(user *model.User) error {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	_, err := r.Exec(ctx,
		"INSERT INTO users (id, username, email, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		user.GetID(), user.GetUsername(), user.GetEmail(), user.GetPassword(), user.GetCreatedAt(), user.GetUpdatedAt())

	if err != nil {
		return fmt.Errorf("%w: %w", errorsPkg.ErrUserCreationFailed, err)
	}

	return nil
}

// GetByID возвращает пользователя по его ID
func (r *PostgresUserRepository) GetByID(id string) (*model.User, error) {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	row := r.QueryRow(ctx, "SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = $1", id)

	var userID, username, email, password string
	var createdAt, updatedAt time.Time

	err := row.Scan(&userID, &username, &email, &password, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", errorsPkg.ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", errorsPkg.ErrUserRetrievalFailed, err)
	}

	user := &model.User{}
	user.SetID(userID)
	user.SetUsername(username)
	user.SetEmail(email)
	user.SetPasswordHash(password) // Используем метод для установки хеша пароля напрямую
	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)

	return user, nil
}

// Update обновляет существующего пользователя в базе данных
func (r *PostgresUserRepository) Update(user *model.User) error {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	_, err := r.Exec(ctx,
		"UPDATE users SET username = $1, email = $2, password = $3, updated_at = $4 WHERE id = $5",
		user.GetUsername(), user.GetEmail(), user.GetPassword(), user.GetUpdatedAt(), user.GetID())

	if err != nil {
		return fmt.Errorf("%w: %w", errorsPkg.ErrUserUpdateFailed, err)
	}

	return nil
}

// DeleteByID удаляет пользователя по его ID
func (r *PostgresUserRepository) DeleteByID(id string) error {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	rowsAffected, err := r.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("%w: %w", errorsPkg.ErrUserDeletionFailed, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", errorsPkg.ErrUserNotFound, id)
	}

	return nil
}

// GetAllUsers возвращает всех пользователей
func (r *PostgresUserRepository) GetAllUsers() ([]*model.User, error) {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	rows, err := r.Query(ctx, "SELECT id, username, email, password, created_at, updated_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errorsPkg.ErrAllUsersRetrievalFailed, err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var userID, username, email, password string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&userID, &username, &email, &password, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		user := &model.User{}
		user.SetID(userID)
		user.SetUsername(username)
		user.SetEmail(email)
		user.SetPasswordHash(password)
		user.SetCreatedAt(createdAt)
		user.SetUpdatedAt(updatedAt)

		users = append(users, user)
	}

	return users, nil
}

// GetUserByEmail возвращает пользователя по его email
func (r *PostgresUserRepository) GetUserByEmail(email string) (*model.User, error) {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	row := r.QueryRow(ctx, "SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = $1", email)

	var userID, username, emailResult, password string
	var createdAt, updatedAt time.Time

	err := row.Scan(&userID, &username, &emailResult, &password, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", errorsPkg.ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", errorsPkg.ErrUserByEmailRetrievalFailed, err)
	}

	user := &model.User{}
	user.SetID(userID)
	user.SetUsername(username)
	user.SetEmail(emailResult)
	user.SetPasswordHash(password)
	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)

	return user, nil
}

// GetUserByUsername возвращает пользователя по его имени пользователя
func (r *PostgresUserRepository) GetUserByUsername(username string) (*model.User, error) {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	row := r.QueryRow(ctx, "SELECT id, username, email, password, created_at, updated_at FROM users WHERE username = $1", username)

	var userID, usernameResult, email, password string
	var createdAt, updatedAt time.Time

	err := row.Scan(&userID, &usernameResult, &email, &password, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", errorsPkg.ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", errorsPkg.ErrUserByUsernameRetrievalFailed, err)
	}

	user := &model.User{}
	user.SetID(userID)
	user.SetUsername(usernameResult)
	user.SetEmail(email)
	user.SetPasswordHash(password)
	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)

	return user, nil
}
