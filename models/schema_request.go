package models

import (
	"gorm.io/datatypes"
)

type ColumnDefinition struct {
	ColumnName       string         `json:"column_name" binding:"required,alphanum,min=2,max=64"`
	DisplayName      string         `json:"display_name" binding:"required,min=2,max=128"`
	DataType         string         `json:"data_type" binding:"required,oneof=TEXT INT BIGINT BOOLEAN TIMESTAMP DATE JSON, ref"`
	ReferencedScheme string         `json:"referenced_scheme,omitempty"`
	IsUnique         *bool          `json:"is_unique"`
	NotNull          *bool          `json:"not_null"`
	DefaultValue     string         `json:"default_value,omitempty"`
	ValidationRules  datatypes.JSON `json:"validation_rules,omitempty"`
}

type CreateSchemeRequest struct {
	Name        string             `json:"table_name" binding:"required,alphanum,min=2,max=64"`
	DisplayName string             `json:"display_name" binding:"required,min=2,max=128"`
	Columns     []ColumnDefinition `json:"columns" binding:"required,min=1"`
}

func ptrBool(data bool) *bool {
	return &data
}

func (ctr *CreateSchemeRequest) CreateDynamicTable() *DynamicScheme {
	var columns []*DynamicColumns
	var defaultColumns = []*DynamicColumns{
		{ColumnName: "id", DisplayName: "ID", DataType: "INT", NotNull: ptrBool(true), IsUnique: ptrBool(true)},
		{ColumnName: "created_at", DisplayName: "Дата создания", DataType: "TIMESTAMP", NotNull: ptrBool(true), DefaultValue: "NOW"},
		{ColumnName: "updated_at", DisplayName: "Дата изменения", DataType: "TIMESTAMP", NotNull: ptrBool(true), DefaultValue: "NOW"},
		{ColumnName: "deleted_at", DisplayName: "Дата удаления", DataType: "TIMESTAMP"},
	}
	columns = append(columns, defaultColumns...)
	for _, colDef := range ctr.Columns {
		var refScheme string
		if colDef.DataType != "ref" {
			refScheme = ""
		} else {
			refScheme = colDef.ReferencedScheme
		}
		column := &DynamicColumns{
			ColumnName:       colDef.ColumnName,
			DisplayName:      colDef.DisplayName,
			DataType:         colDef.DataType,
			ReferencedScheme: refScheme,
			IsUnique:         colDef.IsUnique,
			NotNull:          colDef.NotNull,
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
