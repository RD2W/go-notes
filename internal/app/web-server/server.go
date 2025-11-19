package webserver

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rd2w/go-notes/internal/app/lifecycle"
	tokenauth "github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	httpdelivery "github.com/rd2w/go-notes/internal/delivery/http"
	"github.com/rd2w/go-notes/internal/repository/file"
	authservice "github.com/rd2w/go-notes/internal/service/auth"
	"github.com/rd2w/go-notes/internal/service/note"
	"github.com/rd2w/go-notes/internal/service/user"
)

// WebServer структура веб-сервера
type WebServer struct {
	config *config.Config
	server *http.Server
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

	// Создаем токен-менеджер
	tokenManager := tokenauth.NewTokenManager(cfg)

	// Инициализируем репозиторий
	repo, err := file.NewFileRepository("./data/notes.json")
	if err != nil {
		log.Fatalf("Ошибка инициализации репозитория: %v", err)
	}

	// Создаем бизнес-сервисы
	noteService := note.NewNoteService(repo)
	userService := user.NewUserService(repo)
	authService := authservice.NewAuthService(repo, tokenManager, userService)

	// Создаем HTTP-хендлеры
	noteHandler := httpdelivery.NewNoteHandler(noteService)
	userHandler := httpdelivery.NewUserHandler(userService)
	authHandler := httpdelivery.NewAuthHandler(authService)

	// Настраиваем маршруты
	SetupRoutes(r, noteHandler, userHandler, authHandler)

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: r,
	}

	return &WebServer{
		config: cfg,
		server: srv,
	}
}

// Run запускает веб-сервер с обработкой сигналов завершения
func (ws *WebServer) Run() error {
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
