package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DynamicColumns struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	DynamicTableID   uint           `gorm:"not null;uniqueIndex:idx_table_column;d" json:"dynamic_table_id"`
	ColumnName       string         `gorm:"type:text;not null;uniqueIndex:idx_table_column" json:"column_name"`
	DisplayName      string         `gorm:"type:text;not null" json:"display_name"`
	DataType         string         `gorm:"type:text;not null" json:"data_type"`
	ReferencedScheme string         `gorm:"type:text" json:"referenced_scheme"`
	IsUnique         *bool          `gorm:"type:boolean;not null;default:false" json:"is_unique"`
	NotNull          *bool          `gorm:"type:boolean;not null;default:false" json:"not_null"`
	DefaultValue     string         `gorm:"type:text" json:"default_value"`
	ValidationRules  datatypes.JSON `gorm:"type:jsonb;serializer:json" json:"validation_rules"`
}

type DynamicScheme struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name        string            `gorm:"type:text;not null;uniqueIndex" json:"name"`
	DisplayName string            `gorm:"type:text;not null" json:"display_name"`
	Columns     []*DynamicColumns `gorm:"foreignKey:DynamicTableID;constraint:OnDelete:CASCADE;" json:"columns"`
}
