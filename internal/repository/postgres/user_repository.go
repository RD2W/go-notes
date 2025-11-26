package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/domain/repository"
	"github.com/rd2w/go-notes/internal/repository/postgres/errors"
)

// PostgresUserRepository реализация интерфейсов репозитория для пользователей с использованием PostgreSQL
type PostgresUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepository создает новый экземпляр репозитория пользователей с PostgreSQL
func NewPostgresUserRepository(db *pgxpool.Pool) (repository.UserRepository, error) {
	return &PostgresUserRepository{
		db: db,
	}, nil
}

// Реализация CRUDRepository для пользователей
func (r *PostgresUserRepository) Create(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.db.Exec(ctx,
		"INSERT INTO users (id, username, email, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		user.GetID(), user.GetUsername(), user.GetEmail(), user.GetPassword(), user.GetCreatedAt(), user.GetUpdatedAt())

	if err != nil {
		return fmt.Errorf("%w: %w", errors.ErrUserCreationFailed, err)
	}

	return nil
}

func (r *PostgresUserRepository) GetByID(id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.db.QueryRow(ctx, "SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = $1", id)

	var userID, username, email, password string
	var createdAt, updatedAt time.Time

	err := row.Scan(&userID, &username, &email, &password, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errors.ErrUserRetrievalFailed, err)
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

func (r *PostgresUserRepository) Update(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.db.Exec(ctx,
		"UPDATE users SET username = $1, email = $2, password = $3, updated_at = $4 WHERE id = $5",
		user.GetUsername(), user.GetEmail(), user.GetPassword(), user.GetUpdatedAt(), user.GetID())

	if err != nil {
		return fmt.Errorf("%w: %w", errors.ErrUserUpdateFailed, err)
	}

	return nil
}

func (r *PostgresUserRepository) DeleteByID(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	commandTag, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("%w: %w", errors.ErrUserDeletionFailed, err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", errors.ErrUserNotFound, id)
	}

	return nil
}

// Реализация методов UserRepository
func (r *PostgresUserRepository) GetAllUsers() ([]*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, "SELECT id, username, email, password, created_at, updated_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errors.ErrAllUsersRetrievalFailed, err)
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

func (r *PostgresUserRepository) GetUserByEmail(email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.db.QueryRow(ctx, "SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = $1", email)

	var userID, username, emailResult, password string
	var createdAt, updatedAt time.Time

	err := row.Scan(&userID, &username, &emailResult, &password, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errors.ErrUserByEmailRetrievalFailed, err)
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

func (r *PostgresUserRepository) GetUserByUsername(username string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.db.QueryRow(ctx, "SELECT id, username, email, password, created_at, updated_at FROM users WHERE username = $1", username)

	var userID, usernameResult, email, password string
	var createdAt, updatedAt time.Time

	err := row.Scan(&userID, &usernameResult, &email, &password, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errors.ErrUserByUsernameRetrievalFailed, err)
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
