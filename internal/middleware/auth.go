package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rd2w/go-notes/internal/auth"
)

// AuthMiddleware проверяет наличие и валидность JWT токена
func AuthMiddleware(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		log.Printf("Получен заголовок Authorization: %s", authHeader)
		if authHeader == "" {
			log.Printf("Заголовок Authorization отсутствует")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Printf("Заголовок Authorization не содержит префикс Bearer")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token is required"})
			c.Abort()
			return
		}

		log.Printf("Извлечен JWT токен: %s", tokenString)

		claims, err := tokenManager.ValidateAccessToken(tokenString)
		if err != nil {
			log.Printf("Ошибка при проверке JWT токена: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		log.Printf("JWT токен действителен, пользователь: %s", claims.Username)
		c.Set("username", claims.Username)
		c.Set("tokenID", claims.TokenID) // Устанавливаем также TokenID, если нужно
		c.Next()
	}
}
