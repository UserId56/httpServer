package controllers

import (
	"httpServer/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleController struct {
	DB *gorm.DB
}

func NewRoleController(db *gorm.DB) *RoleController {
	return &RoleController{DB: db}
}

func (rc RoleController) RoleGetByID(c *gin.Context) {
	strRoleId := c.Param("id")
	intRoleId, err := strconv.Atoi(strRoleId)
	if err != nil {
		c.JSON(400, gin.H{"error": "Не валидный ID роли"})
		return
	}

	var role models.Role
	result := rc.DB.First(&role, intRoleId)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "Роль не найдена"})
		} else {
			c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		}
		return
	}
	c.JSON(200, role)
}
