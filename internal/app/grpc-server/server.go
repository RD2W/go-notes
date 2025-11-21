package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/rd2w/go-notes/internal/app/lifecycle"
	tokenauth "github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/database"
	grpcdelivery "github.com/rd2w/go-notes/internal/delivery/grpc"
	"github.com/rd2w/go-notes/internal/repository/postgres"
	authservice "github.com/rd2w/go-notes/internal/service/auth"
	servicenote "github.com/rd2w/go-notes/internal/service/note"
	serviceuser "github.com/rd2w/go-notes/internal/service/user"
	authpb "github.com/rd2w/go-notes/pkg/proto/auth"
	grpcnote "github.com/rd2w/go-notes/pkg/proto/note"
	grpcuser "github.com/rd2w/go-notes/pkg/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// GRPCServer структура gRPC-сервера
type GRPCServer struct {
	config       *config.Config
	grpcServer   *grpc.Server
	listener     net.Listener
	dbClient     *database.PostgresClient
	tokenManager *tokenauth.TokenManager
	redisClient  *database.RedisClient
}

// NewGRPCServer создает новый экземпляр gRPC-сервера
func NewGRPCServer(cfg *config.Config) *GRPCServer {
	// Создаем Redis клиент
	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}

	// Создаем токен-менеджер с переданным Redis клиентом
	tokenManager := tokenauth.NewTokenManager(cfg, redisClient)

	// Создаем клиент подключения к PostgreSQL
	postgresClient, err := database.NewPostgresClient(cfg.Postgres)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}

	// Создаем сетевой слушатель
	lis, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("Ошибка при создании сетевого слушателя: %v", err)
	}

	// Создаем interceptor для проверки токенов
	authInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Пропускаем проверку для методов аутентификации и создания пользователя
		if info.FullMethod == "/auth.AuthService/Login" ||
			info.FullMethod == "/auth.AuthService/Logout" ||
			info.FullMethod == "/auth.AuthService/Refresh" ||
			info.FullMethod == "/auth.AuthService/ValidateToken" ||
			info.FullMethod == "/users.UserService/CreateUser" {
			return handler(ctx, req)
		}

		// Извлекаем токен из контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("metadata is not provided")
		}

		authHeaders := md["authorization"]
		if len(authHeaders) == 0 {
			return nil, fmt.Errorf("authorization token is not provided")
		}

		tokenString := strings.TrimPrefix(authHeaders[0], "Bearer ")
		if tokenString == authHeaders[0] {
			return nil, fmt.Errorf("authorization token is not in Bearer format")
		}

		// Проверяем токен на валидность и наличие в черном списке
		_, err := tokenManager.ValidateAccessToken(tokenString)
		if err != nil {
			return nil, fmt.Errorf("invalid or blacklisted token: %w", err)
		}

		// Токен валиден, продолжаем обработку запроса
		return handler(ctx, req)
	}

	// Создаем gRPC сервер с помощью библиотеки с interceptor
	grpcServerLib := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))

	// Инициализируем репозитории
	noteRepo, err := postgres.NewPostgresNoteRepository(postgresClient.GetPool())
	if err != nil {
		log.Fatalf("Ошибка при создании репозитория заметок: %v", err)
	}

	userRepo, err := postgres.NewPostgresUserRepository(postgresClient.GetPool())
	if err != nil {
		log.Fatalf("Ошибка при создании репозитория пользователей: %v", err)
	}

	// Создаем бизнес-сервисы
	noteService := servicenote.NewNoteService(noteRepo)
	userService := serviceuser.NewUserService(userRepo)
	authService := authservice.NewAuthService(tokenManager, userService)

	// Создаем gRPC-серверы для каждого сервиса
	noteServer := grpcdelivery.NewNoteServiceServer(noteService)
	userServer := grpcdelivery.NewUserServiceServer(userService)
	authServer := grpcdelivery.NewAuthServiceServer(authService)

	// Регистрируем gRPC сервисы
	grpcnote.RegisterNotesServiceServer(grpcServerLib, noteServer)
	grpcuser.RegisterUserServiceServer(grpcServerLib, userServer)
	authpb.RegisterAuthServiceServer(grpcServerLib, authServer)

	// Добавляем reflection для инструментов gRPC
	reflection.Register(grpcServerLib)

	// Создаем экземпляр сервера
	grpcServer := &GRPCServer{
		config:       cfg,
		grpcServer:   grpcServerLib,
		listener:     lis,
		dbClient:     postgresClient,
		tokenManager: tokenManager,
		redisClient:  redisClient,
	}

	return grpcServer
}

// Run запускает gRPC-сервер с обработкой сигналов завершения
func (gs *GRPCServer) Run() error {
	defer func() {
		if gs.dbClient != nil {
			gs.dbClient.Close()
			log.Println("Соединение с базой данных закрыто")
		}
		if gs.redisClient != nil {
			if err := gs.redisClient.Close(); err != nil {
				log.Printf("Ошибка при закрытии Redis соединения: %v", err)
			} else {
				log.Println("Соединение с Redis закрыто")
			}
		}
	}()

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
