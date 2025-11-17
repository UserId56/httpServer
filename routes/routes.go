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
	{
		private.Use(middleware.Auth())
		//USER METHODS
		user := private.Group("/user")
		{
			user.GET("/profile", appController.User.UserGetMyProfile)
			user.GET("/:id", middleware.RequirePermission(db, []string{"user.GET"}), appController.User.UserGetByID)
			user.DELETE("/:id", middleware.RequirePermission(db, []string{"user.DELETE"}), appController.User.UserDeleteByID)
			user.PUT("/:id", middleware.RequirePermission(db, []string{"user.PUT"}), appController.User.UserUpdateByID)
			user.GET("/query", middleware.RequirePermission(db, []string{"user.GET"}), appController.User.UserQuery)
		}
		//ROLE METHODS
		role := private.Group("/role")
		{
			role.GET("/:id", middleware.RequirePermission(db, []string{"role.GET"}), appController.Role.RoleGetByID)
			role.POST("/", middleware.RequirePermission(db, []string{"role.POST"}), appController.Role.RoleCreate)
			role.PUT("/:id", middleware.RequirePermission(db, []string{"role.PUT"}), appController.Role.RoleUpdateByID)
			role.DELETE("/:id", middleware.RequirePermission(db, []string{"role.DELETE"}), appController.Role.RoleDeleteByID)
		}
	}
}
