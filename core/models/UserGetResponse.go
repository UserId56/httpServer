package models

import "gorm.io/gorm"

type UserGetResponse struct {
	gorm.Model
	Username string `gorm:"type:text;unique;not null" json:"username"`
	Avatar   string `gorm:"type:text" json:"avatar"`
	Bio      string `gorm:"type:text" json:"bio"`
}
