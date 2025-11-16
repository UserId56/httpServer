package controllers

import (
	"errors"
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

func (rc RoleController) RoleCreate(c *gin.Context) {
	var roleInput models.CreateRoleRequest
	if err := c.ShouldBindJSON(&roleInput); err != nil {
		c.JSON(400, gin.H{"error": "Не валидный JSON или не валидные поля"})
		return
	}

	role := models.Role{
		Model:      gorm.Model{},
		Name:       roleInput.Name,
		Permission: roleInput.Permission,
	}

	if err := rc.DB.Create(&role).Error; err != nil {
		if errors.As(err, &gorm.ErrDuplicatedKey) {
			c.JSON(400, gin.H{"error": "Роль с таким именем уже существует"})
			return
		}
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.JSON(200, role)
}
