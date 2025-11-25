package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rd2w/go-notes/internal/domain/model"
	"github.com/rd2w/go-notes/internal/domain/repository"
)

// PostgresNoteRepository реализация интерфейсов репозитория для заметок с использованием PostgreSQL
type PostgresNoteRepository struct {
	db *pgxpool.Pool
}

// NewPostgresNoteRepository создает новый экземпляр репозитория заметок с PostgreSQL
func NewPostgresNoteRepository(db *pgxpool.Pool) (repository.NoteRepository, error) {
	return &PostgresNoteRepository{
		db: db,
	}, nil
}

// Реализация CRUDRepository для заметок
func (r *PostgresNoteRepository) Create(note *model.Note) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.db.Exec(ctx,
		"INSERT INTO notes (id, title, content, user_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		note.GetID(), note.GetTitle(), note.GetContent(), note.GetUserID(), note.GetCreatedAt(), note.GetUpdatedAt())

	if err != nil {
		return fmt.Errorf("ошибка создания заметки: %w", err)
	}

	return nil
}

func (r *PostgresNoteRepository) GetByID(id string) (*model.Note, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.db.QueryRow(ctx, "SELECT id, title, content, user_id, created_at, updated_at FROM notes WHERE id = $1", id)

	var noteID, title, content, userID string
	var createdAt, updatedAt time.Time

	err := row.Scan(&noteID, &title, &content, &userID, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заметки: %w", err)
	}

	note := model.NewNote(title, content, userID)
	note.SetID(noteID)
	note.SetCreatedAt(createdAt)
	note.SetUpdatedAt(updatedAt)

	return note, nil
}

func (r *PostgresNoteRepository) Update(note *model.Note) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.db.Exec(ctx,
		"UPDATE notes SET title = $1, content = $2, user_id = $3, updated_at = $4 WHERE id = $5",
		note.GetTitle(), note.GetContent(), note.GetUserID(), note.GetUpdatedAt(), note.GetID())

	if err != nil {
		return fmt.Errorf("ошибка обновления заметки: %w", err)
	}

	return nil
}

func (r *PostgresNoteRepository) DeleteByID(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	commandTag, err := r.db.Exec(ctx, "DELETE FROM notes WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления заметки: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("заметка с ID %s не найдена", id)
	}

	return nil
}

// Реализация методов NoteRepository
func (r *PostgresNoteRepository) GetAllNotesByUserID(userID string) ([]*model.Note, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, "SELECT id, title, content, user_id, created_at, updated_at FROM notes WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заметок пользователя: %w", err)
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

func (r *PostgresNoteRepository) GetListByUserID(userID string, limit, offset int) ([]*model.Note, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, "SELECT id, title, content, user_id, created_at, updated_at FROM notes WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка заметок пользователя: %w", err)
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, "SELECT id, title, content, user_id, created_at, updated_at FROM notes ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех заметок: %w", err)
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
