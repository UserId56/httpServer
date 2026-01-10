package middleware

import (
	"strings"

	"github.com/UserId56/httpServer/core/models"
	"github.com/UserId56/httpServer/core/services"

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
		var role models.Role
		if err := db.Where("id = ?", uint(roleId)).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(401, gin.H{"error": "Неавторизованный пользователь"})
				c.Abort()
				return
			}
			c.JSON(500, gin.H{"error": "Ошибка на сервере"})
			c.Abort()
			return
		}
		c.Set("permission", role.Permission)
		c.Set("user_id", claims["user_id"])
		c.Set("role_id", claims["role_id"])
		c.Next()
	}
}
