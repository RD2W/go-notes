package model

import (
	"crypto/rand"
	"encoding/base64"
)

// Note представляет сущность заметки
type Note struct {
	TimeFields
	id      string
	title   string
	content string
}

// NewNote создает новую заметку с инициализацией временных меток
func NewNote(title, content string) *Note {
	note := &Note{
		id:      generateID(),
		title:   title,
		content: content,
	}
	note.initializeTimestamps()
	return note
}

// GetID возвращает идентификатор заметки (реализация интерфейса Entity)
func (n *Note) GetID() string {
	return n.id
}

// GetType возвращает тип сущности (реализация интерфейса Entity)
func (n *Note) GetType() string {
	return "note"
}

// GetTitle возвращает заголовок заметки
func (n *Note) GetTitle() string {
	return n.title
}

// GetContent возвращает содержимое заметки
func (n *Note) GetContent() string {
	return n.content
}

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

// generateID генерирует уникальный идентификатор
func generateID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic("не удалось сгенерировать ID")
	}
	return base64.URLEncoding.EncodeToString(b)
}
