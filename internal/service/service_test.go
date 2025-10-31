package service

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/model"
	"github.com/rd2w/go-notes/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// safeBuffer для service тестов тоже
type safeBuffer struct {
	buf bytes.Buffer
	mu  sync.RWMutex
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buf.String()
}

// TestService_StopWithDoneChannel тестирует остановку генерации через done канал
func TestService_StopWithDoneChannel(t *testing.T) {
	var buf safeBuffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	entityChan := make(chan repository.Entity, 10)
	done := make(chan struct{})

	service := NewService(entityChan, done)

	// Запускаем генерацию
	service.StartDataGeneration(20 * time.Millisecond)

	// Даем время на отправку первой заметки
	time.Sleep(25 * time.Millisecond)

	// Останавливаем сервис
	close(done)

	// Даем время на обработку завершения
	time.Sleep(30 * time.Millisecond)

	// Проверяем что канал закрыт и новых сообщений нет
	select {
	case note := <-entityChan:
		t.Logf("Получена заметка после остановки: %s", note.GetID())
	case <-time.After(50 * time.Millisecond):
		// Ожидаемое поведение - нет новых сообщений
	}

	logOutput := buf.String()

	// Должно быть сообщение о завершении
	assert.Contains(t, logOutput, "Сервис: завершение работы по сигналу",
		"Должно быть сообщение о завершении по сигналу")

	// Проверяем что была отправлена хотя бы одна заметка до остановки
	assert.Contains(t, logOutput, "Сервис: отправлена заметка",
		"Должна быть отправлена хотя бы одна заметка до остановки")
}

// TestService_LimitTenNotes тестирует ограничение в 10 заметок
func TestService_LimitTenNotes(t *testing.T) {
	var buf safeBuffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	entityChan := make(chan repository.Entity, 15) // Буфер больше 10
	done := make(chan struct{})
	defer close(done)

	service := NewService(entityChan, done)

	// Запускаем генерацию с очень коротким интервалом
	service.StartDataGeneration(5 * time.Millisecond)

	// Собираем все заметки
	var notes []repository.Entity
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case note := <-entityChan:
				notes = append(notes, note)
				if len(notes) >= 10 {
					return // Ожидаем максимум 10 заметок
				}
			case <-time.After(200 * time.Millisecond):
				return // Таймаут
			}
		}
	}()

	wg.Wait()

	// Проверяем что было создано ровно 10 заметок
	assert.Len(t, notes, 10, "Должно быть создано ровно 10 заметок")

	// Проверяем логи
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Сервис: отправлена заметка 10")
	assert.Contains(t, logOutput, "Сервис: генерация тестовых данных завершена",
		"Должно быть сообщение о завершении генерации")

	// Проверяем что нет сообщения о 11й заметке
	assert.NotContains(t, logOutput, "Сервис: отправлена заметка 11")
}

// TestService_NoteTitles тестирует корректность заголовков заметок
func TestService_NoteTitles(t *testing.T) {
	entityChan := make(chan repository.Entity, 5)
	done := make(chan struct{})
	defer close(done)

	service := NewService(entityChan, done)

	// Запускаем генерацию
	service.StartDataGeneration(10 * time.Millisecond)

	// Собираем несколько заметок
	var entities []repository.Entity
	for i := 0; i < 3; i++ {
		select {
		case entity := <-entityChan:
			entities = append(entities, entity)
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Таймаут при ожидании заметок")
		}
	}

	// Проверяем заголовки
	expectedTitles := []string{
		"Тестовая заметка 1",
		"Тестовая заметка 2",
		"Тестовая заметка 3",
	}

	for i, entity := range entities {
		// Приводим тип от Entity к *model.Note
		note, ok := entity.(*model.Note)
		require.True(t, ok, "Сущность должна быть типа *model.Note")

		assert.Equal(t, expectedTitles[i], note.GetTitle(),
			"Заголовок заметки %d должен быть '%s'", i+1, expectedTitles[i])
	}
}

// TestService_ConcurrentSafety тестирует безопасность конкурентного доступа
func TestService_ConcurrentSafety(t *testing.T) {
	entityChan := make(chan repository.Entity, 20)
	done := make(chan struct{})
	defer close(done)

	// Создаем несколько сервисов (хотя в реальности это не нужно, тестируем безопасность)
	service1 := NewService(entityChan, done)
	service2 := NewService(entityChan, done)

	// Запускаем оба сервиса
	service1.StartDataGeneration(15 * time.Millisecond)
	service2.StartDataGeneration(15 * time.Millisecond)

	// Собираем заметки
	var noteCount int
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-entityChan:
				noteCount++
				if noteCount >= 10 {
					return
				}
			case <-time.After(200 * time.Millisecond):
				return
			}
		}
	}()

	wg.Wait()

	// Сервисы должны работать без паники
	assert.True(t, noteCount > 0, "Должны быть отправлены заметки")
	t.Logf("Отправлено заметок в конкурентном режиме: %d", noteCount)
}

// TestService_ChannelBlocking тестирует поведение при блокировке канала
func TestService_ChannelBlocking(t *testing.T) {
	var buf safeBuffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	// Создаем канал с очень маленьким буфером
	entityChan := make(chan repository.Entity, 1)
	done := make(chan struct{})

	service := NewService(entityChan, done)

	// Запускаем генерацию с очень коротким интервалом
	service.StartDataGeneration(5 * time.Millisecond)

	// Не читаем из канала, чтобы он быстро заполнился и заблокировался
	time.Sleep(30 * time.Millisecond)

	// Останавливаем сервис
	close(done)
	time.Sleep(20 * time.Millisecond)

	logOutput := buf.String()

	// Сервис должен корректно завершиться по сигналу done
	assert.Contains(t, logOutput, "Сервис: завершение работы по сигналу",
		"Сервис должен корректно завершиться при блокировке канала. Вывод: %s", logOutput)
}

// TestService_ImmediateStop тестирует немедленную остановку сервиса
func TestService_ImmediateStop(t *testing.T) {
	var buf safeBuffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	entityChan := make(chan repository.Entity, 5)
	done := make(chan struct{})

	// Останавливаем сервис сразу же
	close(done)

	service := NewService(entityChan, done)
	service.StartDataGeneration(10 * time.Millisecond)

	// Даем время на обработку
	time.Sleep(20 * time.Millisecond)

	logOutput := buf.String()

	// Должно быть сообщение о завершении
	assert.Contains(t, logOutput, "Сервис: завершение работы по сигналу",
		"Должно быть сообщение о немедленном завершении")

	// Не должно быть отправленных заметок
	assert.NotContains(t, logOutput, "Сервис: отправлена заметка",
		"Не должно быть отправленных заметок при немедленной остановке")
}

// TestService_NoteContent тестирует содержимое заметок
func TestService_NoteContent(t *testing.T) {
	entityChan := make(chan repository.Entity, 3)
	done := make(chan struct{})
	defer close(done)

	service := NewService(entityChan, done)
	service.StartDataGeneration(10 * time.Millisecond)

	// Получаем заметки и проверяем их содержимое
	for i := 1; i <= 3; i++ {
		select {
		case entity := <-entityChan:
			note, ok := entity.(*model.Note)
			require.True(t, ok, "Сущность должна быть заметкой")

			expectedContent := fmt.Sprintf("Это содержимое тестовой заметки номер %d", i)
			assert.Equal(t, expectedContent, note.GetContent(),
				"Содержимое заметки %d должно быть '%s'", i, expectedContent)
		case <-time.After(50 * time.Millisecond):
			t.Fatalf("Таймаут при ожидании заметки %d", i)
		}
	}
}
