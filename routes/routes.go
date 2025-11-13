package routes

import (
	"httpServer/controllers"
	"httpServer/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(r *gin.Engine, db *gorm.DB) {
	appController := controllers.NewAppController(db)
	//r.POST("/register", appController.User.UserRegistration)
	r.GET("/test", appController.Test.TestGetList)
	r.POST("/test", appController.Test.TestCreate)
	r.POST("/user/register", appController.User.UserRegistration)
	r.POST("/user/login", appController.User.UserLogin)
	private := r.Group("/")
	private.Use(middleware.Auth())
	private.GET("/roles", appController.Test.TestRole)
	private.GET("/user/profile", appController.User.UserGetMyProfile)
	private.GET("/user/:id", appController.User.UserGetByID)
}
