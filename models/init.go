package models

func Models() []interface{} {
	return []interface{}{
		&Test{},
		&Role{},
		&User{},
		&RefreshToken{},
	}
}
