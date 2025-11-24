package controllers

import (
	"encoding/json"
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
	SqlQuery, dev, err := services.GenerateCreateTableSQL(req, false)
	if dev {
		fmt.Printf(SqlQuery)
		return
	}
	if err != nil {
		c.JSON(400, gin.H{"error": "Ошибка создания таблицы: " + err.Error()})
		tx.Rollback()
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
	scheme, err := services.SchemeHelperGetByName(tc.DB, schemeName)
	if err != nil {
		if errors.As(err, &gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Таблица не найдена"})
			return
		}
		logger.Log(err, "Ошибка получения таблицы по имени", logger.Error)
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

func (tc *SchemeController) SchemeUpdateByName(c *gin.Context) {
	schemeName := c.Param("name")
	scheme, err := services.SchemeHelperGetByName(tc.DB, schemeName)
	if err != nil {
		if errors.As(err, &gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Таблица не найдена"})
			return
		}
		logger.Log(err, "Ошибка получения таблицы по имени", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	var req models.DynamicScheme
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

	// Работа с колонками
	if len(req.Columns) > 0 {
		//SQL для обновления таблицы(изменения имеющихся столбцов и удаление старых, новые колонки, обновленные колонки
		sqlStr, newColumns, deleteColumns, err := services.GenerateUpdateTableSQL(req.Columns, scheme)
		if err != nil {
			logger.Log(err, "Ошибка генерации SQL для обновления таблицы", logger.Error)
			c.JSON(400, gin.H{"error": "Ошибка создания таблицы: " + err.Error()})
			tx.Rollback()
			return
		}
		//Удаляем колонки из информации о таблице
		for _, delCol := range deleteColumns {
			fmt.Printf("-Удаляем колонку: %+v\n", delCol)
			//Удаляем из информации о таблице удаленные колонки
			res := tx.Unscoped().Delete(delCol)
			if res.Error != nil {
				logger.Log(res.Error, "Ошибка удаления информации о колонке таблицы", logger.Error)
				c.JSON(500, gin.H{"error": "Ошибка на сервере"})
				tx.Rollback()
				return
			}
			if res.RowsAffected == 0 {
				c.JSON(404, gin.H{"error": "Колонка для удаления не найдена"})
				tx.Rollback()
				return
			}
		}
		//Добавляем новые колонки в таблицу и в информацию о таблице
		if len(newColumns) > 0 {
			var newScheme models.CreateSchemeRequest
			newScheme.Name = schemeName
			for _, columnD := range newColumns {
				for _, currentColumn := range scheme.Columns {
					fmt.Printf("%v == %v = %v\n", columnD.ColumnName, currentColumn.ColumnName, columnD.ColumnName == currentColumn.ColumnName)
					if columnD.ColumnName == currentColumn.ColumnName {
						c.JSON(400, gin.H{"error": "Колонка с именем " + columnD.ColumnName + " уже существует"})
						tx.Rollback()
						return
					}
				}
				sourceColumnD := *columnD
				fmt.Printf("-!!!!%+v!!!!!-", sourceColumnD)
				jsonStruct, err := json.Marshal(sourceColumnD)
				if err != nil {
					logger.Log(err, "Ошибка маршалинга новой колонки", logger.Error)
					c.JSON(500, gin.H{"error": "Ошибка на сервере"})
					tx.Rollback()
					return
				}
				var convertedStruct models.ColumnDefinition
				if err := json.Unmarshal(jsonStruct, &convertedStruct); err != nil {
					logger.Log(err, "Ошибка анмаршалинга новой колонки", logger.Error)
					c.JSON(500, gin.H{"error": "Ошибка на сервере"})
					tx.Rollback()
					return
				}
				newScheme.Columns = append(newScheme.Columns, convertedStruct)
			}
			sqlAddNewColumns, _, err := services.GenerateCreateTableSQL(newScheme, true)
			if err != nil {
				logger.Log(err, "Ошибка генерации SQL для добавления новых колонок", logger.Error)
				c.JSON(500, gin.H{"error": "Ошибка на сервере"})
				tx.Rollback()
				return
			}
			if err := tx.Exec(sqlAddNewColumns).Error; err != nil {
				logger.Log(err, "Ошибка добавления новых колонок в таблицу в базе данных", logger.Error)
				c.JSON(500, gin.H{"error": "Ошибка на сервере"})
				tx.Rollback()
				return
			}
			if err := tx.Create(&newColumns).Error; err != nil {
				logger.Log(err, "Ошибка добавления новых колонок в информацию о таблице", logger.Error)
				c.JSON(500, gin.H{"error": "Ошибка на сервере"})
				tx.Rollback()
				return
			}
		}
		//обновляем таблицу
		if err := tx.Exec(sqlStr).Error; err != nil {
			logger.Log(err, "Ошибка обновления таблицы в базе данных", logger.Error)
			c.JSON(500, gin.H{"error": "Ошибка на сервере"})
			tx.Rollback()
			return
		}
		//Обновляем информацию о столбцах, которые не были удалены и не были добавлены заново
		for _, colUpdate := range req.Columns {
			fmt.Printf("%+v\n", colUpdate)
			if err := tx.Updates(&colUpdate).Error; err != nil {
				logger.Log(err, "Ошибка обновления информации о колонке таблицы", logger.Error)
				c.JSON(500, gin.H{"error": "Ошибка на сервере"})
				tx.Rollback()
				return
			}
		}

	}
	// Обновляем информацию о самой таблице
	if err := tx.Model(&scheme).Omit("Columns").Updates(req).Error; err != nil {
		logger.Log(err, "Ошибка обновления информации о таблице", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		tx.Rollback()
		return
	}
	if err := tx.Commit().Error; err != nil {
		logger.Log(err, "Ошибка коммита транзакции", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.Status(200)
}
