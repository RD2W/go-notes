package lifecycle

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rd2w/go-notes/internal/config"
	"google.golang.org/grpc"
)

// StartHTTPServer запускает HTTP сервер с обработкой сигналов завершения
func StartHTTPServer(cfg *config.Config, server *http.Server) error {
	// Канал для получения сигнала завершения
	sigChan := make(chan struct{})

	// Канал для получения ошибки при запуске сервера
	errChan := make(chan error, 1)

	// Канал для получения системного сигнала
	sysSigChan := make(chan os.Signal, 1)
	signal.Notify(sysSigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Веб-сервер запущен на порту %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
			return
		}
		close(sigChan)
	}()

	// Ждем сигнал завершения или ошибку при запуске
	select {
	case err := <-errChan:
		return err
	case <-sigChan:
		// Сервер завершил работу по другим причинам
		return nil
	case <-sysSigChan:
		// Получен сигнал завершения
		log.Println("Получен сигнал завершения, инициируем graceful shutdown...")

		// Выполняем graceful shutdown
		ShutdownHTTPServer(cfg, server)

		// Ждем завершения работы сервера
		<-sigChan
		log.Println("Веб-сервер остановлен")
	}

	return nil
}

// StartGRPCServer запускает gRPC сервер с обработкой сигналов завершения
func StartGRPCServer(cfg *config.Config, server *grpc.Server, lis net.Listener) error {
	// Канал для получения сигнала завершения
	sigChan := make(chan struct{})

	// Канал для получения ошибки при запуске сервера
	errChan := make(chan error, 1)

	// Канал для получения системного сигнала
	sysSigChan := make(chan os.Signal, 1)
	signal.Notify(sysSigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем gRPC сервер в отдельной горутине
	go func() {
		log.Printf("gRPC сервер запущен на порту %s", lis.Addr().String())
		if err := server.Serve(lis); err != nil {
			errChan <- err
			return
		}
		close(sigChan)
	}()

	// Ждем сигнал завершения или ошибку при запуске
	select {
	case err := <-errChan:
		return err
	case <-sigChan:
		// Сервер завершил работу по другим причинам
		return nil
	case <-sysSigChan:
		// Получен сигнал завершения
		log.Println("Получен сигнал завершения, инициируем graceful shutdown...")

		// Выполняем graceful shutdown
		ShutdownGRPCServer(server)

		// Ждем завершения работы сервера
		<-sigChan
		log.Println("gRPC сервер остановлен")
	}

	return nil
}
