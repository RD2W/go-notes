package main

import (
	"fmt"
	"log"
	"time"

	"github.com/rd2w/go-notes/internal/model"
)

func main() {
	note := model.NewNote("Первая заметка", "Это содержимое моей первой заметки")
	displayNoteInfo("Исходная заметка:", note)
	updateNote(note, "Обновленный заголовок", "Обновленное содержимое")
	displayNoteInfo("После обновления:", note)

	log.Println("Приложение \"Заметки\" успешно завершило выполнение программы!")
}

// displayNoteInfo отображает информацию о заметке в форматированном виде
func displayNoteInfo(header string, note *model.Note) {
	if note == nil {
		fmt.Println("Ошибка: заметка не существует")
		return
	}

	fmt.Println(header)
	fmt.Printf("  Заголовок: %s\n", note.GetTitle())
	fmt.Printf("  Содержимое: %s\n", note.GetContent())
	fmt.Printf("  Создана: %s\n", formatTime(note.GetCreatedAt()))
	fmt.Printf("  Обновлена: %s\n", formatTime(note.GetUpdatedAt()))
	fmt.Println()
}

// updateNote обновляет заголовок и содержимое заметки
func updateNote(note *model.Note, newTitle, newContent string) {
	if note == nil {
		fmt.Println("Ошибка: нельзя обновить несуществующую заметку!")
		return
	}

	fmt.Println("Обновление заметки...")

	if newTitle != "" {
		note.SetTitle(newTitle)
	}

	if newContent != "" {
		note.SetContent(newContent)
	}

	fmt.Printf("✅ Заметка успешно обновлена\n\n")
}

// formatTime форматирует время в едином стиле
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
