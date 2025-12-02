package services

import (
	"fmt"
	"httpServer/models"
)

func CheckTableName(nameTable string) bool {
	invalidNames := []string{"users", "roles", "dynamic_schemes", "dynamic_columns", "refresh_tokens"}
	for _, invalidName := range invalidNames {
		if nameTable == invalidName {
			return true
		}
	}
	return false
}

func CheckFieldsAndValue(obj map[string]interface{}, tableFields []models.DynamicColumns, create bool) error {
	for _, field := range tableFields {
		if create {
			if field.ColumnName == "id" || field.ColumnName == "created_at" || field.ColumnName == "updated_at" || field.ColumnName == "deleted_at" {
				continue
			}
		}
		var searchField bool
		for key, value := range obj {
			if key == field.ColumnName {
				searchField = true
				switch field.DataType {
				case "INT", "BIGINT", "ref":
					_, ok := value.(float64)
					if !ok {
						return fmt.Errorf("поле %s имеет неверный тип данных", key)
					}
					break
				case "TEXT", "JSON", "DATE", "TIMESTAMP":
					str, ok := value.(string)
					if !ok {
						return fmt.Errorf("поле %s имеет неверный тип данных", key)
					}
					if field.NotNull != nil && *field.NotNull && str == "" {
						return fmt.Errorf("поле %s не может быть пустым", key)
					}
					break
				case "BOOLEAN":
					_, ok := value.(bool)
					if !ok {
						return fmt.Errorf("поле %s имеет неверный тип данных", key)
					}
					break
				default:
					return fmt.Errorf("неизвестный тип данных поля %s", key)
				}
			}
		}
		if !searchField {
			for key, _ := range obj {
				notIn := false
				for _, fieldT := range tableFields {
					if key == fieldT.ColumnName {
						notIn = true
						break
					}
				}
				if !notIn {
					return fmt.Errorf("поле %s не существует в таблице", key)
				}
			}
			return fmt.Errorf("обязательное поле %s отсутствует", field.ColumnName)
		}
	}
	return nil
}
