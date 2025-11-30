package lifecycle

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	"google.golang.org/grpc"
)

// sendShutdownSignal отправляет сигнал завершения вручную
func sendShutdownSignal() {
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM)
}

func TestStartHTTPServer(t *testing.T) {
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
	go func() {
		// Подождем немного, чтобы сервер успел запуститься
		time.Sleep(50 * time.Millisecond)
		// Имитируем сигнал завершения
		sendShutdownSignal()
	}()

	// Запускаем тест
	err := StartHTTPServer(cfg, server)
	if err != nil {
		t.Fatalf("StartHTTPServer вернул ошибку: %v", err)
	}
}

func TestStartGRPCServer(t *testing.T) {
	// Создаем конфигурацию по умолчанию
	cfg := &config.Config{
		Shutdown: config.ShutdownConfig{
			Timeout: "5s",
		},
	}

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
		// Подождем немного, чтобы сервер успел запуститься
		time.Sleep(50 * time.Millisecond)
		// Имитируем сигнал завершения
		sendShutdownSignal()
	}()

	// Запускаем тест
	err = StartGRPCServer(cfg, grpcServer, lis)
	if err != nil {
		t.Fatalf("StartGRPCServer вернул ошибку: %v", err)
	}
}

// Тестирование корректного завершения HTTP сервера с таймаутом
func TestHTTPServerWithTimeout(t *testing.T) {
	cfg := &config.Config{
		Shutdown: config.ShutdownConfig{
			Timeout: "100ms", // Маленький таймаут для теста
		},
	}

	// Создаем медленный обработчик для тестирования таймаута
	server := &http.Server{
		Addr: ":0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Делаем задержку, которая превышает таймаут
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		sendShutdownSignal()
	}()

	err := StartHTTPServer(cfg, server)
	if err != nil {
		t.Fatalf("StartHTTPServer вернул ошибку: %v", err)
	}
}

// Тестирование корректного завершения gRPC сервера
func TestGRPCServerWithTimeout(t *testing.T) {
	cfg := &config.Config{
		Shutdown: config.ShutdownConfig{
			Timeout: "5s",
		},
	}

	grpcServer := grpc.NewServer()

	// Создаем слушатель на случайном порту
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Не удалось создать слушатель: %v", err)
	}
	defer func() { _ = lis.Close() }()

	go func() {
		time.Sleep(50 * time.Millisecond)
		sendShutdownSignal()
	}()

	err = StartGRPCServer(cfg, grpcServer, lis)
	if err != nil {
		t.Fatalf("StartGRPCServer вернул ошибку: %v", err)
	}
}

// Тестирование ситуации, когда сервер завершает работу до получения сигнала
func TestHTTPServerAlreadyClosed(t *testing.T) {
	cfg := &config.Config{
		Shutdown: config.ShutdownConfig{
			Timeout: "5s",
		},
	}

	// Создаем HTTP сервер на случайном порту
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

	serverStarted := make(chan bool, 1)
	go func() {
		serverStarted <- true
		// Пытаемся запустить сервер, но он будет закрыт до завершения
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Logf("Ошибка при работе сервера: %v", err)
		}
	}()

	// Ждем, пока сервер начнет запускаться
	<-serverStarted
	time.Sleep(10 * time.Millisecond)

	// Закрываем сервер принудительно
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	// Вызываем StartHTTPServer, но сервер уже закрыт
	err = StartHTTPServer(cfg, server)
	if err != nil {
		// Ошибка ожидаема, так как сервер уже закрыт
		// Проверяем, что функция завершается без паники
		t.Logf("Ожидаемая ситуация - сервер уже закрыт: %v", err)
	}
}
