package errors

import "errors"

// Ошибки репозитория заметок
var (
	ErrNoteNotFound               = errors.New("заметка не найдена")
	ErrNoteCreationFailed         = errors.New("ошибка создания заметки")
	ErrNoteRetrievalFailed        = errors.New("ошибка получения заметки")
	ErrNoteUpdateFailed           = errors.New("ошибка обновления заметки")
	ErrNoteDeletionFailed         = errors.New("ошибка удаления заметки")
	ErrNotesRetrievalByUserFailed = errors.New("ошибка получения заметок пользователя")
	ErrNotesListRetrievalFailed   = errors.New("ошибка получения списка заметок пользователя")
	ErrAllNotesRetrievalFailed    = errors.New("ошибка получения всех заметок")
)

// Ошибки репозитория пользователей
var (
	ErrUserNotFound                  = errors.New("пользователь не найден")
	ErrUserCreationFailed            = errors.New("ошибка создания пользователя")
	ErrUserRetrievalFailed           = errors.New("ошибка получения пользователя")
	ErrUserUpdateFailed              = errors.New("ошибка обновления пользователя")
	ErrUserDeletionFailed            = errors.New("ошибка удаления пользователя")
	ErrAllUsersRetrievalFailed       = errors.New("ошибка получения всех пользователей")
	ErrUserByEmailRetrievalFailed    = errors.New("ошибка получения пользователя по email")
	ErrUserByUsernameRetrievalFailed = errors.New("ошибка получения пользователя по username")
)
