package controllers

import "gorm.io/gorm"

type AppController struct {
	User  *UserController
	Test  *TestController
	Role  *RoleController
	Sheme *SchemeController
}

func NewAppController(db *gorm.DB) *AppController {
	return &AppController{
		User:  NewUserController(db),
		Test:  NewTestController(db),
		Role:  NewRoleController(db),
		Sheme: NewSchemeController(db),
	}
}
