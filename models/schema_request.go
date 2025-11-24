package models

import (
	"gorm.io/datatypes"
)

type ColumnDefinition struct {
	ColumnName       string         `json:"column_name" binding:"required,alphanum,min=2,max=64"`
	DisplayName      string         `json:"display_name" binding:"required,min=2,max=128"`
	DataType         string         `json:"data_type" binding:"required,oneof=TEXT INT BIGINT BOOLEAN TIMESTAMP DATE JSON, ref"`
	ReferencedScheme string         `json:"referenced_scheme,omitempty"`
	IsRequired       bool           `json:"is_required"`
	NotNull          bool           `json:"not_null"`
	DefaultValue     string         `json:"default_value,omitempty"`
	ValidationRules  datatypes.JSON `json:"validation_rules,omitempty"`
}

type CreateSchemeRequest struct {
	Name        string             `json:"table_name" binding:"required,alphanum,min=2,max=64"`
	DisplayName string             `json:"display_name" binding:"required,min=2,max=128"`
	Columns     []ColumnDefinition `json:"columns" binding:"required,min=1"`
}

func (ctr *CreateSchemeRequest) CreateDynamicTable() *DynamicScheme {
	var columns []*DynamicColumns
	for _, colDef := range ctr.Columns {
		column := &DynamicColumns{
			ColumnName:       colDef.ColumnName,
			DisplayName:      colDef.DisplayName,
			DataType:         colDef.DataType,
			ReferencedScheme: colDef.ReferencedScheme,
			IsRequired:       colDef.IsRequired,
			DefaultValue:     colDef.DefaultValue,
			ValidationRules:  colDef.ValidationRules,
		}
		columns = append(columns, column)
	}
	return &DynamicScheme{
		Name:        ctr.Name,
		DisplayName: ctr.DisplayName,
		Columns:     columns,
	}
}
