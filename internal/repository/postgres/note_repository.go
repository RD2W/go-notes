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

// PostgresNoteRepository реализация интерфейсов репозитория для заметок с использованием PostgreSQL
type PostgresNoteRepository struct {
	*BaseRepository
}

// NewPostgresNoteRepository создает новый экземпляр репозитория заметок с PostgreSQL
func NewPostgresNoteRepository(db *pgxpool.Pool) (repository.NoteRepository, error) {
	baseRepo := NewBaseRepository(db, 5*time.Second)

	return &PostgresNoteRepository{
		BaseRepository: baseRepo,
	}, nil
}

// NewPostgresNoteRepositoryWithOpts создает новый экземпляр репозитория заметок с PostgreSQL с опциями
func NewPostgresNoteRepositoryWithOpts(db *pgxpool.Pool, opts ...RepositoryOption) (repository.NoteRepository, error) {
	baseRepo := NewBaseRepository(db, 5*time.Second)
	for _, opt := range opts {
		opt(baseRepo)
	}

	return &PostgresNoteRepository{
		BaseRepository: baseRepo,
	}, nil
}

// Create добавляет новую заметку в базу данных
func (r *PostgresNoteRepository) Create(note *model.Note) error {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	_, err := r.Exec(ctx,
		"INSERT INTO notes (id, title, content, user_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		note.GetID(), note.GetTitle(), note.GetContent(), note.GetUserID(), note.GetCreatedAt(), note.GetUpdatedAt())

	if err != nil {
		return fmt.Errorf("%w: %w", errorsPkg.ErrNoteCreationFailed, err)
	}

	return nil
}

// GetByID возвращает заметку по её ID
func (r *PostgresNoteRepository) GetByID(id string) (*model.Note, error) {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	row := r.QueryRow(ctx, "SELECT id, title, content, user_id, created_at, updated_at FROM notes WHERE id = $1", id)

	var noteID, title, content, userID string
	var createdAt, updatedAt time.Time

	err := row.Scan(&noteID, &title, &content, &userID, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", errorsPkg.ErrNoteNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", errorsPkg.ErrNoteRetrievalFailed, err)
	}

	note := model.NewNote(title, content, userID)
	note.SetID(noteID)
	note.SetCreatedAt(createdAt)
	note.SetUpdatedAt(updatedAt)

	return note, nil
}

// Update обновляет существующую заметку в базе данных
func (r *PostgresNoteRepository) Update(note *model.Note) error {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	_, err := r.Exec(ctx,
		"UPDATE notes SET title = $1, content = $2, user_id = $3, updated_at = $4 WHERE id = $5",
		note.GetTitle(), note.GetContent(), note.GetUserID(), note.GetUpdatedAt(), note.GetID())

	if err != nil {
		return fmt.Errorf("%w: %w", errorsPkg.ErrNoteUpdateFailed, err)
	}

	return nil
}

// DeleteByID удаляет заметку по её ID
func (r *PostgresNoteRepository) DeleteByID(id string) error {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	rowsAffected, err := r.Exec(ctx, "DELETE FROM notes WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("%w: %w", errorsPkg.ErrNoteDeletionFailed, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", errorsPkg.ErrNoteNotFound, id)
	}

	return nil
}

// GetAllNotesByUserID возвращает все заметки пользователя по его ID
func (r *PostgresNoteRepository) GetAllNotesByUserID(userID string) ([]*model.Note, error) {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	rows, err := r.Query(ctx, "SELECT id, title, content, user_id, created_at, updated_at FROM notes WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errorsPkg.ErrNotesRetrievalByUserFailed, err)
	}
	defer rows.Close()

	var notes []*model.Note
	for rows.Next() {
		var noteID, title, content, noteUserID string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&noteID, &title, &content, &noteUserID, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		note := model.NewNote(title, content, noteUserID)
		note.SetID(noteID)
		note.SetCreatedAt(createdAt)
		note.SetUpdatedAt(updatedAt)

		notes = append(notes, note)
	}

	return notes, nil
}

// GetListByUserID возвращает список заметок пользователя с пагинацией
func (r *PostgresNoteRepository) GetListByUserID(userID string, limit, offset int) ([]*model.Note, error) {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	rows, err := r.Query(ctx, "SELECT id, title, content, user_id, created_at, updated_at FROM notes WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errorsPkg.ErrNotesListRetrievalFailed, err)
	}
	defer rows.Close()

	var notes []*model.Note
	for rows.Next() {
		var noteID, title, content, noteUserID string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&noteID, &title, &content, &noteUserID, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		note := model.NewNote(title, content, noteUserID)
		note.SetID(noteID)
		note.SetCreatedAt(createdAt)
		note.SetUpdatedAt(updatedAt)

		notes = append(notes, note)
	}

	return notes, nil
}

// GetAllNotes возвращает все заметки
func (r *PostgresNoteRepository) GetAllNotes() ([]*model.Note, error) {
	ctx, cancel := r.WithContext(context.Background())
	defer cancel()

	rows, err := r.Query(ctx, "SELECT id, title, content, user_id, created_at, updated_at FROM notes ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errorsPkg.ErrAllNotesRetrievalFailed, err)
	}
	defer rows.Close()

	var notes []*model.Note
	for rows.Next() {
		var id, title, content, userID string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &title, &content, &userID, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		note := model.NewNote(title, content, userID)
		note.SetID(id)
		note.SetCreatedAt(createdAt)
		note.SetUpdatedAt(updatedAt)

		notes = append(notes, note)
	}

	return notes, nil
}
