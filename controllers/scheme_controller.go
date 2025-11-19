package controllers

import (
	"errors"
	"fmt"
	"httpServer/logger"
	"httpServer/models"
	"httpServer/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SchemeController struct {
	DB *gorm.DB
}

func NewSchemeController(db *gorm.DB) *SchemeController {
	return &SchemeController{DB: db}
}

func (tc *SchemeController) SchemeCreate(c *gin.Context) {
	var req models.CreateSchemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log(err, "Ошибка привязки JSON", logger.Error)
		c.JSON(400, gin.H{"error": "Не валидный JSON или не валидные поля"})
		return
	}
	tx := tc.DB.Begin()
	if tx.Error != nil {
		logger.Log(tx.Error, "Ошибка создания транзакции", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	var isExist int64
	tx.Model(&models.DynamicScheme{}).Where("name = ?", req.Name).Count(&isExist)
	if isExist > 0 {
		c.JSON(400, gin.H{"error": "Таблица с таким именем уже существует"})
		tx.Rollback()
		return
	}
	SqlQuery, dev, _ := services.GenerateCreateTableSQL(req)
	if dev {
		fmt.Printf(SqlQuery)
		return
	}
	if err := tx.Exec(SqlQuery).Error; err != nil {
		logger.Log(err, "Ошибка создания таблицы", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		tx.Rollback()
		return
	}
	dynamicTable := req.CreateDynamicTable()
	if err := tx.Create(&dynamicTable).Error; err != nil {
		logger.Log(err, "Ошибка сохранения информации о таблице", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		tx.Rollback()
		return
	}
	tx.Commit()
	c.Status(201)
}

func (tc *SchemeController) SchemeDelete(c *gin.Context) {
	schemeName := c.Param("name")
	fmt.Println("!!!!!!!!!!!!!!!!!schemeName:", schemeName)
	tx := tc.DB.Begin()
	if tx.Error != nil {
		logger.Log(tx.Error, "Ошибка создания транзакции", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if err := tx.Where("name = ?", schemeName).First(&models.DynamicScheme{}).Error; err != nil {
		if errors.As(err, &gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Таблица не найдена"})
			tx.Rollback()
			return
		}
		logger.Log(err, "Ошибка поиска таблицы", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		tx.Rollback()
		return
	}
	migrator := tx.Migrator()
	if err := migrator.DropTable(schemeName); err != nil {
		logger.Log(err, "Ошибка удаления таблицы из базы данных", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		tx.Rollback()
		return
	}
	if err := tx.Unscoped().Where("name = ?", schemeName).Delete(&models.DynamicScheme{}).Error; err != nil {
		logger.Log(err, "Ошибка удаления информации о таблице", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		tx.Rollback()
		return
	}
	tx.Commit()
	c.Status(201)
}

func (tc *SchemeController) SchemeGetByName(c *gin.Context) {
	schemeName := c.Param("name")
	var scheme models.DynamicScheme
	if err := tc.DB.Preload("Columns").Where("name = ?", schemeName).First(&scheme).Error; err != nil {
		if errors.As(err, &gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Таблица не найдена"})
			return
		}
		logger.Log(err, "Ошибка поиска таблицы", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.JSON(200, scheme)
}

func (tc *SchemeController) SchemeGetLst(c *gin.Context) {
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
	var schemes []models.DynamicScheme
	if err := tc.DB.Limit(takeInt).Offset(skipInt).Find(&schemes).Error; err != nil {
		logger.Log(err, "Ошибка получения списка таблиц", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.JSON(200, schemes)
}
