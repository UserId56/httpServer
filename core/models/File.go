package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	FileId   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();index;column:file_id" json:"file_id"`
	Name     string    `gorm:"type:text;not null" json:"name"`
	FileSize int64     `gorm:"type:bigint;not null" json:"file_size"`
	Path     string    `gorm:"type:text;not null" json:"-"`
	OwnerID  uint      `gorm:"type:bigint" json:"owner_id"`
}
