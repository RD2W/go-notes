package webserver

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rd2w/go-notes/internal/app/lifecycle"
	tokenauth "github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/database"
	httpdelivery "github.com/rd2w/go-notes/internal/delivery/http"
	"github.com/rd2w/go-notes/internal/repository/postgres"
	authservice "github.com/rd2w/go-notes/internal/service/auth"
	"github.com/rd2w/go-notes/internal/service/note"
	"github.com/rd2w/go-notes/internal/service/user"
)

// WebServer структура веб-сервера
type WebServer struct {
	config       *config.Config
	server       *http.Server
	dbClient     *database.PostgresClient
	tokenManager *tokenauth.TokenManager
	redisClient  *database.RedisClient
}

// NewWebServer создает новый экземпляр веб-сервера
func NewWebServer(cfg *config.Config) *WebServer {
	// Устанавливаем Gin в нужный режим
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Создаем Gin роутер
	r := gin.Default()

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
	noteService := note.NewNoteService(noteRepo)
	userService := user.NewUserService(userRepo)
	authService := authservice.NewAuthService(tokenManager, userService)

	// Создаем HTTP-хендлеры
	noteHandler := httpdelivery.NewNoteHandler(noteService)
	userHandler := httpdelivery.NewUserHandler(userService)
	authHandler := httpdelivery.NewAuthHandler(authService)

	// Настраиваем маршруты
	SetupRoutes(r, noteHandler, userHandler, authHandler, tokenManager)

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: r,
	}

	webServer := &WebServer{
		config:       cfg,
		server:       srv,
		dbClient:     postgresClient,
		tokenManager: tokenManager,
		redisClient:  redisClient,
	}

	return webServer
}

// Run запускает веб-сервер с обработкой сигналов завершения
func (ws *WebServer) Run() error {
	defer func() {
		if ws.dbClient != nil {
			ws.dbClient.Close()
			log.Println("Соединение с базой данных закрыто")
		}
		if ws.redisClient != nil {
			if err := ws.redisClient.Close(); err != nil {
				log.Printf("Ошибка при закрытии Redis соединения: %v", err)
			} else {
				log.Println("Соединение с Redis закрыто")
			}
		}
	}()

	return lifecycle.StartHTTPServer(ws.config, ws.server)
}

// GetServer возвращает http.Server
func (ws *WebServer) GetServer() *http.Server {
	return ws.server
}

// GetConfig возвращает конфигурацию сервера
func (ws *WebServer) GetConfig() *config.Config {
	return ws.config
}
