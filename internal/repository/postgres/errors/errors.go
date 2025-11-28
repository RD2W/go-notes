package errors

import (
	"errors"
	"fmt"
)

// ErrorCode тип для кодов ошибок
type ErrorCode string

const (
	// Ошибки для заметок
	NoteNotFound               ErrorCode = "NOTE_NOT_FOUND"
	NoteCreationFailed         ErrorCode = "NOTE_CREATION_FAILED"
	NoteRetrievalFailed        ErrorCode = "NOTE_RETRIEVAL_FAILED"
	NoteUpdateFailed           ErrorCode = "NOTE_UPDATE_FAILED"
	NoteDeletionFailed         ErrorCode = "NOTE_DELETION_FAILED"
	NotesRetrievalByUserFailed ErrorCode = "NOTES_RETRIEVAL_BY_USER_FAILED"
	NotesListRetrievalFailed   ErrorCode = "NOTES_LIST_RETRIEVAL_FAILED"
	AllNotesRetrievalFailed    ErrorCode = "ALL_NOTES_RETRIEVAL_FAILED"

	// Ошибки для пользователей
	UserNotFound                  ErrorCode = "USER_NOT_FOUND"
	UserCreationFailed            ErrorCode = "USER_CREATION_FAILED"
	UserRetrievalFailed           ErrorCode = "USER_RETRIEVAL_FAILED"
	UserUpdateFailed              ErrorCode = "USER_UPDATE_FAILED"
	UserDeletionFailed            ErrorCode = "USER_DELETION_FAILED"
	AllUsersRetrievalFailed       ErrorCode = "ALL_USERS_RETRIEVAL_FAILED"
	UserByEmailRetrievalFailed    ErrorCode = "USER_BY_EMAIL_RETRIEVAL_FAILED"
	UserByUsernameRetrievalFailed ErrorCode = "USER_BY_USERNAME_RETRIEVAL_FAILED"
)

// RepositoryError структура для детализированных ошибок репозитория
type RepositoryError struct {
	Code    ErrorCode
	Message string
	Err     error
	Entity  string // для какой сущности произошла ошибка
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

func (e *RepositoryError) Is(target error) bool {
	var repoErr *RepositoryError
	if errors.As(target, &repoErr) {
		return e.Code == repoErr.Code
	}
	return errors.Is(e.Err, target)
}

// Функции для создания специфичных ошибок
func NewNoteNotFoundError(id string) error {
	return &RepositoryError{
		Code:    NoteNotFound,
		Message: fmt.Sprintf("note with id %s not found", id),
		Entity:  "note",
	}
}

func NewNoteCreationError(err error) error {
	return &RepositoryError{
		Code:    NoteCreationFailed,
		Message: "failed to create note",
		Err:     err,
		Entity:  "note",
	}
}

func NewNoteRetrievalError(err error) error {
	return &RepositoryError{
		Code:    NoteRetrievalFailed,
		Message: "failed to retrieve note",
		Err:     err,
		Entity:  "note",
	}
}

func NewNoteUpdateError(err error) error {
	return &RepositoryError{
		Code:    NoteUpdateFailed,
		Message: "failed to update note",
		Err:     err,
		Entity:  "note",
	}
}

func NewNoteDeletionError(err error) error {
	return &RepositoryError{
		Code:    NoteDeletionFailed,
		Message: "failed to delete note",
		Err:     err,
		Entity:  "note",
	}
}

func NewNotesRetrievalByUserError(err error) error {
	return &RepositoryError{
		Code:    NotesRetrievalByUserFailed,
		Message: "failed to retrieve notes by user",
		Err:     err,
		Entity:  "note",
	}
}

func NewNotesListRetrievalError(err error) error {
	return &RepositoryError{
		Code:    NotesListRetrievalFailed,
		Message: "failed to retrieve notes list by user",
		Err:     err,
		Entity:  "note",
	}
}

func NewAllNotesRetrievalError(err error) error {
	return &RepositoryError{
		Code:    AllNotesRetrievalFailed,
		Message: "failed to retrieve all notes",
		Err:     err,
		Entity:  "note",
	}
}

func NewUserNotFoundError(id string) error {
	return &RepositoryError{
		Code:    UserNotFound,
		Message: fmt.Sprintf("user with id %s not found", id),
		Entity:  "user",
	}
}

func NewUserCreationError(err error) error {
	return &RepositoryError{
		Code:    UserCreationFailed,
		Message: "failed to create user",
		Err:     err,
		Entity:  "user",
	}
}

func NewUserRetrievalError(err error) error {
	return &RepositoryError{
		Code:    UserRetrievalFailed,
		Message: "failed to retrieve user",
		Err:     err,
		Entity:  "user",
	}
}

func NewUserUpdateError(err error) error {
	return &RepositoryError{
		Code:    UserUpdateFailed,
		Message: "failed to update user",
		Err:     err,
		Entity:  "user",
	}
}

func NewUserDeletionError(err error) error {
	return &RepositoryError{
		Code:    UserDeletionFailed,
		Message: "failed to delete user",
		Err:     err,
		Entity:  "user",
	}
}

func NewAllUsersRetrievalError(err error) error {
	return &RepositoryError{
		Code:    AllUsersRetrievalFailed,
		Message: "failed to retrieve all users",
		Err:     err,
		Entity:  "user",
	}
}

func NewUserByEmailRetrievalError(err error) error {
	return &RepositoryError{
		Code:    UserByEmailRetrievalFailed,
		Message: "failed to retrieve user by email",
		Err:     err,
		Entity:  "user",
	}
}

func NewUserByUsernameRetrievalError(err error) error {
	return &RepositoryError{
		Code:    UserByUsernameRetrievalFailed,
		Message: "failed to retrieve user by username",
		Err:     err,
		Entity:  "user",
	}
}

// Сохраняем старые ошибки для обратной совместимости с тестами
var (
	// Ошибки репозитория заметок
	ErrNoteNotFound               = errors.New("заметка не найдена")
	ErrNoteCreationFailed         = errors.New("ошибка создания заметки")
	ErrNoteRetrievalFailed        = errors.New("ошибка получения заметки")
	ErrNoteUpdateFailed           = errors.New("ошибка обновления заметки")
	ErrNoteDeletionFailed         = errors.New("ошибка удаления заметки")
	ErrNotesRetrievalByUserFailed = errors.New("ошибка получения заметок пользователя")
	ErrNotesListRetrievalFailed   = errors.New("ошибка получения списка заметок пользователя")
	ErrAllNotesRetrievalFailed    = errors.New("ошибка получения всех заметок")

	// Ошибки репозитория пользователей
	ErrUserNotFound                  = errors.New("пользователь не найден")
	ErrUserCreationFailed            = errors.New("ошибка создания пользователя")
	ErrUserRetrievalFailed           = errors.New("ошибка получения пользователя")
	ErrUserUpdateFailed              = errors.New("ошибка обновления пользователя")
	ErrUserDeletionFailed            = errors.New("ошибка удаления пользователя")
	ErrAllUsersRetrievalFailed       = errors.New("ошибка получения всех пользователей")
	ErrUserByEmailRetrievalFailed    = errors.New("ошибка получения пользователя по email")
	ErrUserByUsernameRetrievalFailed = errors.New("ошибка получения пользователя по username")
)
