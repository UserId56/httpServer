package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name       string   `gorm:"type:text;unique;not null" json:"name"`
	Permission []string `gorm:"type:jsonb;serializer:json" json:"permission,omitempty"`
}
