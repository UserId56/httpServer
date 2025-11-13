package controllers

import (
	"errors"
	"httpServer/logger"
	"httpServer/models"
	"httpServer/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
		logger.LogError(err, "Ошибка хеширования пароля", logger.Error)
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
		logger.LogError(tx.Error, "Ошибка создания транзакции", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	if err := tx.Create(&user).Error; err != nil {
		logger.LogError(err, "Ошибка создания пользователя", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	refreshToken := models.RefreshToken{
		Model:  gorm.Model{},
		UserID: user.ID,
		Token:  uuid.NewString(),
	}

	if err := tx.Create(&refreshToken).Error; err != nil {
		logger.LogError(err, "Ошибка создания refresh токена", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}
	if err := tx.Commit().Error; err != nil {
		logger.LogError(err, "Ошибка коммита транзакции", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	jwt, err := services.GenerateJWT(&user)
	if err != nil {
		logger.LogError(err, "Ошибка генерации JWT", logger.Error)
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

func (uc UserController) UserLogin(c *gin.Context) {
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
		logger.LogError(err, "Ошибка создания refresh токена", logger.Error)
		c.JSON(500, gin.H{"error": "Ошибка на сервере"})
		return
	}

	jwt, err := services.GenerateJWT(&user)
	if err != nil {
		logger.LogError(err, "Ошибка генерации JWT", logger.Error)
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

func (uc UserController) UserGetMyProfile(c *gin.Context) {
	JWT := c.GetHeader("Authorization")
	token, err := services.ParseJWT(JWT)
	if err != nil || !token.Valid {
		c.JSON(401, gin.H{"error": "Не авторизован"})
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(401, gin.H{"error": "Не авторизован"})
			return
		}
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
