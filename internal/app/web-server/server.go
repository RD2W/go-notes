package webserver

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rd2w/go-notes/internal/app/lifecycle"
	"github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/repository"
	"github.com/rd2w/go-notes/internal/repository/storage/fs"
	"github.com/rd2w/go-notes/internal/repository/storage/ram"
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

	// Регистрируем реализации репозитория
	repository.Register(repository.RAM, ram.NewRamRepository)
	repository.Register(repository.JSON, fs.NewJSONRepository)

	// Создаем токен-менеджер
	tokenManager := auth.NewTokenManager(cfg)

	// Инициализируем репозиторий
	repo := repository.NewRepositoryByType(repository.JSON)

	// Настраиваем маршруты
	SetupRoutes(r, repo, tokenManager)

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
