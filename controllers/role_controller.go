package controllers

import (
	"encoding/json"
	"errors"
	"httpServer/logger"
	"httpServer/models"
	"httpServer/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/datatypes"
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

	var roleInput map[string]interface{}
	if err := c.ShouldBindJSON(&roleInput); err != nil {
		c.JSON(400, gin.H{"error": "Не валидный JSON или не валидные поля"})
		return
	}
	if permission, ok := roleInput["permission"]; ok {
		b, err := json.Marshal(permission)
		if err != nil {
			c.JSON(400, gin.H{"error": "Некорректный формат permission"})
			return
		}
		roleInput["permission"] = datatypes.JSON(b)
	}
	var pgErr *pgconn.PgError
	if err := rc.DB.Model(&models.Role{}).Where("id = ?", intRoleId).Updates(roleInput).Error; err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(400, gin.H{"error": "Роль с таким именем уже существует"})
			return
		}
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.Status(200)
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
	var Query models.Query
	if err := c.ShouldBindJSON(&Query); err != nil {
		c.JSON(400, gin.H{"error": "Не валидный JSON или не валидные поля"})
		return
	}
	var fields []models.DynamicColumns
	if err := rc.DB.Model(&models.DynamicColumns{}).Where("dynamic_table_id = (SELECT id FROM dynamic_schemes WHERE name = ?)", "roles").Find(&fields).Error; err != nil {
		c.JSON(500, gin.H{"error": "Ошибка получения полей для объекта roles"})
		return
	}
	whereSQL, args, err := services.WhereGeneration(Query.Where, fields, "AND")
	if err != nil {
		c.JSON(400, gin.H{"error": "Ошибка генерации условия: " + err.Error()})
		return
	}
	var roles []models.Role
	queryBuilder := rc.DB.Model(&models.Role{})
	if whereSQL != "" {
		queryBuilder = queryBuilder.Where(whereSQL, args...)
	}
	if len(Query.Include) == 0 {
		Query.Include = []string{"created_at", "updated_at", "deleted_at, id, name, permission"}
	}
	queryBuilder = queryBuilder.Select(Query.Include)
	if Query.Count {
		var count int64
		if err := queryBuilder.Count(&count).Error; err != nil {
			c.JSON(500, gin.H{"error": "Ошибка подсчета записей"})
			return
		}
		c.Header("X-Total-Count", strconv.FormatInt(count, 10))
	}
	if Query.Take > 0 {
		queryBuilder = queryBuilder.Limit(Query.Take)
	}
	if Query.Skip > 0 {
		queryBuilder = queryBuilder.Offset(Query.Skip)
	}
	if len(Query.Order) > 0 {
		orderStr, err := services.OrderGeneration(Query.Order, fields)
		if err != nil {
			c.JSON(400, gin.H{"error": "Ошибка в поле сортировки: " + err.Error()})
			return
		}
		queryBuilder = queryBuilder.Order(orderStr)

	}
	if err := queryBuilder.Find(&roles).Error; err != nil {
		logger.Log(err, "Ошибка получения роли", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка выполнения запроса"})
		return
	}
	c.JSON(200, gin.H{"roles": roles})
}
