package controllers

import (
	"github.com/UserId56/httpServer/core/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SettingsController struct {
	DB *gorm.DB
}

func NewSettingsController(db *gorm.DB) *SettingsController {
	return &SettingsController{
		DB: db,
	}
}

func (sc *SettingsController) SettingsGet(c *gin.Context) {
	var settings models.Settings
	if err := sc.DB.Where("ID = ?", 1).First(&settings).Error; err != nil {
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.JSON(200, settings.Value)
}

func (sc *SettingsController) SettingsUpdate(c *gin.Context) {
	var newSettings models.SettingsValue
	if err := c.ShouldBindJSON(&newSettings); err != nil {
		c.JSON(400, gin.H{"error": "Не валидный JSON"})
		return
	}

	settings := models.Settings{
		ID:    1,
		Value: newSettings,
	}

	if err := sc.DB.Model(&models.Settings{}).Where("ID = ?", 1).Updates(settings).Error; err != nil {
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	c.Status(200)
}

func (sc *SettingsController) SettingsGetStyle(c *gin.Context) {
	var settings models.Settings
	if err := sc.DB.Where("ID = ?", 1).First(&settings).Error; err != nil {
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.JSON(200, gin.H{"style": settings.Value.Style})
}
