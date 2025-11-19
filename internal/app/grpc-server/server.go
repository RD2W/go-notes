package grpcserver

import (
	"log"
	"net"

	"github.com/rd2w/go-notes/internal/app/lifecycle"
	tokenauth "github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	grpcdelivery "github.com/rd2w/go-notes/internal/delivery/grpc"
	"github.com/rd2w/go-notes/internal/repository/file"
	authservice "github.com/rd2w/go-notes/internal/service/auth"
	servicenote "github.com/rd2w/go-notes/internal/service/note"
	serviceuser "github.com/rd2w/go-notes/internal/service/user"
	authpb "github.com/rd2w/go-notes/pkg/proto/auth"
	grpcnote "github.com/rd2w/go-notes/pkg/proto/note"
	grpcuser "github.com/rd2w/go-notes/pkg/proto/user"
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
	// Создаем токен-менеджер
	tokenManager := tokenauth.NewTokenManager(cfg)

	// Инициализируем репозиторий
	repo, err := file.NewFileRepository("./data/notes.json")
	if err != nil {
		log.Fatalf("Ошибка инициализации репозитория: %v", err)
	}

	// Создаем бизнес-сервисы
	noteService := servicenote.NewNoteService(repo)
	userService := serviceuser.NewUserService(repo)
	authService := authservice.NewAuthService(repo, tokenManager, userService)

	// Создаем gRPC-серверы для каждого сервиса
	noteServer := grpcdelivery.NewNoteServiceServer(noteService)
	userServer := grpcdelivery.NewUserServiceServer(userService)
	authServer := grpcdelivery.NewAuthServiceServer(authService)

	// Создаем сетевой слушатель
	lis, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("Ошибка при создании сетевого слушателя: %v", err)
	}

	// Создаем gRPC сервер с помощью библиотеки
	grpcServerLib := grpc.NewServer()

	// Регистрируем gRPC сервисы
	grpcnote.RegisterNotesServiceServer(grpcServerLib, noteServer)
	grpcuser.RegisterUserServiceServer(grpcServerLib, userServer)
	authpb.RegisterAuthServiceServer(grpcServerLib, authServer)

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
