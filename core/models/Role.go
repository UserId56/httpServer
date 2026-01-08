package models

import (
	"time"

	"gorm.io/gorm"
)

type CreateRoleRequest struct {
	Name       string   `json:"name"`
	Permission []string `json:"permission"`
}

type Role struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name       string   `gorm:"type:text;unique;not null" json:"name"`
	Permission []string `gorm:"type:jsonb;serializer:json" json:"permission"`
}
