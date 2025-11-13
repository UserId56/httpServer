package routes

import (
	"httpServer/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(r *gin.Engine, db *gorm.DB) {
	appController := controllers.NewAppController(db)
	r.POST("/register", appController.User.UserRegistration)
	r.GET("/test", appController.Test.TestGetList)
	r.POST("/test", appController.Test.TestCreate)
	r.GET("/roles", appController.Test.TestRole)
	r.POST("/user/register", appController.User.UserRegistration)
	r.POST("/user/login", appController.User.UserLogin)
	r.GET("/user/profile", appController.User.UserGetMyProfile)
	r.GET("/user/:id", appController.User.UserGetByID)
}
