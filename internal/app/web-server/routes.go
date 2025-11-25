package webserver

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	tokenauth "github.com/rd2w/go-notes/internal/auth"
	"github.com/rd2w/go-notes/internal/delivery/http"
	"github.com/rd2w/go-notes/internal/middleware"
)

// SetupRoutes настраивает все маршруты для веб-сервера
func SetupRoutes(r *gin.Engine, noteHandler *http.NoteHandler, userHandler *http.UserHandler, authHandler *http.AuthHandler, tokenManager *tokenauth.TokenManager) {
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
}
