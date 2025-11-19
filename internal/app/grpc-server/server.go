package grpcserver

import (
	"log"
	"net"

	"github.com/rd2w/go-notes/internal/app/lifecycle"
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

// GRPCServer структура gRPC-сервера
type GRPCServer struct {
	config     *config.Config
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewGRPCServer создает новый экземпляр gRPC-сервера
func NewGRPCServer(cfg *config.Config) *GRPCServer {
	// Регистрируем реализации репозитория
	repository.Register(repository.RAM, ram.NewRamRepository)
	repository.Register(repository.JSON, fs.NewJSONRepository)

	// Инициализируем компоненты
	repo := repository.NewRepositoryByType(repository.JSON)

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

	return &GRPCServer{
		config:     cfg,
		grpcServer: grpcServerLib,
		listener:   lis,
	}
}

// Run запускает gRPC-сервер с обработкой сигналов завершения
func (gs *GRPCServer) Run() error {
	return lifecycle.StartGRPCServer(gs.config, gs.grpcServer, gs.listener)
}

// GetGRPCServer возвращает gRPC сервер
func (gs *GRPCServer) GetGRPCServer() *grpc.Server {
	return gs.grpcServer
}

// GetConfig возвращает конфигурацию сервера
func (gs *GRPCServer) GetConfig() *config.Config {
	return gs.config
}

// GetListener возвращает сетевой слушатель
func (gs *GRPCServer) GetListener() net.Listener {
	return gs.listener
}

// Stop останавливает gRPC-сервер
func (gs *GRPCServer) Stop() {
	gs.grpcServer.GracefulStop()
}
