package main

import (
	"fmt"
	"log"
	"time"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
	"github.com/rd2w/go-notes/internal/service"
)

func main() {
	repo := repository.NewRepository()
	svc := service.NewService(repo)

	log.Println("Запуск генерации тестовых данных...")
	svc.StartDataGeneration(1 * time.Second)

	fmt.Printf("\n=== РЕЗУЛЬТАТЫ ===\n")
	fmt.Printf("Всего заметок создано: %d\n", repo.GetNotesCount())

	notes := repo.GetAllNotes()
	for i, note := range notes {
		displayNoteInfo(i, note)
	}

	log.Println("Приложение \"Заметки\" успешно завершило выполнение программы!")
}

// displayNoteInfo отображает информацию о заметке в форматированном виде
func displayNoteInfo(count int, note *model.Note) {
	if note == nil {
		fmt.Println("Ошибка: заметка не существует")
		return
	}

	fmt.Printf("\nЗаметка %d:\n", count+1)
	fmt.Printf("  ID: %s\n", note.GetID())
	fmt.Printf("  Заголовок: %s\n", note.GetTitle())
	fmt.Printf("  Содержимое: %s\n", note.GetContent())
	fmt.Printf("  Создана: %s\n", formatTime(note.GetCreatedAt()))
	fmt.Printf("  Обновлена: %s\n", formatTime(note.GetUpdatedAt()))
	fmt.Println()
}

// formatTime форматирует время в едином стиле
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
