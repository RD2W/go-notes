package model

// Note представляет сущность заметки
type Note struct {
	TimeFields
	title   string
	content string
}

// NewNote создает новую заметку с инициализацией временных меток
func NewNote(title, content string) *Note {
	note := &Note{
		title:   title,
		content: content,
	}
	note.initializeTimestamps()
	return note
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
