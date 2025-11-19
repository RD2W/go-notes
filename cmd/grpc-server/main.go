package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	grpcServer "github.com/rd2w/go-notes/internal/grpc"
	"github.com/rd2w/go-notes/internal/repository"
	"github.com/rd2w/go-notes/internal/repository/storage/fs"
	"github.com/rd2w/go-notes/internal/repository/storage/ram"
	authpb "github.com/rd2w/go-notes/pkg/proto/auth"
	"github.com/rd2w/go-notes/pkg/proto/note"
	"github.com/rd2w/go-notes/pkg/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Обработка сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Регистрируем реализации репозитория
	repository.Register(repository.RAM, ram.NewRamRepository)
	repository.Register(repository.JSON, fs.NewJSONRepository)

	// Инициализируем компоненты
	repo := repository.NewRepositoryByType(repository.JSON)

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config/config_dev.toml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем TokenManager
	tokenManager := auth.NewTokenManager(cfg)

	// Создаем наш gRPC сервер
	grpcService := grpcServer.NewServer(repo, tokenManager)

	// Создаем сетевой слушатель
	lis, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("Ошибка при создании сетевого слушателя: %v", err)
	}

	// Создаем gRPC сервер с помощью библиотеки
	grpcServerLib := grpc.NewServer()

	// Регистрируем gRPC сервис
	note.RegisterNotesServiceServer(grpcServerLib, grpcService)
	user.RegisterUserServiceServer(grpcServerLib, grpcService)
	authpb.RegisterAuthServiceServer(grpcServerLib, grpcService)

	// Добавляем reflection для инструментов gRPC
	reflection.Register(grpcServerLib)

	// Запускаем gRPC сервер в отдельной горутине
	go func() {
		log.Printf("gRPC сервер запущен на порту %s", cfg.Server.GRPCPort)
		if err := grpcServerLib.Serve(lis); err != nil {
			log.Fatalf("Ошибка при запуске gRPC сервера: %v", err)
		}
	}()

	// Ждем сигнал завершения
	<-sigChan
	log.Println("Получен сигнал завершения, инициируем graceful shutdown...")

	// Останавливаем gRPC сервер
	grpcServerLib.GracefulStop()
	log.Println("gRPC сервер остановлен")
}
