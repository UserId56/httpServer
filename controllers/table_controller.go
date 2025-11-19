package controllers

import (
	"fmt"
	"httpServer/logger"
	"httpServer/models"
	"httpServer/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TableController struct {
	DB *gorm.DB
}

func NewTableController(db *gorm.DB) *TableController {
	return &TableController{DB: db}
}

func (tc *TableController) CreateDynamicTable(c *gin.Context) {
	var req models.CreateTableRequest
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
	tx.Model(&models.DynamicTable{}).Where("name = ?", req.Name).Count(&isExist)
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
