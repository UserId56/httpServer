package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Username string `gorm:"type:text;unique;not null" json:"username"`
	Email    string `gorm:"type:text;unique;not null" json:"email"`
	Password string `gorm:"type:text;not null" json:"password"`
	Avatar   string `gorm:"type:text" json:"avatar"`
	Bio      string `gorm:"type:text" json:"bio"`
	RoleID   *uint  `gorm:"index;default:2" json:"role_id,omitempty"`
	Role     *Role  `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"role,omitempty"`
}
