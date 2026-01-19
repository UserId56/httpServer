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
	IsMultiple       *bool          `gorm:"type:boolean;not null;default:false" json:"is_multiple"`
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
	OwnerID   *uint          `json:"owner_id"`
	Owner     *User          `gorm:"foreignKey:OwnerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"owner,omitempty"`

	Name        string            `gorm:"type:text;not null;uniqueIndex" json:"name"`
	DisplayName string            `gorm:"type:text;not null" json:"display_name"`
	ViewData    *ViewData         `gorm:"type:jsonb;serializer:json" json:"view_data,omitempty"`
	Columns     []*DynamicColumns `gorm:"foreignKey:DynamicTableID;constraint:OnDelete:CASCADE;" json:"columns"`
}

type ViewData struct {
	ShortView    string                  `json:"short_view"`
	HideMenu     bool                    `json:"hide_menu"`
	FieldOptions map[string]FieldOptions `json:"field_options"`
}

type FieldOptions struct {
	Hidden     bool       `gorm:"default:false" json:"hidden"`
	Filterable bool       `gorm:"default:false" json:"filterable"`
	Order      int        `json:"order"`
	PreValues  []PreValue `json:"pre_values"`
}

type PreValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
