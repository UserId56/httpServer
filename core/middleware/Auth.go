package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/UserId56/httpServer/core/models"
	"github.com/UserId56/httpServer/core/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func Auth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var Authorization string
		Authorization = c.GetHeader("Authorization")
		if Authorization == "" {
			if cookieVal, err := c.Cookie("Authorization"); err == nil && cookieVal != "" {
				Authorization = cookieVal
			}
		}
		if Authorization == "" {
			var roleId float64 = 3
			c.Set("role_id", roleId) // Гость
			var role models.Role
			if err := db.Where("id = ?", roleId).First(&role).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(401, gin.H{"error": "Пользователь не аутентифицирован"})
					c.Abort()
					return
				}
				c.JSON(500, gin.H{"error": "Ошибка на сервере"})
				c.Abort()
				return
			}
			c.Set("permission", role.Permission)
			c.Next()
			return
		}
		claims := jwt.MapClaims{}
		if value, exists := c.Get("claims"); !exists {
			var JWT string
			if strings.HasPrefix(Authorization, "Bearer ") {
				JWT = strings.TrimPrefix(Authorization, "Bearer ")
			} else {
				JWT = Authorization // если токен пришёл без "Bearer "
			}
			token, err := services.ParseJWT(JWT)
			if err != nil || !token.Valid {
				c.JSON(401, gin.H{"error": "Пользователь не аутентифицирован"})
				c.Abort()
				return
			}
			claims = token.Claims.(jwt.MapClaims)
		} else {
			if os.Getenv("DEBUG") == "TRUE" {
				fmt.Printf("Auth: claims: %+v\n", claims)
			}
			claims = value.(jwt.MapClaims)
		}
		userId := claims["user_id"].(float64)
		roleId := claims["role_id"].(float64)
		if os.Getenv("DEBUG") == "TRUE" {
			fmt.Printf("Auth: userId: %v, roleId: %v\n", userId, roleId)
		}
		var userExists int64
		db.Model(&models.User{}).Where("id = ? AND role_id = ?", userId, roleId).Count(&userExists)
		if userExists == 0 {
			c.JSON(401, gin.H{"error": "Пользователь не аутентифицирован"})
			c.Abort()
			return
		}
		var role models.Role
		if err := db.Where("id = ?", uint(roleId)).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(401, gin.H{"error": "Пользователь не аутентифицирован"})
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
