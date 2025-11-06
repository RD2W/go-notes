package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rd2w/go-notes/internal/logger"
	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
	"github.com/rd2w/go-notes/internal/service"
)

// Константы приложения
const (
	LoggerInterval        = 200 * time.Millisecond
	DataGenInterval       = 500 * time.Millisecond
	GracefulShutdownDelay = 100 * time.Millisecond

	TimeFormat = "2006-01-02 15:04:05"
)

// Строковые константы
const (
	AppStartMsg         = "Запуск приложения с горутинами и каналами..."
	AppShutdownMsg      = "Приложение \"Заметки\" успешно завершило выполнение программы!"
	ShutdownStartMsg    = "Получен сигнал завершения, инициируем graceful shutdown..."
	ResultsHeader       = "\n=== РЕЗУЛЬТАТЫ ===\n"
	NoteCountMsg        = "Всего заметок создано: %d\n"
	NoteDoesNotExistMsg = "Ошибка: заметка не существует"
	NoteHeaderMsg       = "Заметка %d:\n"
	NoteIDMsg           = "  ID: %s\n"
	NoteTitleMsg        = "  Заголовок: %s\n"
	NoteContentMsg      = "  Содержимое: %s\n"
	NoteCreatedAtMsg    = "  Создана: %s\n"
	NoteUpdatedAtMsg    = "  Обновлена: %s\n"
)

func main() {
	log.Println(AppStartMsg)

	// Создаем контекст с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Обработка сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Инициализируем компоненты
	repo := repository.NewRepository()
	svc := service.NewService(repo, ctx, DataGenInterval)
	newLogger := logger.NewLogger(repo, ctx, LoggerInterval)

	// Запускаем горутины
	go newLogger.Start() // Логгер мониторит изменения
	go svc.Start()       // Сервис запускает генерацию и сохранение данных

	// Ждем сигнал завершения
	<-sigChan
	log.Println(ShutdownStartMsg)

	// Отменяем контекст для завершения всех горутин
	cancel()

	// Даем время на корректное завершение
	time.Sleep(GracefulShutdownDelay)

	fmt.Print(ResultsHeader)
	fmt.Printf(NoteCountMsg, repo.GetNotesCount())

	notes := repo.GetAllNotes()
	for i, note := range notes {
		displayNoteInfo(i, note)
	}

	log.Println(AppShutdownMsg)
}

// displayNoteInfo отображает информацию о заметке в форматированном виде
func displayNoteInfo(count int, note *model.Note) {
	if note == nil {
		fmt.Println(NoteDoesNotExistMsg)
		return
	}

	fmt.Printf(NoteHeaderMsg, count+1)
	fmt.Printf(NoteIDMsg, note.GetID())
	fmt.Printf(NoteTitleMsg, note.GetTitle())
	fmt.Printf(NoteContentMsg, note.GetContent())
	fmt.Printf(NoteCreatedAtMsg, formatTime(note.GetCreatedAt()))
	fmt.Printf(NoteUpdatedAtMsg, formatTime(note.GetUpdatedAt()))
	fmt.Println()
}

// formatTime форматирует время в едином стиле
func formatTime(t time.Time) string {
	return t.Format(TimeFormat)
}
