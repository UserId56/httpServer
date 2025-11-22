package middleware

import (
	"httpServer/models"
	"httpServer/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func Auth(db *gorm.DB) gin.HandlerFunc {
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
		userId := claims["user_id"].(float64)
		roleId := claims["role_id"].(float64)
		var userExists int64
		db.Model(&models.User{}).Where("id = ? AND role_id = ?", userId, roleId).Count(&userExists)
		if userExists == 0 {
			c.JSON(401, gin.H{"error": "Неавторизованный пользователь"})
			c.Abort()
			return
		}
		c.Set("user_id", claims["user_id"])
		c.Set("role_id", claims["role_id"])
		c.Next()
	}
}
