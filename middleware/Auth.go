package middleware

import (
	"httpServer/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		Authorization := c.GetHeader("Authorization")
		if Authorization == "" {
			c.JSON(401, gin.H{"error": "Нет токена авторизации"})
			c.Abort()
			return
		}
		JWT := strings.Split(Authorization, "Bearer ")[1]
		token, err := services.ParseJWT(JWT)
		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Неавторизованный пользователь"})
			c.Abort()
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])
		c.Set("role_id", claims["role_id"])
		c.Next()
	}
}
