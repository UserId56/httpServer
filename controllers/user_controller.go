package controllers

import (
	"errors"
	"httpServer/logger"
	"httpServer/models"
	"httpServer/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		Model:    gorm.Model{},
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

	if err := tx.Create(&user).Error; err != nil {
		if errors.As(err, &gorm.ErrDuplicatedKey) {
			c.JSON(400, gin.H{"error": "Пользователь с таким именем или email уже существует"})
			tx.Rollback()
			return
		}
		logger.Log(err, "Ошибка создания пользователя", logger.Error)
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	refreshToken := models.RefreshToken{
		Model:  gorm.Model{},
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

	c.JSON(200, gin.H{"user": userResponse})
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
		Model:  gorm.Model{},
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

	c.JSON(200, gin.H{"user": userResponse})
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
		c.JSON(200, gin.H{"user": user})
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
	var user models.UserGetResponse
	if err := uc.DB.Model(&models.User{}).Where("id = ?", id).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Пользователь не найден"})
		return
	}
	c.JSON(200, gin.H{"user": user})
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
	if password, ok := userInput["password"].(string); ok {
		hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Log(err, "Ошибка хеширования пароля", logger.Error)
			c.JSON(500, gin.H{"error": "Ошибка на сервере"})
			return
		}
		userInput["password"] = string(hashPassword)
	}
	result := uc.DB.Model(&models.User{}).Where("id = ?", id).Updates(userInput)
	if result.Error != nil {
		logger.Log(result.Error, "Ошибка обновления пользователя", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Пользователь не найден"})
		return
	}
	c.JSON(200, gin.H{"message": "Пользователь успешно обновлен"})
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
		c.JSON(201, gin.H{"message": "Пользователь успешно удален окончательно"})
		return
	}
	if err := uc.DB.Delete(&models.User{}, id).Error; err != nil {
		logger.Log(err, "Ошибка удаления пользователя", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.JSON(201, gin.H{"message": "Пользователь успешно удален"})
}

func (uc *UserController) UserQuery(c *gin.Context) {
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
	var count int64
	var users []models.UserGetResponse
	err = uc.DB.Model(&models.User{}).Count(&count).Error
	if count == 0 {
		c.Header("X-Total-Count", "0")
		c.JSON(200, gin.H{"users": []models.User{}})
		return
	}
	if err != nil {
		logger.Log(err, "Ошибка получения количества пользователей", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	err = uc.DB.Model(&models.User{}).Offset(skipInt).Limit(takeInt).Find(&users).Error
	if err != nil {
		logger.Log(err, "Ошибка получения списка пользователей", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(count, 10))
	c.JSON(200, gin.H{"users": users})
}
