package routes

import (
	"github.com/UserId56/httpServer/core/controllers"
	"github.com/UserId56/httpServer/core/middleware"

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
		private.Use(middleware.Auth(db))
		//USER METHODS
		user := private.Group("/user")
		{
			user.GET("/profile", appController.User.UserGetMyProfile)
			user.GET("/:id", middleware.RequirePermission(db, []string{"user.GET"}, false), appController.User.UserGetByID)
			user.DELETE("/:id", middleware.RequirePermission(db, []string{"user.DELETE"}, false), appController.User.UserDeleteByID)
			user.PUT("/:id", middleware.RequirePermission(db, []string{"user.PUT"}, false), appController.User.UserUpdateByID)
			user.POST("/query", middleware.RequirePermission(db, []string{"user.GET"}, false), appController.User.UserQuery)
		}
		//ROLE METHODS
		role := private.Group("/role")
		{
			role.GET("/:id", middleware.RequirePermission(db, []string{"role.GET"}, false), appController.Role.RoleGetByID)
			role.POST("/", middleware.RequirePermission(db, []string{"role.POST"}, false), appController.Role.RoleCreate)
			role.PUT("/:id", middleware.RequirePermission(db, []string{"role.PUT"}, false), appController.Role.RoleUpdateByID)
			role.DELETE("/:id", middleware.RequirePermission(db, []string{"role.DELETE"}, false), appController.Role.RoleDeleteByID)
			role.POST("/query", middleware.RequirePermission(db, []string{"role.GET"}, false), appController.Role.RoleQuery)
		}
		//	SCHEME METHODS
		scheme := private.Group("/scheme")
		{
			scheme.POST("/", middleware.RequirePermission(db, []string{"scheme.POST"}, false), appController.Scheme.SchemeCreate)
			scheme.GET("/:name", middleware.RequirePermission(db, []string{"scheme.GET"}, false), appController.Scheme.SchemeGetByName)
			scheme.GET("/", middleware.RequirePermission(db, []string{"scheme.GET"}, false), appController.Scheme.SchemeGetLst)
			scheme.PUT("/:name", middleware.RequirePermission(db, []string{"scheme.PUT"}, false), appController.Scheme.SchemeUpdateByName)
			scheme.DELETE("/:name", middleware.RequirePermission(db, []string{"scheme.DELETE"}, false), appController.Scheme.SchemeDelete)
		}
		object := private.Group("/:object")
		{
			object.POST("/", middleware.RequirePermission(db, []string{}, true), appController.Object.ObjectCreate)
			object.GET("/:id", middleware.RequirePermission(db, []string{}, true), appController.Object.ObjectGetByID)
			object.PUT("/:id", middleware.RequirePermission(db, []string{}, true), appController.Object.ObjectUpdateByID)
			object.DELETE("/:id", middleware.RequirePermission(db, []string{}, true), appController.Object.ObjectDeleteByID)
			object.POST("/query", middleware.RequirePermission(db, []string{}, true), appController.Object.ObjectQuery)
		}
	}
}
