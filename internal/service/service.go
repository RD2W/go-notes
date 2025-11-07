package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
)

const (
	EntityChanBuffer = 10
	MaxNotesCount    = 10
)

// Service содержит бизнес-логику приложения
type Service struct {
	repo     repository.Repository
	ctx      context.Context
	interval time.Duration
}

// NewService создает новый экземпляр сервиса
func NewService(repo repository.Repository, ctx context.Context, interval time.Duration) *Service {
	return &Service{
		repo:     repo,
		ctx:      ctx,
		interval: interval,
	}
}

// Start запускает все горутины сервиса
func (s *Service) Start() {
	// Создаем канал для передачи сущностей между горутинами
	entityChan := make(chan repository.Entity, EntityChanBuffer)

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
				case <-s.ctx.Done():
					log.Println("Сервис: завершение генерации данных")
					return
				}

				noteCounter++
				if noteCounter > MaxNotesCount {
					log.Println("Сервис: генерация тестовых данных завершена")
					return
				}
			case <-s.ctx.Done():
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
		case <-s.ctx.Done():
			log.Println("Сервис: завершение сохранения данных")
			return
		}
	}
}
