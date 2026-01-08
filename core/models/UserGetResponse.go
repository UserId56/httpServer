package models

import (
	"time"

	"gorm.io/gorm"
)

type UserGetResponse struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Username string `gorm:"type:text;unique;not null" json:"username"`
	Avatar   string `gorm:"type:text" json:"avatar"`
	Bio      string `gorm:"type:text" json:"bio"`
}
