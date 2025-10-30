package repository

import (
	"testing"

	"github.com/rd2w/go-notes/internal/model"
)

// MockEntity для тестирования неподдерживаемых сущностей
type MockEntity struct{}

func (m *MockEntity) GetID() string   { return "mock-id" }
func (m *MockEntity) GetType() string { return "mock-type" }

func TestNewRepository(t *testing.T) {
	repo := NewRepository()

	if repo == nil {
		t.Fatal("NewRepository returned nil")
	}

	if repo.notes == nil {
		t.Error("Notes slice should be initialized")
	}

	if len(repo.notes) != 0 {
		t.Errorf("New repository should have 0 notes, got %d", len(repo.notes))
	}
}

func TestSaveNote(t *testing.T) {
	repo := NewRepository()
	note := model.NewNote("Test Title", "Test Content")

	// Сохраняем заметку
	err := repo.Save(note)
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Проверяем что заметка сохранилась
	if repo.GetNotesCount() != 1 {
		t.Errorf("Expected 1 note, got %d", repo.GetNotesCount())
	}

	// Проверяем что это именно та заметка
	notes := repo.GetAllNotes()
	if len(notes) != 1 {
		t.Fatalf("Expected 1 note in GetAllNotes, got %d", len(notes))
	}

	if notes[0].GetID() != note.GetID() {
		t.Error("Saved note ID doesn't match")
	}

	if notes[0].GetTitle() != note.GetTitle() {
		t.Error("Saved note title doesn't match")
	}
}

func TestSaveMultipleNotes(t *testing.T) {
	repo := NewRepository()

	// Создаем и сохраняем несколько заметок
	notes := []*model.Note{
		model.NewNote("Note 1", "Content 1"),
		model.NewNote("Note 2", "Content 2"),
		model.NewNote("Note 3", "Content 3"),
	}

	for i, note := range notes {
		err := repo.Save(note)
		if err != nil {
			t.Errorf("Failed to save note %d: %v", i, err)
		}
	}

	// Проверяем количество
	if repo.GetNotesCount() != 3 {
		t.Errorf("Expected 3 notes, got %d", repo.GetNotesCount())
	}

	// Проверяем что все заметки сохранились
	allNotes := repo.GetAllNotes()
	if len(allNotes) != 3 {
		t.Fatalf("Expected 3 notes in GetAllNotes, got %d", len(allNotes))
	}

	// Проверяем целостность данных
	for i, savedNote := range allNotes {
		if savedNote.GetID() != notes[i].GetID() {
			t.Errorf("Note %d ID mismatch", i)
		}
		if savedNote.GetTitle() != notes[i].GetTitle() {
			t.Errorf("Note %d title mismatch", i)
		}
	}
}

func TestSaveUnsupportedEntity(t *testing.T) {
	repo := NewRepository()
	mockEntity := &MockEntity{}

	// Пытаемся сохранить неподдерживаемую сущность
	err := repo.Save(mockEntity)
	if err == nil {
		t.Error("Expected error for unsupported entity type")
	}

	expectedError := "неподдерживаемый тип сущности: *repository.MockEntity"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}

	// Проверяем что ничего не сохранилось
	if repo.GetNotesCount() != 0 {
		t.Errorf("Repository should be empty after failed save, got %d notes", repo.GetNotesCount())
	}
}

func TestGetAllNotes(t *testing.T) {
	repo := NewRepository()

	// Проверяем пустой репозиторий
	emptyNotes := repo.GetAllNotes()
	if len(emptyNotes) != 0 {
		t.Errorf("GetAllNotes should return empty slice for new repository, got %d", len(emptyNotes))
	}

	// Добавляем заметки и проверяем
	note1 := model.NewNote("Note 1", "Content 1")
	note2 := model.NewNote("Note 2", "Content 2")

	// Обрабатываем ошибки при сохранении
	if err := repo.Save(note1); err != nil {
		t.Fatalf("Failed to save note1: %v", err)
	}
	if err := repo.Save(note2); err != nil {
		t.Fatalf("Failed to save note2: %v", err)
	}

	allNotes := repo.GetAllNotes()
	if len(allNotes) != 2 {
		t.Fatalf("Expected 2 notes, got %d", len(allNotes))
	}
}

func TestGetNotesCount(t *testing.T) {
	repo := NewRepository()

	// Проверяем начальное состояние
	if count := repo.GetNotesCount(); count != 0 {
		t.Errorf("New repository should have 0 notes, got %d", count)
	}

	// Добавляем заметки и проверяем счетчик
	if err := repo.Save(model.NewNote("Note 1", "Content 1")); err != nil {
		t.Fatalf("Failed to save note 1: %v", err)
	}
	if count := repo.GetNotesCount(); count != 1 {
		t.Errorf("Expected 1 note, got %d", count)
	}

	if err := repo.Save(model.NewNote("Note 2", "Content 2")); err != nil {
		t.Fatalf("Failed to save note 2: %v", err)
	}
	if count := repo.GetNotesCount(); count != 2 {
		t.Errorf("Expected 2 notes, got %d", count)
	}

	if err := repo.Save(model.NewNote("Note 3", "Content 3")); err != nil {
		t.Fatalf("Failed to save note 3: %v", err)
	}
	if count := repo.GetNotesCount(); count != 3 {
		t.Errorf("Expected 3 notes, got %d", count)
	}
}

func TestRepositoryIsolation(t *testing.T) {
	// Проверяем что разные репозитории изолированы друг от друга
	repo1 := NewRepository()
	repo2 := NewRepository()

	note1 := model.NewNote("Repo1 Note", "Content")
	note2 := model.NewNote("Repo2 Note", "Content")

	// Обрабатываем ошибки при сохранении
	if err := repo1.Save(note1); err != nil {
		t.Fatalf("Failed to save note1 in repo1: %v", err)
	}
	if err := repo2.Save(note2); err != nil {
		t.Fatalf("Failed to save note2 in repo2: %v", err)
	}

	// Проверяем изоляцию
	if repo1.GetNotesCount() != 1 {
		t.Errorf("Repo1 should have 1 note, got %d", repo1.GetNotesCount())
	}
	if repo2.GetNotesCount() != 1 {
		t.Errorf("Repo2 should have 1 note, got %d", repo2.GetNotesCount())
	}

	repo1Notes := repo1.GetAllNotes()
	repo2Notes := repo2.GetAllNotes()

	if repo1Notes[0].GetID() != note1.GetID() {
		t.Error("Repo1 contains wrong note")
	}
	if repo2Notes[0].GetID() != note2.GetID() {
		t.Error("Repo2 contains wrong note")
	}
}

func TestSaveNilEntity(t *testing.T) {
	repo := NewRepository()

	// Пытаемся сохранить nil
	err := repo.Save(nil)
	if err == nil {
		t.Error("Expected error when saving nil entity")
	}

	expectedError := "неподдерживаемый тип сущности: <nil>"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}
