package models

type Settings struct {
	ID    uint          `gorm:"primaryKey"`
	Value SettingsValue `gorm:"type:jsonb;serializer:json;not null"`
}

type SettingsValue struct {
	Lang          []string `json:"lang"`
	DefaultRoleId uint     `json:"default_role_id"`
	TimeZone      int      `json:"time_zone"`
	Style         string   `json:"style"`
}
