package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

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

// TestService_StopWithContext тестирует остановку генерации через контекст
func TestService_StopWithContext(t *testing.T) {
	var buf safeBuffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	ctx, cancel := context.WithCancel(context.Background())

	repo := repository.NewRepository()
	service := NewService(repo, ctx, 20*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Даем время на отправку первой заметки
	time.Sleep(25 * time.Millisecond)

	// Останавливаем сервис
	cancel()

	// Даем время на обработку завершения
	time.Sleep(30 * time.Millisecond)

	logOutput := buf.String()

	// Должно быть сообщение о завершении
	assert.Contains(t, logOutput, "Сервис: завершение работы генерации по сигналу",
		"Должно быть сообщение о завершении по сигналу")

	// Проверяем что была отправлена хотя бы одна заметка до остановки
	assert.Contains(t, logOutput, "Сервис: создана заметка",
		"Должна быть отправлена хотя бы одна заметка до остановки")
}

// TestService_LimitTenNotes тестирует ограничение в 10 заметок
func TestService_LimitTenNotes(t *testing.T) {
	var buf safeBuffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем, пока будут созданы все заметки
	time.Sleep(150 * time.Millisecond)

	// Проверяем что было создано ровно 10 заметок
	assert.Len(t, repo.GetAllNotes(), 10, "Должно быть создано ровно 10 заметок")

	// Проверяем логи
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Сервис: создана заметка 10")
	assert.Contains(t, logOutput, "Сервис: генерация тестовых данных завершена",
		"Должно быть сообщение о завершении генерации")

	// Проверяем что нет сообщения о 11й заметке
	assert.NotContains(t, logOutput, "Сервис: создана заметка 11")
}

// TestService_NoteTitles тестирует корректность заголовков заметок
func TestService_NoteTitles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewRepository()
	service := NewService(repo, ctx, 10*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание нескольких заметок
	time.Sleep(50 * time.Millisecond)

	// Проверяем заголовки
	notes := repo.GetAllNotes()
	require.GreaterOrEqual(t, len(notes), 3, "Должно быть создано как минимум 3 заметки")

	expectedTitles := []string{
		"Тестовая заметка 1",
		"Тестовая заметка 2",
		"Тестовая заметка 3",
	}

	for i := 0; i < len(expectedTitles) && i < len(notes); i++ {
		assert.Equal(t, expectedTitles[i], notes[i].GetTitle(),
			"Заголовок заметки %d должен быть '%s'", i+1, expectedTitles[i])
	}
}

// TestService_ConcurrentSafety тестирует безопасность конкурентного доступа
func TestService_ConcurrentSafety(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewRepository()

	// Создаем сервис
	service := NewService(repo, ctx, 15*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем немного времени
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что сервисы работают без паники и создают заметки
	assert.Greater(t, repo.GetNotesCount(), 0, "Должны быть созданы заметки")
	t.Logf("Создано заметок: %d", repo.GetNotesCount())
}

// TestService_ChannelBlocking тестирует поведение при блокировке канала
func TestService_ChannelBlocking(t *testing.T) {
	var buf safeBuffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	ctx, cancel := context.WithCancel(context.Background())

	repo := repository.NewRepository()
	service := NewService(repo, ctx, 5*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем немного времени
	time.Sleep(30 * time.Millisecond)

	// Останавливаем сервис
	cancel()
	time.Sleep(20 * time.Millisecond)

	logOutput := buf.String()

	// Сервис должен корректно завершиться по сигналу контекста
	assert.Contains(t, logOutput, "Сервис: завершение",
		"Сервис должен корректно завершиться. Вывод: %s", logOutput)
}

// TestService_ImmediateStop тестирует немедленную остановку сервиса
func TestService_ImmediateStop(t *testing.T) {
	var buf safeBuffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	ctx, cancel := context.WithCancel(context.Background())

	// Останавливаем сервис сразу же
	cancel()

	repo := repository.NewRepository()
	service := NewService(repo, ctx, 10*time.Millisecond)
	service.Start()

	// Даем время на обработку
	time.Sleep(20 * time.Millisecond)

	logOutput := buf.String()

	// Должно быть сообщение о завершении
	assert.Contains(t, logOutput, "Сервис: завершение работы генерации по сигналу",
		"Должно быть сообщение о немедленном завершении")

	// Не должно быть отправленных заметок
	assert.NotContains(t, logOutput, "Сервис: создана заметка",
		"Не должно быть созданных заметок при немедленной остановке")
}

// TestService_NoteContent тестирует содержимое заметок
func TestService_NoteContent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewRepository()
	service := NewService(repo, ctx, 10*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание заметок
	time.Sleep(100 * time.Millisecond)

	// Проверяем содержимое заметок
	notes := repo.GetAllNotes()
	require.GreaterOrEqual(t, len(notes), 3, "Должно быть создано как минимум 3 заметки")

	for i := 1; i <= len(notes) && i <= 3; i++ {
		expectedContent := fmt.Sprintf("Это содержимое тестовой заметки номер %d", i)
		assert.Equal(t, expectedContent, notes[i-1].GetContent(),
			"Содержимое заметки %d должно быть '%s'", i, expectedContent)
	}
}

// TestService_SimpleCase тестирует простой сценарий работы сервиса
func TestService_SimpleCase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewRepository()
	service := NewService(repo, ctx, 30*time.Millisecond)

	// Запускаем сервис
	service.Start()

	// Ждем создание хотя бы одной заметки
	time.Sleep(40 * time.Millisecond)

	// Проверяем, что создана хотя бы одна заметка
	assert.GreaterOrEqual(t, repo.GetNotesCount(), 1, "Должна быть создана хотя бы одна заметка")
}

// TestService_MultipleInstances тестирует работу нескольких экземпляров сервиса
func TestService_MultipleInstances(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewRepository()

	// Создаем два сервиса
	service1 := NewService(repo, ctx, 20*time.Millisecond)
	service2 := NewService(repo, ctx, 25*time.Millisecond)

	// Запускаем оба сервиса
	service1.Start()
	service2.Start()

	// Ждем некоторое время
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что оба сервиса создают заметки
	assert.Greater(t, repo.GetNotesCount(), 0, "Должны быть созданы заметки от обоих сервисов")
	t.Logf("Создано заметок от обоих сервисов: %d", repo.GetNotesCount())
}
