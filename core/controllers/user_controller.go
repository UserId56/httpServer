package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/UserId56/httpServer/core/logger"
	"github.com/UserId56/httpServer/core/models"
	"github.com/UserId56/httpServer/core/services"
	"gorm.io/gorm/clause"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

func (uc *UserController) UserRegistration(c *gin.Context) {
	var userInput models.RegisterUserRequest
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(400, gin.H{"error": "Не валидный JSON или не валидные поля"})
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log(err, "Ошибка хеширования пароля", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	user := models.User{
		Username: userInput.Username,
		Email:    userInput.Email,
		Password: string(hashPassword),
		Avatar:   userInput.Avatar,
		Bio:      userInput.Bio,
		RoleID:   userInput.RoleID,
		Role:     nil,
	}

	if user.RoleID == nil {
		var defaultRole uint = 2
		user.RoleID = &defaultRole
	}

	tx := uc.DB.Begin()
	if tx.Error != nil {
		logger.Log(tx.Error, "Ошибка создания транзакции", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	var pgErr *pgconn.PgError
	if err := tx.Create(&user).Error; err != nil {
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "fk_users_role":
				c.JSON(400, gin.H{"error": "Указанная роль не существует"})
				tx.Rollback()
				return
			case "uni_users_username", "uni_users_email":
				// Пользователь с таким email или username уже существует
				c.JSON(400, gin.H{"error": "Пользователь с таким именем или email уже существует"})
				tx.Rollback()
				return
			default:
				// Неизвестная ошибка базы данных
				logger.Log(err, "Ошибка создания пользователя", logger.Error)
				tx.Rollback()
				c.JSON(500, gin.H{"error": "Ошибка на сервере"})
				return

			}
		}
		logger.Log(err, "Ошибка создания пользователя", logger.Error)
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	refreshToken := models.RefreshToken{
		UserID: user.ID,
		Token:  uuid.NewString(),
	}

	if err := tx.Create(&refreshToken).Error; err != nil {
		logger.Log(err, "Ошибка создания refresh токена", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if err := tx.Commit().Error; err != nil {
		logger.Log(err, "Ошибка коммита транзакции", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	jwt, err := services.GenerateJWT(&user)
	if err != nil {
		logger.Log(err, "Ошибка генерации JWT", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	userResponse := models.UserAuthResponse{
		UserName:     user.Username,
		UserID:       user.ID,
		RefreshToken: refreshToken.Token,
		AccessToken:  jwt,
		RoleID:       user.RoleID,
	}

	c.JSON(200, userResponse)
}

func (uc *UserController) UserLogin(c *gin.Context) {
	var loginInput models.UserLoginRequest
	if err := c.ShouldBindJSON(&loginInput); err != nil {
		c.JSON(400, gin.H{"error": "Не валидный JSON или не валидные поля"})
		return
	}

	var user models.User
	if loginInput.Username == "" {
		if err := uc.DB.Where("email = ?", loginInput.Email).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(401, gin.H{"error": "Неверный логин или пароль"})
			return
		}
	} else {
		if err := uc.DB.Where("username = ?", loginInput.Username).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(401, gin.H{"error": "Неверный логин или пароль"})
			return
		}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginInput.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	refreshToken := models.RefreshToken{
		UserID: user.ID,
		Token:  uuid.NewString(),
	}
	if err := uc.DB.Create(&refreshToken).Error; err != nil {
		logger.Log(err, "Ошибка создания refresh токена", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	jwt, err := services.GenerateJWT(&user)
	if err != nil {
		logger.Log(err, "Ошибка генерации JWT", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	userResponse := models.UserAuthResponse{
		UserName:     user.Username,
		UserID:       user.ID,
		RefreshToken: refreshToken.Token,
		AccessToken:  jwt,
		RoleID:       user.RoleID,
	}

	c.JSON(200, userResponse)
}

func (uc *UserController) UserGetMyProfile(c *gin.Context) {
	userIDAuth, exists := c.Get("user_id")
	if exists {
		userID := userIDAuth.(float64)
		var user models.User
		if err := uc.DB.Model(&models.User{}).Select("created_at", "updated_at", "id", "username", "email", "avatar", "bio", "role_id").Where("id = ?", int(userID)).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Пользователь не найден"})
			return
		}
		c.JSON(200, user)
	} else {
		c.JSON(401, gin.H{"error": "Не авторизован"})
	}
}

func (uc *UserController) UserGetByID(c *gin.Context) {
	userID := c.Param("id")
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Неверный ID пользователя"})
		return
	}
	var user models.User
	if err := uc.DB.Model(&models.User{}).Where("id = ?", id).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Пользователь не найден"})
		return
	}
	user.Password = ""
	userRoleId, exist := c.Get("role_id")
	if !exist {
		c.JSON(403, gin.H{"error": "Роль не указана"})
		c.Abort()
		return
	}
	currentUserIdAuth, existsId := c.Get("user_id")
	currentUserId, err := strconv.ParseUint(fmt.Sprintf("%v", currentUserIdAuth), 10, 64)
	if err != nil {
		logger.Log(err, "Ошибка преобразования user_id из контекста", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if !existsId {
		c.JSON(401, gin.H{"error": "Не авторизован"})
		return
	}

	fmt.Printf("%d %d\n", user.ID, currentUserId)
	if userRoleId.(float64) == 1 || uint64(user.ID) == currentUserId {
		c.JSON(200, user)
		return
	}
	var userReq models.UserGetResponse
	userReq = *models.NewUserGetResponseFromUser(&user)
	c.JSON(200, userReq)
}

func (uc *UserController) UserUpdateByID(c *gin.Context) {
	userID := c.Param("id")
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Неверный ID пользователя"})
		return
	}
	UserIDAuth, exists := c.Get("user_id")
	floatUserIDAuth, ok := UserIDAuth.(float64)
	if !ok {
		logger.Log(errors.New("Ошибка преобразования user_id"), "Ошибка преобразования user_id из JWT", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if (!exists || uint64(floatUserIDAuth) != id) && uint64(floatUserIDAuth) != 1 {
		c.JSON(403, gin.H{"error": "Нет прав для изменения этого пользователя"})
		return
	}
	var userInput map[string]interface{}
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(400, gin.H{"error": "Не валидный JSON или не валидные поля"})
		return
	}
	err = services.ValidationFields(userInput)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if password, ok := userInput["password"].(string); ok {
		hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Log(err, "Ошибка хеширования пароля", logger.Error)
			c.JSON(500, gin.H{"error": "Ошибка на сервере"})
			return
		}
		userInput["password"] = string(hashPassword)
	}
	var pgErr *pgconn.PgError
	result := uc.DB.Model(&models.User{}).Where("id = ?", id).Updates(userInput)
	if result.Error != nil {
		if errors.As(result.Error, &pgErr) {
			switch pgErr.ConstraintName {
			case "fk_users_role":
				c.JSON(400, gin.H{"error": "Указанная роль не существует"})
				return
			case "uni_users_username":
				// Пользователь с таким username уже существует
				c.JSON(400, gin.H{"error": "Пользователь с таким именем уже существует"})
				return
			case "uni_users_email":
				// Пользователь с таким email уже существует
				c.JSON(400, gin.H{"error": "Пользователь с таким email уже существует"})
				return
			default:
				// Неизвестная ошибка базы данных
				logger.Log(result.Error, "Ошибка обновления пользователя", logger.Error)
				c.JSON(500, gin.H{"error": "Ошибка на сервере"})
				return
			}
		}
		logger.Log(result.Error, "Ошибка обновления пользователя", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Пользователь не найден"})
		return
	}
	c.Status(200)
}

func (uc *UserController) UserDeleteByID(c *gin.Context) {
	userID := c.Param("id")
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Неверный ID пользователя"})
		return
	}
	var user models.User
	if err := uc.DB.Model(&models.User{}).Unscoped().Where("id = ?", id).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Пользователь не найден"})
		return
	}
	if !user.DeletedAt.Time.IsZero() {
		if err := uc.DB.Unscoped().Delete(&models.User{}, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "Пользователь не найден"})
			return
		}
		c.JSON(200, gin.H{"message": "Пользователь успешно удален окончательно"})
		return
	}
	if err := uc.DB.Delete(&models.User{}, id).Error; err != nil {
		logger.Log(err, "Ошибка удаления пользователя", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.Status(200)
}

func (uc *UserController) UserQuery(c *gin.Context) {
	var Query models.Query
	object := "users"
	if err := c.ShouldBindJSON(&Query); err != nil {
		c.JSON(400, gin.H{"error": "Не верные данные"})
		return
	}
	var fields []models.DynamicColumns
	if err := uc.DB.Model(&models.DynamicColumns{}).Where("dynamic_table_id = (SELECT id FROM dynamic_schemes WHERE name = ?)", object).Find(&fields).Error; err != nil {
		logger.Log(err, "Ошибка получения полей для объекта "+object, logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка получения полей для объекта " + object})
		return
	}
	whereSQL, args, err := services.WhereGeneration(Query.Where, fields, "AND")
	if err != nil {
		logger.Log(err, "Ошибка генерации условия", logger.Error)
		c.JSON(400, gin.H{"error": "Ошибка генерации условия: " + err.Error()})
		return
	}
	var users []map[string]interface{}
	if len(Query.Include) == 0 {
		Query.Include = []string{"created_at", "updated_at", "id", "username", "avatar", "bio"}
	} else {
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
					if qField == strArray[1] {
						// NOTE: Если право есть в правах роли, то ему НЕЛЬЗЯ делать действие
						c.JSON(403, gin.H{"error": "Недостаточно прав для получения поля: " + qField})
						return
					}
				}
			}
		}
	}
	includes := make([]string, 0)
	for _, include := range Query.Include {
		if include == "password" {
			continue
		}
		includes = append(includes, fmt.Sprintf("\"%s\"", include))
	}
	dbQuery := uc.DB.Model(&models.User{}).Select(includes)
	if whereSQL != "" {
		dbQuery = dbQuery.Where(whereSQL, args...)
	}
	fmt.Printf("%v\n", Query.Count)
	if Query.Count {
		var count int64
		if err := dbQuery.Count(&count).Error; err != nil {
			logger.Log(err, "Ошибка подсчета количества пользователей", logger.Error)
			c.JSON(500, gin.H{"error": "Ошибка на сервере"})
			return
		}
		fmt.Printf("SUKA EBANAYA A: %d\n", count)
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
	if err := dbQuery.Find(&users).Error; err != nil {
		logger.Log(err, "Ошибка получения пользователей", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if users == nil {
		users = make([]map[string]interface{}, 0)
	}
	c.JSON(200, users)
}

func (uc *UserController) UserCreate(c *gin.Context) {
	var userData models.User
	if err := c.ShouldBindJSON(&userData); err != nil {
		fmt.Printf("%v\n", err)
		c.JSON(400, gin.H{"error": "Не валидный JSON или не валидные поля"})
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log(err, "Ошибка хеширования пароля", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	userData.Password = string(hashPassword)
	var pgErr *pgconn.PgError
	if err := uc.DB.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Create(&userData).Error; err != nil {
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "fk_users_role":
				c.JSON(400, gin.H{"error": "Указанная роль не существует"})
				return
			case "uni_users_username", "uni_users_email":
				// Пользователь с таким email или username уже существует
				c.JSON(400, gin.H{"error": "Пользователь с таким именем или email уже существует"})
				return
			default:
				// Неизвестная ошибка базы данных
				logger.Log(err, "Ошибка создания пользователя", logger.Error)
				c.JSON(500, gin.H{"error": "Ошибка на сервере"})
				return

			}
		}
		logger.Log(err, "Ошибка создания пользователя", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	c.JSON(201, gin.H{"id": userData.ID})
}
