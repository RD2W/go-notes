package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

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

	// Инициализируем репозиторий
	repo := repository.NewRepositoryByType(repository.JSON)

	// Создаем обработчики
	noteHandler := handler.NewNoteHandler(repo)
	userHandler := handler.NewUserHandler(repo)

	// Создаем группу маршрутов для API
	api := r.Group("/api")
	{
		// Маршруты для аутентификации
		api.POST("/login", userHandler.Login)

		// Защищенные маршруты для заметок
		notes := api.Group("/notes")
		notes.Use(middleware.AuthMiddleware())
		{
			notes.POST("", noteHandler.CreateNote)
			notes.GET("/:id", noteHandler.GetNote)
			notes.PUT("/:id", noteHandler.UpdateNote)
			notes.DELETE("/:id", noteHandler.DeleteNote)
		}

		// Открытый маршрут для получения всех заметок
		api.GET("/notes", noteHandler.GetAllNotes)

		// Защищенные маршруты для пользователей
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware())
		{
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Открытый маршрут для создания пользователя
		api.POST("/users", userHandler.CreateUser)
		// Открытый маршрут для получения всех пользователей
		api.GET("/users", userHandler.GetAllUsers)
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
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка при запуске веб-сервера: %v", err)
		}
	}()

	// Ждем сигнал завершения
	<-sigChan
	log.Println("Получен сигнал завершения, инициируем graceful shutdown...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем сервер с graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при graceful shutdown веб-сервера: %v", err)
	}

	log.Println("Веб-сервер остановлен")
}
