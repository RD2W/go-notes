package model

import (
	"testing"
	"time"
)

func TestNewNote(t *testing.T) {
	title := "Test Title"
	content := "Test Content"

	note := NewNote(title, content)

	if note == nil {
		t.Fatal("NewNote returned nil")
	}

	// Проверяем установку полей
	if note.GetTitle() != title {
		t.Errorf("Expected title %q, got %q", title, note.GetTitle())
	}

	if note.GetContent() != content {
		t.Errorf("Expected content %q, got %q", content, note.GetContent())
	}

	// Проверяем инициализацию временных меток
	if note.GetCreatedAt().IsZero() {
		t.Error("CreatedAt should be initialized")
	}

	if note.GetUpdatedAt().IsZero() {
		t.Error("UpdatedAt should be initialized")
	}

	// Проверяем, что created и updated равны при создании
	if !note.GetCreatedAt().Equal(note.GetUpdatedAt()) {
		t.Error("CreatedAt and UpdatedAt should be equal for new note")
	}

	// Проверяем, что временные метки близки к текущему времени
	now := time.Now()
	createdAt := note.GetCreatedAt()

	if createdAt.After(now) {
		t.Error("CreatedAt should not be in the future")
	}

	// Допускаем небольшую погрешность во времени выполнения
	if now.Sub(createdAt) > time.Second {
		t.Error("CreatedAt should be close to current time")
	}
}

func TestNoteSetters(t *testing.T) {
	note := NewNote("Initial Title", "Initial Content")
	initialCreatedAt := note.GetCreatedAt()
	initialUpdatedAt := note.GetUpdatedAt()

	// Даем небольшое время для обеспечения разных временных меток
	time.Sleep(10 * time.Millisecond)

	// Тестируем SetTitle
	newTitle := "Updated Title"
	note.SetTitle(newTitle)

	if note.GetTitle() != newTitle {
		t.Errorf("Expected title %q after SetTitle, got %q", newTitle, note.GetTitle())
	}

	// Проверяем, что updatedAt изменился
	if note.GetUpdatedAt().Equal(initialUpdatedAt) {
		t.Error("UpdatedAt should change after SetTitle")
	}

	// Проверяем, что createdAt не изменился
	if !note.GetCreatedAt().Equal(initialCreatedAt) {
		t.Error("CreatedAt should not change after SetTitle")
	}

	// Проверяем, что updatedAt стал позже
	if note.GetUpdatedAt().Before(initialUpdatedAt) {
		t.Error("UpdatedAt should be after previous UpdatedAt")
	}

	// Сохраняем updatedAt после первого изменения
	updatedAfterTitle := note.GetUpdatedAt()
	time.Sleep(10 * time.Millisecond)

	// Тестируем SetContent
	newContent := "Updated Content"
	note.SetContent(newContent)

	if note.GetContent() != newContent {
		t.Errorf("Expected content %q after SetContent, got %q", newContent, note.GetContent())
	}

	// Проверяем, что updatedAt снова изменился
	if note.GetUpdatedAt().Equal(updatedAfterTitle) {
		t.Error("UpdatedAt should change after SetContent")
	}

	// Проверяем, что createdAt все еще не изменился
	if !note.GetCreatedAt().Equal(initialCreatedAt) {
		t.Error("CreatedAt should not change after SetContent")
	}
}

func TestNoteGetters(t *testing.T) {
	title := "Getter Test Title"
	content := "Getter Test Content"

	note := NewNote(title, content)

	// Тестируем геттеры
	if got := note.GetTitle(); got != title {
		t.Errorf("GetTitle() = %q, want %q", got, title)
	}

	if got := note.GetContent(); got != content {
		t.Errorf("GetContent() = %q, want %q", got, content)
	}

	// Проверяем, что геттеры временных меток возвращают не-zero значения
	if note.GetCreatedAt().IsZero() {
		t.Error("GetCreatedAt() returned zero time")
	}

	if note.GetUpdatedAt().IsZero() {
		t.Error("GetUpdatedAt() returned zero time")
	}
}

func TestTimeFieldsMethods(t *testing.T) {
	tf := &TimeFields{}

	// До инициализации временные метки должны быть zero
	if !tf.GetCreatedAt().IsZero() {
		t.Error("CreatedAt should be zero before initialization")
	}

	if !tf.GetUpdatedAt().IsZero() {
		t.Error("UpdatedAt should be zero before initialization")
	}

	// Инициализируем временные метки
	tf.initializeTimestamps()

	// После инициализации не должны быть zero
	if tf.GetCreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero after initialization")
	}

	if tf.GetUpdatedAt().IsZero() {
		t.Error("UpdatedAt should not be zero after initialization")
	}

	// Сохраняем текущие значения
	initialCreatedAt := tf.GetCreatedAt()
	initialUpdatedAt := tf.GetUpdatedAt()

	time.Sleep(10 * time.Millisecond)

	// Обновляем временную метку
	tf.updateTimestamp()

	// Проверяем, что updatedAt изменился, а createdAt остался прежним
	if tf.GetCreatedAt() != initialCreatedAt {
		t.Error("CreatedAt should not change after updateTimestamp")
	}

	if tf.GetUpdatedAt().Equal(initialUpdatedAt) {
		t.Error("UpdatedAt should change after updateTimestamp")
	}

	if tf.GetUpdatedAt().Before(initialUpdatedAt) {
		t.Error("UpdatedAt should be after previous value")
	}
}

func TestNoteEdgeCases(t *testing.T) {
	// Тестируем пустые значения
	emptyNote := NewNote("", "")
	if emptyNote.GetTitle() != "" {
		t.Error("Should handle empty title")
	}
	if emptyNote.GetContent() != "" {
		t.Error("Should handle empty content")
	}

	// Тестируем установку пустых значений
	note := NewNote("Title", "Content")
	note.SetTitle("")
	if note.GetTitle() != "" {
		t.Error("SetTitle should handle empty string")
	}

	note.SetContent("")
	if note.GetContent() != "" {
		t.Error("SetContent should handle empty string")
	}
}

func TestMultipleNotes(t *testing.T) {
	// Создаем несколько заметок и проверяем их независимость
	note1 := NewNote("Note 1", "Content 1")
	note2 := NewNote("Note 2", "Content 2")

	// Проверяем, что у них разные временные метки
	// (могут быть равны если созданы в одну наносекунду, но это маловероятно)
	if note1.GetCreatedAt().Equal(note2.GetCreatedAt()) {
		t.Log("Note: Both notes have same CreatedAt (very close creation time)")
	}

	// Проверяем независимость данных
	if note1.GetTitle() == note2.GetTitle() {
		t.Error("Notes should have different titles")
	}

	if note1.GetContent() == note2.GetContent() {
		t.Error("Notes should have different content")
	}

	// Изменяем одну заметку и проверяем, что другая не изменилась
	note1OriginalTitle := note1.GetTitle()
	note2.SetTitle("Modified Note 2")

	if note1.GetTitle() != note1OriginalTitle {
		t.Error("Modifying one note should not affect another")
	}
}
