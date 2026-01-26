package models

import (
	"time"

	"gorm.io/gorm"
)

type CreateRoleRequest struct {
	Name       string          `json:"name"`
	Permission map[string]bool `json:"permission"`
}

type Role struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name       string          `gorm:"type:text;unique;not null" json:"name"`
	Permission map[string]bool `gorm:"type:jsonb;serializer:json" json:"permission"`
	IsSystem   bool            `gorm:"type:boolean;default:false" json:"is_system"`
}
