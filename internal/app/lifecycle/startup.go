package lifecycle

import (
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/rd2w/go-notes/internal/config"
	"google.golang.org/grpc"
)

// StartHTTPServer запускает HTTP сервер с обработкой сигналов завершения
func StartHTTPServer(cfg *config.Config, server *http.Server) error {
	// Канал для получения сигнала завершения
	sigChan := make(chan struct{})

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Веб-сервер запущен на порту %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка при запуске веб-сервера: %v", err)
		}
		close(sigChan)
	}()

	// Ждем сигнал завершения
	WaitForShutdownSignal()

	log.Println("Получен сигнал завершения, инициируем graceful shutdown...")

	// Выполняем graceful shutdown
	ShutdownHTTPServer(cfg, server)

	// Ждем завершения работы сервера
	<-sigChan
	log.Println("Веб-сервер остановлен")

	return nil
}

// StartGRPCServer запускает gRPC сервер с обработкой сигналов завершения
func StartGRPCServer(cfg *config.Config, server *grpc.Server, lis net.Listener) error {
	// Канал для получения сигнала завершения
	sigChan := make(chan struct{})

	// Запускаем gRPC сервер в отдельной горутине
	go func() {
		log.Printf("gRPC сервер запущен на порту %s", lis.Addr().String())
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Ошибка при запуске gRPC сервера: %v", err)
		}
		close(sigChan)
	}()

	// Ждем сигнал завершения
	WaitForShutdownSignal()

	log.Println("Получен сигнал завершения, инициируем graceful shutdown...")

	// Выполняем graceful shutdown
	ShutdownGRPCServer(server)

	// Ждем завершения работы сервера
	<-sigChan
	log.Println("gRPC сервер остановлен")

	return nil
}
