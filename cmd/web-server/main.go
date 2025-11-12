package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/config"
	"github.com/rd2w/go-notes/internal/handler"
	"github.com/rd2w/go-notes/internal/middleware"
	"github.com/rd2w/go-notes/internal/repository"
	"github.com/rd2w/go-notes/internal/repository/storage/fs"
	"github.com/rd2w/go-notes/internal/repository/storage/ram"

	_ "github.com/rd2w/go-notes/docs"
)

// @title Go Notes API
// @version 1.0
// @description API для управления заметками и пользователями
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme
func main() {
	// Устанавливаем Gin в режиме release
	gin.SetMode(gin.ReleaseMode)

	// Создаем Gin роутер
	r := gin.Default()

	// Регистрируем реализации репозитория
	repository.Register(repository.RAM, ram.NewRamRepository)
	repository.Register(repository.JSON, fs.NewJSONRepository)

	// Загружаем конфигурацию из файла (предполагаем, что config_dev.toml находится в /config корне проекта)
	cfg, err := config.LoadConfig("./config/config_dev.toml")
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить конфигурацию из config_dev.toml: %v", err)
		log.Println("Используем конфигурацию по умолчанию")
		cfg = config.NewDefaultConfigWithValues()
	}

	// Создаем токен-менеджер
	tokenManager := auth.NewTokenManager(cfg)

	// Инициализируем репозиторий
	repo := repository.NewRepositoryByType(repository.JSON)

	// Создаем обработчики
	noteHandler := handler.NewNoteHandler(repo)
	userHandler := handler.NewUserHandler(repo)
	authHandler := handler.NewAuthHandler(repo, tokenManager)

	// Создаем группу маршрутов для API
	api := r.Group("/api")
	{
		// Маршруты для аутентификации
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/validate", authHandler.ValidateToken)
		}

		// Открытый маршрут для получения всех заметок
		api.GET("/notes", noteHandler.GetAllNotes)

		// Защищенные маршруты для заметок
		notes := api.Group("/notes")
		notes.Use(middleware.AuthMiddleware(tokenManager))
		{
			notes.POST("", noteHandler.CreateNote)
			notes.GET("/:id", noteHandler.GetNote)
			notes.PUT("/:id", noteHandler.UpdateNote)
			notes.DELETE("/:id", noteHandler.DeleteNote)
		}

		// Открытые маршруты для пользователей
		api.POST("/users", userHandler.CreateUser)
		api.GET("/users", userHandler.GetAllUsers)

		// Защищенные маршруты для пользователей
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(tokenManager))
		{
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	// Добавляем маршрут для Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Канал для получения сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Веб-сервер запущен на порту %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка при запуске веб-сервера: %v", err)
		}
	}()

	// Ждем сигнал завершения
	<-sigChan
	log.Println("Получен сигнал завершения, инициируем graceful shutdown...")

	// Создаем контекст с таймаутом для graceful shutdown
	shutdownTimeout, err := time.ParseDuration(cfg.Shutdown.Timeout)
	if err != nil {
		log.Printf("Ошибка при парсинге таймаута graceful shutdown: %v, используем значение по умолчанию 5s", err)
		shutdownTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Останавливаем сервер с graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при graceful shutdown веб-сервера: %v", err)
	}

	log.Println("Веб-сервер остановлен")
}
