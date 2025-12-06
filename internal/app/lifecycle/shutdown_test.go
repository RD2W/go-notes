package lifecycle

import (
	"context"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	"google.golang.org/grpc"
)

func TestWaitForShutdownSignal(t *testing.T) {
	// Тестирование ожидания сигнала завершения
	done := make(chan bool, 1)

	// Запускаем ожидание сигнала в отдельной горутине
	go func() {
		WaitForShutdownSignal()
		done <- true
	}()

	// Отправляем сигнал завершения
	go func() {
		time.Sleep(100 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
	}()

	// Ждем получения сигнала
	select {
	case <-done:
		// Успешно дождались сигнала
	case <-time.After(2 * time.Second):
		t.Error("Таймаут ожидания сигнала завершения")
	}
}

func TestShutdownHTTPServer(t *testing.T) {
	// Создаем конфигурацию по умолчанию
	cfg := &config.Config{
		Shutdown: config.ShutdownConfig{
			Timeout: "5s",
		},
	}

	// Создаем HTTP сервер
	server := &http.Server{
		Addr: ":0", // Используем случайный порт
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}),
	}

	// Запускаем сервер в отдельной горутине
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatalf("Не удалось создать слушатель: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		_ = server.Serve(listener)
	}()

	// Ждем немного, чтобы сервер успел запуститься
	time.Sleep(50 * time.Millisecond)

	// Вызываем graceful shutdown
	ShutdownHTTPServer(cfg, server)
}

func TestShutdownHTTPServerWithInvalidTimeout(t *testing.T) {
	// Создаем конфигурацию с невалидным таймаутом
	cfg := &config.Config{
		Shutdown: config.ShutdownConfig{
			Timeout: "invalid", // Невалидное значение
		},
	}

	// Создаем HTTP сервер
	server := &http.Server{
		Addr: ":0", // Используем случайный порт
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}),
	}

	// Запускаем сервер в отдельной горутине
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatalf("Не удалось создать слушатель: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		_ = server.Serve(listener)
	}()

	// Ждем немного, чтобы сервер успел запуститься
	time.Sleep(50 * time.Millisecond)

	// Вызываем graceful shutdown - должно использоваться значение по умолчанию
	ShutdownHTTPServer(cfg, server)
}

func TestShutdownHTTPServerWithCustomTimeout(t *testing.T) {
	// Создаем конфигурацию с коротким таймаутом
	cfg := &config.Config{
		Shutdown: config.ShutdownConfig{
			Timeout: "100ms", // Короткий таймаут
		},
	}

	// Создаем HTTP сервер
	server := &http.Server{
		Addr: ":0", // Используем случайный порт
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}),
	}

	// Запускаем сервер в отдельной горутине
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatalf("Не удалось создать слушатель: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		_ = server.Serve(listener)
	}()

	// Ждем немного, чтобы сервер успел запуститься
	time.Sleep(50 * time.Millisecond)

	// Вызываем graceful shutdown
	ShutdownHTTPServer(cfg, server)
}

func TestShutdownGRPCServer(t *testing.T) {
	// Создаем gRPC сервер
	grpcServer := grpc.NewServer()

	// Создаем слушатель на случайном порту
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Не удалось создать слушатель: %v", err)
	}
	defer func() { _ = lis.Close() }()

	// Запускаем сервер в отдельной горутине
	go func() {
		_ = grpcServer.Serve(lis)
	}()

	// Ждем немного, чтобы сервер успел запуститься
	time.Sleep(50 * time.Millisecond)

	// Вызываем graceful shutdown
	ShutdownGRPCServer(grpcServer)
}

// Тестирование ситуации, когда сервер уже остановлен
func TestShutdownHTTPServerAlreadyClosed(t *testing.T) {
	// Создаем конфигурацию по умолчанию
	cfg := &config.Config{
		Shutdown: config.ShutdownConfig{
			Timeout: "5s",
		},
	}

	// Создаем HTTP сервер
	server := &http.Server{
		Addr: ":0", // Используем случайный порт
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}),
	}

	// Создаем контекст с таймаутом для закрытия сервера
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Закрываем сервер до вызова ShutdownHTTPServer
	_ = server.Shutdown(ctx)

	// Пытаемся вызвать graceful shutdown - не должно быть паники
	ShutdownHTTPServer(cfg, server)
}
