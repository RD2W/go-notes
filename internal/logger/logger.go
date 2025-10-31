package logger

import (
	"log"
	"time"

	"github.com/rd2w/go-notes/internal/repository"
)

// Logger отвечает за логирование изменений в данных
type Logger struct {
	repo     *repository.Repository
	done     <-chan struct{}
	interval time.Duration
}

// NewLogger создает новый экземпляр логгера
func NewLogger(repo *repository.Repository, done <-chan struct{}, interval time.Duration) *Logger {
	return &Logger{
		repo:     repo,
		done:     done,
		interval: interval,
	}
}

// Start запускает процесс логирования изменений
func (l *Logger) Start() {
	go func() {
		lastNoteCount := 0
		ticker := time.NewTicker(l.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				currentNoteCount := l.repo.GetNotesCount()

				if currentNoteCount > lastNoteCount {
					newNotes := l.repo.GetNewNotes(lastNoteCount)
					log.Printf("Логгер: обнаружено %d новых заметок", len(newNotes))

					for _, note := range newNotes {
						log.Printf("Логгер: НОВАЯ ЗАМЕТКА - ID: %s, Заголовок: %s, Создана: %s",
							note.GetID(),
							note.GetTitle(),
							note.GetCreatedAt().Format("15:04:05"))
					}

					lastNoteCount = currentNoteCount
				}
			case <-l.done:
				log.Println("Логгер: завершение работы")
				return
			}
		}
	}()
}
