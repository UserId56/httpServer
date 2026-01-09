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

func NewUserGetResponseFromUser(user *User) *UserGetResponse {
	if user == nil {
		return nil
	}
	return &UserGetResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: user.DeletedAt,
		Username:  user.Username,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
	}
}
