package models

type UserAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       uint   `json:"user_id"`
	UserName     string `json:"username"`
	RoleID       *uint  `json:"role_id"`
}
