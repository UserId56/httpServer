package models

import "gorm.io/gorm"

type RefreshToken struct {
	gorm.Model
	Token  string `gorm:"uniqueIndex;size:36" json:"token"`
	UserID uint   `gorm:"not null;index" json:"user_id"`
	User   *User  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}
