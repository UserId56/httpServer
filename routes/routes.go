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
}
