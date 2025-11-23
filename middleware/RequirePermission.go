package middleware

import (
	"errors"
	"httpServer/models"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequirePermission(db *gorm.DB, permission []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleId, exist := c.Get("role_id")
		if !exist {
			c.JSON(403, gin.H{"error": "Роль не указана"})
			c.Abort()
			return
		}
		roleId := int(userRoleId.(float64))
		if roleId == 1 {
			c.Next()
			return
		}
		var role models.Role
		if err := db.Model(&models.Role{}).Where("id = ?", roleId).First(&role).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(403, gin.H{"error": "Роль не найдена"})
			c.Abort()
			return
		}
		valid := true
		for _, perm := range permission {
			valid = valid && slices.Contains(role.Permission, perm)
		}
		permissionString := strings.Join(permission, ", ")
		if !valid {
			c.JSON(403, gin.H{"error": "Необходимы права: " + permissionString + " или права Администратора"})
			c.Abort()
			return
		}
		c.Next()
	}

}
