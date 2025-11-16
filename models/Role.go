package models

import "gorm.io/gorm"

type CreateRoleRequest struct {
	Name       string   `json:"name"`
	Permission []string `json:"permissions"`
}

type Role struct {
	gorm.Model
	Name       string   `gorm:"type:text;unique;not null" json:"name"`
	Permission []string `gorm:"type:jsonb;serializer:json" json:"permission,omitempty"`
}
