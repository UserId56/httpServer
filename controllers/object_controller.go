package controllers

import (
	"httpServer/logger"
	"httpServer/models"
	"httpServer/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ObjectController struct {
	DB *gorm.DB
}

func NewObjectController(db *gorm.DB) *ObjectController {
	return &ObjectController{
		DB: db,
	}
}

func (o *ObjectController) ObjectCreate(c *gin.Context) {
	objectName := c.Param("object")
	var obj map[string]interface{}
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var table models.DynamicScheme
	if err := o.DB.Where("name = ?", objectName).First(&table).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "Таблица не найдена"})
			return
		}
		logger.Log(err, "Ошибка получения схемы таблицы", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка получения схемы таблицы: " + err.Error()})
		return
	}
	var fields []models.DynamicColumns
	if err := o.DB.Model(&models.DynamicColumns{}).Where("dynamic_table_id = ?", table.ID).Find(&fields).Error; err != nil {
		logger.Log(err, "Ошибка получения полей таблицы", logger.Error)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if err := services.CheckFieldsAndValue(obj, fields, true); err != nil {
		c.JSON(400, gin.H{"error": "Ошибка валидации полей: " + err.Error()})
		return
	}
	if err := o.DB.Table(objectName).Create(&obj).Error; err != nil {
		logger.Log(err, "Ошибка создания элемента", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка создания элемента: " + err.Error()})
		return
	}
	c.Status(201)
}
