package controllers

import (
	"httpServer/logger"
	"httpServer/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TestController struct {
	DB *gorm.DB
}

func NewTestController(db *gorm.DB) *TestController {
	return &TestController{DB: db}
}

func (test *TestController) TestGetList(c *gin.Context) {
	var testList []models.Test
	if result := test.DB.Find(&testList); result.Error != nil {
		logger.LogError(result.Error, "Ошибка получения тестовых записей", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка получения тестовых записей"})
		return
	}
	c.JSON(200, gin.H{"list": testList})
}

func (test *TestController) TestCreate(c *gin.Context) {
	testList := &models.Test{}
	if err := c.ShouldBindJSON(testList); err != nil {
		c.JSON(400, gin.H{"error": "Не верные данные"})
		return
	}
	if err := test.DB.Create(testList).Error; err != nil {
		c.JSON(500, gin.H{"error": "Ошибка при создании тестовой записи"})
		return
	}
	c.JSON(201, gin.H{"message": "Тестовая запись создана", "test": testList})
}
