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
	entityChan chan<- repository.Entity
	done       chan struct{}
}

// NewService создает новый экземпляр сервиса
func NewService(entityChan chan<- repository.Entity, done chan struct{}) *Service {
	return &Service{
		entityChan: entityChan,
		done:       done,
	}
}

// StartDataGeneration запускает периодическое создание тестовых данных
func (s *Service) StartDataGeneration(interval time.Duration) {
	go func() {
		noteCounter := 1
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				title := fmt.Sprintf("Тестовая заметка %d", noteCounter)
				content := fmt.Sprintf("Это содержимое тестовой заметки номер %d", noteCounter)

				note := model.NewNote(title, content)

				select {
				case s.entityChan <- note:
					log.Printf("Сервис: отправлена заметка %d", noteCounter)
				case <-s.done:
					log.Println("Сервис: завершение генерации данных")
					return
				}

				noteCounter++
				if noteCounter > 10 {
					log.Println("Сервис: генерация тестовых данных завершена")
					return
				}
			case <-s.done:
				log.Println("Сервис: завершение работы по сигналу")
				return
			}
		}
	}()
}
