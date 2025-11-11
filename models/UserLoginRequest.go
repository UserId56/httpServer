package models

type UserLoginRequest struct {
	Username string `json:"user_name" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=6"`
}
