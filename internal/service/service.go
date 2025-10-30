package service

import (
	"fmt"
	"log"
	"time"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
)

// Service содержит бизнес-логику приложения
type Service struct {
	repo *repository.Repository
}

// NewService создает новый экземпляр сервиса
func NewService(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// StartDataGeneration запускает периодическое создание тестовых данных
func (s *Service) StartDataGeneration(interval time.Duration) {
	if s.repo == nil {
		log.Println("Ошибка: репозиторий не инициализирован")
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	noteCounter := 1

	for range ticker.C {
		// Создаем новую заметку
		title := fmt.Sprintf("Тестовая заметка %d", noteCounter)
		content := fmt.Sprintf("Это содержимое тестовой заметки номер %d", noteCounter)

		note := model.NewNote(title, content)

		// Передаем в репозиторий
		if err := s.repo.Save(note); err != nil {
			log.Printf("Ошибка сохранения заметки: %v", err)
		} else {
			log.Printf("Сгенерирована заметка: %s", title)
		}

		noteCounter++

		// Останавливаем после создания 5 заметок для демонстрации
		if noteCounter > 5 {
			log.Println("Генерация тестовых данных завершена")
			break
		}
	}
}
