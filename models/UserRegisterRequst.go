package models

type RegisterUserRequest struct {
	Username string `json:"user_name" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Avatar   string `json:"avatar,omitempty"`
	Bio      string `json:"bio,omitempty"`
	RoleID   *uint  `json:"role_id,omitempty"`
}
