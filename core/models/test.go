package models

import "gorm.io/gorm"

type Test struct {
	gorm.Model
	Name  string `gorm:"type:varchar(100);not null" json:"name"`
	Value string `gorm:"type:varchar(100);not null" json:"value"`
	Type  string `gorm:"type:varchar(100);not null;default:'Old'" json:"type"`
}
