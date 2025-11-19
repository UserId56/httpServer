package services

import (
	"fmt"
	"httpServer/models"
	"strings"
)

func GenerateCreateTableSQL(req models.CreateTableRequest) (string, bool, error) {
	cols := []string{
		"id SERIAL PRIMARY KEY",
	}
	for _, col := range req.Columns {
		var colString string
		if col.DataType == "ref" && col.ReferencedSheme != "" {
			colString += fmt.Sprintf(`"%s" %s`, col.ColumnName, "INT")
		} else {
			colString += fmt.Sprintf(`"%s" %s`, col.ColumnName, col.DataType)
		}
		//fmt.Printf("%+v\n", col)
		if col.DefaultValue != "" {
			switch col.DataType {
			case "TEXT":
				colString += fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue)
			case "INT", "BIGINT":
				colString += fmt.Sprintf(" DEFAULT %s", col.DefaultValue)
			case "BOOLEAN":
				colString += fmt.Sprintf(" DEFAULT %s", col.DefaultValue)
			case "TIMESTAMP", "DATE":
				colString += fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue)
			case "JSON":
				colString += fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue)
			case "ref":
				colString += fmt.Sprintf(" DEFAULT %s", col.DefaultValue)
			default:
				// Unsupported data type for default value
				return "", false, fmt.Errorf("Не верный тип данных %s", col.DataType)
			}

		}
		if col.NotNull {
			colString += " NOT NULL"
		}
		if col.IsRequired {
			colString += " UNIQUE"
		}
		if col.DataType == "ref" && col.ReferencedSheme != "" {
			colString += fmt.Sprintf(` REFERENCES "%s" (ID) ON DELETE SET NULL`, col.ReferencedSheme)
		}
		cols = append(cols, colString)
	}
	sql := fmt.Sprintf(`CREATE TABLE "%s" (%s);`, req.Name, strings.Join(cols, ", "))

	return sql, false, nil
}
