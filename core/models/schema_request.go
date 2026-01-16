package models

import (
	"fmt"

	"gorm.io/datatypes"
)

type ColumnDefinition struct {
	ColumnName       string         `json:"column_name" binding:"required,identifier,min=2,max=64"`
	DisplayName      string         `json:"display_name" binding:"required,min=2,max=128"`
	DataType         string         `json:"data_type" binding:"required,oneof=TEXT STRING INT BIGINT BOOLEAN TIMESTAMP DATE JSON, ref"`
	ReferencedScheme string         `json:"referenced_scheme,omitempty"`
	IsMultiple       *bool          `json:"is_multiple"`
	IsUnique         *bool          `json:"is_unique"`
	NotNull          *bool          `json:"not_null"`
	DefaultValue     string         `json:"default_value,omitempty"`
	ValidationRules  datatypes.JSON `json:"validation_rules,omitempty"`
}

type CreateSchemeRequest struct {
	Name        string             `json:"name" binding:"required,identifier,min=2,max=64"`
	DisplayName string             `json:"display_name" binding:"required,min=2,max=128"`
	ViewData    *ViewData          `json:"view_data"`
	Columns     []ColumnDefinition `json:"columns" binding:"required,min=1"`
}

func ptrBool(data bool) *bool {
	return &data
}

func (ctr *CreateSchemeRequest) CreateDynamicTable(ownerId uint) *DynamicScheme {
	var columns []*DynamicColumns
	var defaultColumns = []*DynamicColumns{
		{ColumnName: "id", DisplayName: "ID", DataType: "BIGINT", NotNull: ptrBool(true), IsUnique: ptrBool(true)},
		{ColumnName: "created_at", DisplayName: "Дата создания", DataType: "TIMESTAMP", NotNull: ptrBool(true), DefaultValue: "NOW"},
		{ColumnName: "updated_at", DisplayName: "Дата изменения", DataType: "TIMESTAMP", NotNull: ptrBool(true), DefaultValue: "NOW"},
		{ColumnName: "deleted_at", DisplayName: "Дата удаления", DataType: "TIMESTAMP"},
	}
	columns = append(columns, defaultColumns...)
	var viewDataExists = ctr.ViewData != nil
	if !viewDataExists {
		ctr.ViewData = &ViewData{}
		ctr.ViewData.ShortView = "{id}"
	}
	for _, colDef := range ctr.Columns {
		if colDef.ColumnName == "title" || colDef.ColumnName == "name" {
			ctr.ViewData.ShortView = fmt.Sprintf("{%s}", colDef.ColumnName)
		}
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
			IsMultiple:       colDef.IsMultiple,
			IsUnique:         colDef.IsUnique,
			NotNull:          colDef.NotNull,
			DefaultValue:     colDef.DefaultValue,
			ValidationRules:  colDef.ValidationRules,
		}
		columns = append(columns, column)
	}
	for index, column := range columns {
		if !viewDataExists {
			var isFilterable bool
			if column.ColumnName != "id" || column.DataType != "BOOLEAN" || column.DataType != "JSON" {
				isFilterable = true
			}
			filedOptions := FieldOptions{
				Name:       column.ColumnName,
				Hidden:     false,
				Filterable: isFilterable,
				Order:      index + 1,
			}
			ctr.ViewData.FieldOptions = append(ctr.ViewData.FieldOptions, filedOptions)
		}
	}
	var ownerIdColumn = &DynamicColumns{
		ColumnName:       "owner_id",
		DisplayName:      "Автор",
		DataType:         "ref",
		ReferencedScheme: "users",
		IsMultiple:       ptrBool(false),
		NotNull:          ptrBool(false),
		IsUnique:         ptrBool(false),
	}
	columns = append(columns, ownerIdColumn)
	return &DynamicScheme{
		Name:        ctr.Name,
		DisplayName: ctr.DisplayName,
		ViewData:    ctr.ViewData,
		Columns:     columns,
		OwnerID:     &ownerId,
	}
}
