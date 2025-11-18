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

func (rc RoleController) RoleUpdateByID(c *gin.Context) {
	strRoleId := c.Param("id")
	intRoleId, err := strconv.Atoi(strRoleId)
	if err != nil {
		c.JSON(400, gin.H{"error": "Не валидный ID роли"})
		return
	}

	var roleInput models.CreateRoleRequest
	if err := c.ShouldBindJSON(&roleInput); err != nil {
		c.JSON(400, gin.H{"error": "Не валидный JSON или не валидные поля"})
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

	role.Name = roleInput.Name
	role.Permission = roleInput.Permission
	if err := rc.DB.Save(&role).Error; err != nil {
		if errors.As(err, &gorm.ErrDuplicatedKey) {
			c.JSON(400, gin.H{"error": "Роль с таким именем уже существует"})
			return
		}
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.Status(201)
}

func (rc RoleController) RoleDeleteByID(c *gin.Context) {
	strRoleId := c.Param("id")
	intRoleId, err := strconv.Atoi(strRoleId)
	if err != nil {
		c.JSON(400, gin.H{"error": "Не валидный ID роли"})
		return
	}

	var role models.Role
	if err := rc.DB.Model(&role).Unscoped().Where("id = ?", intRoleId).First(&role).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Роль не найдена"})
		return
	}
	if !role.DeletedAt.Time.IsZero() {
		if err := rc.DB.Unscoped().Delete(&models.Role{}, intRoleId).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Роль не найдена"})
			return
		}
		c.JSON(201, gin.H{"message": "Роль успешно удалена окончательно"})
		return
	}
	if err := rc.DB.Delete(&models.Role{}, intRoleId).Error; err != nil {
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.JSON(201, gin.H{"message": "Роль успешно удалена"})
}

func (rc RoleController) RoleQuery(c *gin.Context) {
	take := c.DefaultQuery("take", "10")
	skip := c.DefaultQuery("skip", "0")

	takeInt, err := strconv.Atoi(take)
	if err != nil || takeInt <= 0 {
		c.JSON(400, gin.H{"error": "Параметр 'take' должен быть положительным целым числом"})
		return
	}
	skipInt, err := strconv.Atoi(skip)
	if err != nil || skipInt < 0 {
		c.JSON(400, gin.H{"error": "Параметр 'skip' должен быть неотрицательным целым числом"})
		return
	}
	var count int64
	var roles []models.Role
	err = rc.DB.Model(&models.Role{}).Count(&count).Error
	if count == 0 {
		c.Header("X-Total-Count", "0")
		c.JSON(200, gin.H{"roles": []models.Role{}})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	err = rc.DB.Model(&models.Role{}).Limit(takeInt).Offset(skipInt).Find(&roles).Error
	if err != nil {
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(count, 10))
	c.JSON(200, gin.H{"roles": roles})
}
