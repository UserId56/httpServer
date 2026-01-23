package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/UserId56/httpServer/core/logger"
	"github.com/UserId56/httpServer/core/models"
	"github.com/UserId56/httpServer/core/services"
	"gorm.io/gorm/clause"

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
	if services.CheckTableName(objectName) {
		c.JSON(404, gin.H{"error": "Таблица не найдена"})
		return
	}
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
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	var fields []models.DynamicColumns
	if err := o.DB.Model(&models.DynamicColumns{}).Where("dynamic_table_id = ?", table.ID).Find(&fields).Error; err != nil {
		logger.Log(err, "Ошибка получения полей таблицы", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if err := services.CheckFieldsAndValue(obj, fields, true); err != nil {
		c.JSON(400, gin.H{"error": "Ошибка валидации полей: " + err.Error()})
		return
	}
	userId, exist := c.Get("user_id")
	if !exist {
		c.JSON(401, gin.H{"error": "Пользователь не аутентифицирован"})
		return
	}
	obj["owner_id"] = uint(userId.(float64))
	// Чекаем множественные значения в полях ref и конвертируем их в []int
	obj = services.ParsIntField(fields, obj)
	obj, err := services.ParsDataTime(fields, obj, time.Local)
	if err != nil {
		c.JSON(400, gin.H{"error": "Ошибка даты: " + err.Error()})
		return
	}
	if err := o.DB.Table(objectName).Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Create(&obj).Error; err != nil {
		logger.Log(err, "Ошибка создания элемента", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.JSON(201, gin.H{"id": obj["id"]})
}

func (o *ObjectController) ObjectGetByID(c *gin.Context) {
	ObjectID := c.Param("object")
	if services.CheckTableName(ObjectID) {
		c.JSON(404, gin.H{"error": "Таблица не найдена"})
		return
	}
	ElementID := c.Param("id")
	result := make(map[string]interface{})
	tx := o.DB.Table(ObjectID).Where("id = ?", ElementID).Limit(1).Find(&result)
	if tx.Error != nil {
		logger.Log(tx.Error, "Ошибка получения элемента", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if tx.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Элемент не найден"})
		return
	}
	var fields []models.DynamicColumns
	if err := o.DB.Model(&models.DynamicColumns{}).Where("dynamic_table_id = (SELECT id FROM dynamic_schemes WHERE name = ?)", ObjectID).Find(&fields).Error; err != nil {
		logger.Log(err, "Ошибка получения полей для объекта "+ObjectID, logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	result = services.ExtractInt64Slice(fields, result)
	c.JSON(200, result)
}

func (o *ObjectController) ObjectUpdateByID(c *gin.Context) {
	ObjectID := c.Param("object")
	if services.CheckTableName(ObjectID) {
		c.JSON(404, gin.H{"error": "Таблица не найдена"})
		return
	}
	ElementID := c.Param("id")
	var obj map[string]interface{}
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var table models.DynamicScheme
	if err := o.DB.Where("name = ?", ObjectID).First(&table).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "Таблица не найдена"})
			return
		}
		logger.Log(err, "Ошибка получения схемы таблицы", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	var fields []models.DynamicColumns
	if err := o.DB.Model(&models.DynamicColumns{}).Where("dynamic_table_id = ?", table.ID).Find(&fields).Error; err != nil {
		logger.Log(err, "Ошибка получения полей таблицы", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if err := services.CheckFieldsAndValue(obj, fields, false); err != nil {
		c.JSON(400, gin.H{"error": "Ошибка валидации полей: " + err.Error()})
		return
	}
	obj["updated_at"] = time.Now()
	_, exist := obj["owner_id"]
	if exist {
		delete(obj, "owner_id")
	}
	obj = services.ParsIntField(fields, obj)
	obj, err := services.ParsDataTime(fields, obj, time.Local)
	if err != nil {
		c.JSON(400, gin.H{"error": "Ошибка даты: " + err.Error()})
		return
	}
	if err := o.DB.Table(ObjectID).Where("id = ?", ElementID).Updates(obj).Error; err != nil {
		logger.Log(err, "Ошибка обновления элемента", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.Status(200)
}

func (o *ObjectController) ObjectDeleteByID(c *gin.Context) {
	ObjectID := c.Param("object")
	if services.CheckTableName(ObjectID) {
		c.JSON(404, gin.H{"error": "Таблица не найдена"})
		return
	}
	ElementID := c.Param("id")
	type DeleteStruct struct {
		ID        string
		DeletedAt gorm.DeletedAt
	}
	var Element DeleteStruct
	Element.ID = ElementID
	if err := o.DB.Table(ObjectID).Unscoped().Where("id = ?", ElementID).First(&Element).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "Элемент не найден"})
			return
		}
		logger.Log(err, "Ошибка получения элемента", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if !Element.DeletedAt.Time.IsZero() {
		if err := o.DB.Table(ObjectID).Unscoped().Delete(Element).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Элемент не найден"})
			return
		}
		c.JSON(200, gin.H{"message": "Элемент успешно удален окончательно"})
		return
	}
	now := time.Now()
	if err := o.DB.Table(ObjectID).Where("id = ?", ElementID).UpdateColumns(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error; err != nil {
		logger.Log(err, "Ошибка удаления элемента", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.Status(200)
}

func (o *ObjectController) ObjectQuery(c *gin.Context) {
	var Query models.Query
	object := c.Param("object")
	if services.CheckTableName(object) {
		c.JSON(404, gin.H{"error": "Таблица не найдена"})
		return
	}
	if err := c.ShouldBindJSON(&Query); err != nil {
		c.JSON(400, gin.H{"error": "Не верные данные"})
		return
	}
	var fields []models.DynamicColumns
	if err := o.DB.Model(&models.DynamicColumns{}).Where("dynamic_table_id = (SELECT id FROM dynamic_schemes WHERE name = ?)", object).Find(&fields).Error; err != nil {
		logger.Log(err, "Ошибка получения полей для объекта "+object, logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	whereSQL, args, err := services.WhereGeneration(Query.Where, fields, "AND")
	fmt.Println(whereSQL)
	if err != nil {
		logger.Log(err, "Ошибка генерации условия", logger.Error)
		c.JSON(400, gin.H{"error": "Ошибка генерации условия: " + err.Error()})
		return
	}
	var results []map[string]interface{}
	if len(Query.Include) == 0 {
		Query.Include = services.GenInclude(fields)
	} else {
		Query.IncludeBox()
	}
	userPermissions, exists := c.Get("permission")
	if !exists {
		logger.Log(errors.New("роль не указана"), "Ошибка получения прав роли из контекста", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	for _, field := range userPermissions.([]string) {
		if strings.Contains(field, "forbidden") {
			strArray := strings.Split(field, ".")
			for _, qField := range Query.Include {
				qField = strings.Trim(qField, "\"")
				if qField == strArray[1] {
					// NOTE: Если право есть в правах роли, то ему НЕЛЬЗЯ делать действие
					c.JSON(403, gin.H{"error": "Недостаточно прав для получения поля: " + qField})
					return
				}
			}
		}
	}
	dbQuery := o.DB.Table(object).Select(Query.Include)
	if whereSQL != "" {
		dbQuery = dbQuery.Where(whereSQL, args...)
	}
	if Query.Count {
		var count int64
		if err := dbQuery.Count(&count).Error; err != nil {
			logger.Log(err, "Ошибка подсчета количества пользователей", logger.Error)
			c.JSON(500, gin.H{"error": "Ошибка на сервере"})
			return
		}
		c.Header("X-Total-Count", strconv.FormatInt(count, 10))
	}
	if Query.Take > 0 {
		dbQuery = dbQuery.Limit(Query.Take)
	}
	if Query.Skip > 0 {
		dbQuery = dbQuery.Offset(Query.Skip)
	}
	if len(Query.Order) > 0 {
		orderStr, err := services.OrderGeneration(Query.Order, fields)
		if err != nil {
			c.JSON(400, gin.H{"error": "Ошибка генерации сортировки: " + err.Error()})
			return
		}
		dbQuery = dbQuery.Order(orderStr)
	}
	if err := dbQuery.Find(&results).Error; err != nil {
		logger.Log(err, "Ошибка получения пользователей", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if (results == nil) || (len(results) == 0) {
		results = make([]map[string]interface{}, 0)
	}
	for i, res := range results {
		results[i] = services.ExtractInt64Slice(fields, res)
	}
	c.JSON(200, results)
}
