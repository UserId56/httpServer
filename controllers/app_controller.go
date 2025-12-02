package controllers

import "gorm.io/gorm"

type AppController struct {
	User   *UserController
	Test   *TestController
	Role   *RoleController
	Scheme *SchemeController
	Object *ObjectController
}

func NewAppController(db *gorm.DB) *AppController {
	return &AppController{
		User:   NewUserController(db),
		Test:   NewTestController(db),
		Role:   NewRoleController(db),
		Scheme: NewSchemeController(db),
		Object: NewObjectController(db),
	}
}
