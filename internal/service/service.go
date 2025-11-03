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
	repo     *repository.Repository
	done     <-chan struct{}
	interval time.Duration
}

// NewService создает новый экземпляр сервиса
func NewService(repo *repository.Repository, done <-chan struct{}, interval time.Duration) *Service {
	return &Service{
		repo:     repo,
		done:     done,
		interval: interval,
	}
}

// Start запускает все горутины сервиса
func (s *Service) Start() {
	// Создаем канал для передачи сущностей между горутинами
	entityChan := make(chan repository.Entity, 10)

	// Запускаем горутину для генерации данных
	go s.startDataGeneration(entityChan)

	// Запускаем горутину для сохранения данных
	go s.startDataSaving(entityChan)
}

// startDataGeneration запускает периодическое создание тестовых данных
func (s *Service) startDataGeneration(entityChan chan<- repository.Entity) {
	go func() {
		noteCounter := 1
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				title := fmt.Sprintf("Тестовая заметка %d", noteCounter)
				content := fmt.Sprintf("Это содержимое тестовой заметки номер %d", noteCounter)

				note := model.NewNote(title, content)

				select {
				case entityChan <- note:
					log.Printf("Сервис: создана заметка %d", noteCounter)
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
				log.Println("Сервис: завершение работы генерации по сигналу")
				return
			}
		}
	}()
}

// startDataSaving запускает сохранение данных в репозиторий
func (s *Service) startDataSaving(entityChan <-chan repository.Entity) {
	for {
		select {
		case entity := <-entityChan:
			// Вызываем синхронный метод сохранения в репозитории
			s.repo.Save(entity)
			log.Printf("Сервис: сохранена сущность %s", entity.GetID())
		case <-s.done:
			log.Println("Сервис: завершение сохранения данных")
			return
		}
	}
}
