package main

import (
	"fmt"
	"log"
	"time"

	"github.com/rd2w/go-notes/internal/logger"
	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
	"github.com/rd2w/go-notes/internal/service"
)

func main() {
	log.Println("Запуск приложения с горутинами и каналами...")
	// Создаем каналы для коммуникации
	entityChan := make(chan repository.Entity, 10)
	done := make(chan struct{})

	// Инициализируем компоненты
	repo := repository.NewRepository()
	svc := service.NewService(entityChan, done)
	newLogger := logger.NewLogger(repo, done, 200*time.Millisecond)

	// Запускаем горутины
	go repo.Save(entityChan, done) // Репозиторий слушает канал
	go newLogger.Start()           // Логгер мониторит изменения

	log.Println("Запуск генерации тестовых данных...")
	svc.StartDataGeneration(500 * time.Millisecond) // Сервис генерирует данные

	// Ждем некоторое время для демонстрации работы
	time.Sleep(6 * time.Second)

	// Сигнал завершения всем горутинам
	close(done)

	// Даем время на корректное завершение
	time.Sleep(100 * time.Millisecond)

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
