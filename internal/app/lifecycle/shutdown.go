package lifecycle

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rd2w/go-notes/internal/config"
	"google.golang.org/grpc"
)

// WaitForShutdownSignal ожидает сигнал завершения программы
func WaitForShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

// ShutdownHTTPServer останавливает HTTP сервер с graceful shutdown
func ShutdownHTTPServer(cfg *config.Config, server *http.Server) {
	shutdownTimeout, err := time.ParseDuration(cfg.Shutdown.Timeout)
	if err != nil {
		log.Printf("Ошибка при парсинге таймаута graceful shutdown: %v, используем значение по умолчанию 5s", err)
		shutdownTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при graceful shutdown веб-сервера: %v", err)
	}
}

// ShutdownGRPCServer останавливает gRPC сервер с graceful shutdown
func ShutdownGRPCServer(server *grpc.Server) {
	server.GracefulStop()
}
