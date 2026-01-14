package services

import (
	"fmt"

	"github.com/UserId56/httpServer/core/models"
	"github.com/lib/pq"
)

func GenInclude(fields []models.DynamicColumns) []string {
	var result []string
	for _, field := range fields {
		result = append(result, fmt.Sprintf(`"%s"`, field.ColumnName))
	}
	return result
}

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
			if field.ColumnName == "id" || field.ColumnName == "created_at" || field.ColumnName == "updated_at" || field.ColumnName == "deleted_at" || field.ColumnName == "owner_id" {
				continue
			}
		} else {
			if field.ColumnName == "id" {
				continue
			}
		}
		var searchField bool
		for key, value := range obj {
			if key == field.ColumnName {
				searchField = true
				switch field.DataType {
				case "INT", "BIGINT", "ref", "FLOAT", "MONEY":
					if value == nil {
						if field.NotNull != nil && *field.NotNull {
							return fmt.Errorf("поле %s не может быть пустым", key)
						}
						break
					}
					if field.IsMultiple != nil && *field.IsMultiple {
						arr, ok := value.([]interface{})
						if !ok {
							return fmt.Errorf("поле %s имеет неверный тип данных", key)
						}
						for _, v := range arr {
							_, ok := v.(float64)
							if !ok {
								return fmt.Errorf("поле %s имеет неверный тип данных", key)
							}
						}
						break
					}
					_, ok := value.(float64)
					if !ok {
						return fmt.Errorf("поле %s имеет неверный тип данных", key)
					}
					break
				case "TEXT", "JSON", "DATE", "TIMESTAMP":
					if value == nil {
						if field.NotNull != nil && *field.NotNull {
							return fmt.Errorf("поле %s не может быть пустым", key)
						}
						break
					}
					str, ok := value.(string)
					if !ok {
						return fmt.Errorf("поле %s имеет неверный тип данных", key)
					}
					if field.NotNull != nil && *field.NotNull && str == "" {
						return fmt.Errorf("поле %s не может быть пустым", key)
					}
					break
				case "BOOLEAN":
					if value == nil {
						if field.NotNull != nil && *field.NotNull {
							return fmt.Errorf("поле %s не может быть пустым", key)
						}
						break
					}
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
			if create && field.NotNull != nil && *field.NotNull {
				return fmt.Errorf("обязательное поле %s отсутствует", field.ColumnName)
			}
		}
	}
	return nil
}

func ParsIntField(fields []models.DynamicColumns, obj map[string]interface{}) map[string]interface{} {
	var result = obj
	for _, field := range fields {
		if field.DataType == "ref" && field.IsMultiple != nil && *field.IsMultiple {
			if val, ok := obj[field.ColumnName]; ok {
				var intResult []int64
				arr, ok := val.([]interface{})
				if ok {
					for _, v := range arr {
						if num, ok := v.(float64); ok {
							intResult = append(intResult, int64(num))
						}
					}
				}
				result[field.ColumnName] = pq.Array(intResult)
			}
		}
	}
	return result
}
