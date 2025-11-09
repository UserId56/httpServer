package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

func (uc *UserController) UserRegistration(c *gin.Context) {
	c.JSON(200, gin.H{"message": "User registration successful"})
}
