package model

import (
	"encoding/json"
	"time"

	"github.com/rd2w/go-notes/internal/util"
)

// Note представляет сущность заметки
type Note struct {
	TimeFields
	id      string
	title   string
	content string
	userId  string // ID пользователя, которому принадлежит заметка
}

// NewNote создает новую заметку с инициализацией временных меток
func NewNote(title, content, userId string) *Note {
	note := &Note{
		id:      util.GenerateID(),
		title:   title,
		content: content,
		userId:  userId,
	}
	note.initializeTimestamps()
	return note
}

// GetID возвращает идентификатор заметки (реализация интерфейса Entity)
func (n *Note) GetID() string {
	return n.id
}

// SetID устанавливает идентификатор заметки
func (n *Note) SetID(id string) { n.id = id }

// SetCreatedAt устанавливает время создания
func (n *Note) SetCreatedAt(createdAt time.Time) { n.createdAt = createdAt }

// SetUpdatedAt устанавливает время обновления
func (n *Note) SetUpdatedAt(updatedAt time.Time) { n.updatedAt = updatedAt }

// GetType возвращает тип сущности (реализация интерфейса Entity)
func (n *Note) GetType() string {
	return "note"
}

// GetTitle возвращает заголовок заметки
func (n *Note) GetTitle() string { return n.title }

// GetContent возвращает содержимое заметки
func (n *Note) GetContent() string {
	return n.content
}

// GetUserID возвращает ID пользователя, которому принадлежит заметка
func (n *Note) GetUserID() string {
	return n.userId
}

// SetUserID устанавливает ID пользователя, которому принадлежит заметка
func (n *Note) SetUserID(userId string) { n.userId = userId }

// SetTitle устанавливает новый заголовок и обновляет временную метку
func (n *Note) SetTitle(newTitle string) {
	n.title = newTitle
	n.updateTimestamp()
}

// SetContent устанавливает новое содержимое и обновляет временную метку
func (n *Note) SetContent(newContent string) {
	n.content = newContent
	n.updateTimestamp()
}

// JSONNote вспомогательная структура для JSON сериализации
type JSONNote struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MarshalJSON реализует интерфейс json.Marshaler
func (n *Note) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONNote{
		ID:        n.id,
		Title:     n.title,
		Content:   n.content,
		UserID:    n.userId,
		CreatedAt: n.createdAt,
		UpdatedAt: n.updatedAt,
	})
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler
func (n *Note) UnmarshalJSON(data []byte) error {
	var jsonNote JSONNote
	if err := json.Unmarshal(data, &jsonNote); err != nil {
		return err
	}

	n.id = jsonNote.ID
	n.title = jsonNote.Title
	n.content = jsonNote.Content
	n.userId = jsonNote.UserID
	n.createdAt = jsonNote.CreatedAt
	n.updatedAt = jsonNote.UpdatedAt

	return nil
}
