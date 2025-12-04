package controllers

import (
	"github.com/UserId56/httpServer/core/logger"
	"github.com/UserId56/httpServer/core/models"
	"github.com/UserId56/httpServer/core/services"

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
		logger.Log(result.Error, "Ошибка получения тестовых записей", logger.Error)
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

func (test *TestController) TestRole(c *gin.Context) {
	var rolesList []models.Role
	if result := test.DB.Find(&rolesList); result.Error != nil {
		logger.Log(result.Error, "Ошибка получения ролей", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка получения ролей"})
		return
	}
	c.JSON(200, gin.H{"roles": rolesList})
}

func (test *TestController) TestWhere(c *gin.Context) {
	var data models.Query
	object := c.Param("object")
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, gin.H{"error": "Не верные данные"})
		return
	}
	var fields []models.DynamicColumns
	if err := test.DB.Model(&models.DynamicColumns{}).Where("dynamic_table_id = (SELECT id FROM dynamic_schemes WHERE name = ?)", object).Find(&fields).Error; err != nil {
		logger.Log(err, "Ошибка получения полей для объекта "+object, logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка получения полей для объекта " + object})
		return
	}
	whereSQL, args, err := services.WhereGeneration(data.Where, fields, "AND")
	if err != nil {
		logger.Log(err, "Ошибка генерации условия", logger.Error)
		c.JSON(400, gin.H{"error": "Ошибка генерации условия: " + err.Error()})
		return
	}
	var results []map[string]interface{}
	query := test.DB.Table(object).Where(whereSQL, args...)
	if err := query.Find(&results).Error; err != nil {
		logger.Log(err, "Ошибка выполнения запроса", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка выполнения запроса"})
		return
	}
	c.JSON(200, gin.H{"results": results})

}
