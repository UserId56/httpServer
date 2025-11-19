package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DynamicColumns struct {
	gorm.Model
	DynamicTableID  uint           `gorm:"not null;uniqueIndex:idx_table_column;d" binding:"required" json:"dynamic_table_id"`
	ColumnName      string         `gorm:"type:text;not null;uniqueIndex:idx_table_column" binding:"required,min=2" json:"column_name"`
	DisplayName     string         `gorm:"type:text;not null" binding:"required,min=2" json:"display_name"`
	DataType        string         `gorm:"type:text;not null" binding:"required" json:"data_type"`
	IsRequired      bool           `gorm:"type:boolean;not null" json:"is_required"`
	DefaultValue    string         `gorm:"type:text" json:"default_value"`
	ValidationRules datatypes.JSON `gorm:"type:jsonb;serializer:json" json:"validation_rules"`
}

type DynamicScheme struct {
	gorm.Model
	Name        string            `gorm:"type:text;not null;uniqueIndex" binding:"required,min=2" json:"name"`
	DisplayName string            `gorm:"type:text;not null" binding:"required,min=2" json:"display_name"`
	Columns     []*DynamicColumns `gorm:"foreignKey:DynamicTableID;constraint:OnDelete:CASCADE;" json:"columns"`
}
