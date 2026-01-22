package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
						if field.ReferencedScheme == "files" {
							arr, ok := value.([]interface{})
							if !ok {
								return fmt.Errorf("поле %s имеет неверный тип данных", key)
							}
							for _, v := range arr {
								_, ok := v.(string)
								if !ok {
									return fmt.Errorf("поле %s имеет неверный тип данных", key)
								}
							}
							break
						}
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
					if field.ReferencedScheme == "files" {
						_, ok := value.(string)
						if !ok {
							return fmt.Errorf("поле %s имеет неверный тип данных", key)
						}
						break
					}
					_, ok := value.(float64)
					if !ok {
						return fmt.Errorf("поле %s имеет неверный тип данных", key)
					}
					break
				case "TEXT", "STRING", "JSON", "DATE", "TIMESTAMPTZ":
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

func ExtractInt64Slice(fields []models.DynamicColumns, obj map[string]interface{}) map[string]interface{} {
	result := obj
	for _, field := range fields {
		if field.DataType == "ref" && field.IsMultiple != nil && *field.IsMultiple {
			val, ok := obj[field.ColumnName]
			if ok {
				val = strings.Trim(val.(string), "{}")
				if val == "" {
					obj[field.ColumnName] = []int64{}
				}

				// Делим по запятой и конвертируем
				parts := strings.Split(val.(string), ",")
				parsArr := make([]int64, len(parts))
				for i, p := range parts {
					parsArr[i], _ = strconv.ParseInt(p, 10, 64)
				}
				obj[field.ColumnName] = parsArr
			}

		}
	}
	return result
}

func ParsDataTime(fields []models.DynamicColumns, obj map[string]interface{}, timeZone *time.Location) (map[string]interface{}, error) {
	result := obj

	for _, field := range fields {
		switch field.DataType {
		case "TIMESTAMPTZ":
			if val, ok := obj[field.ColumnName]; ok {
				str, ok := val.(string)
				if !ok || str == "" {
					continue
				}
				t, err := time.ParseInLocation("2006-01-02T15:04", str, timeZone)
				if err != nil {
					return nil, fmt.Errorf("поле %s имеет неверный формат даты. Ожидается формат RFC3339: %v", field.ColumnName, err)
				}
				// сохраняем в UTC в формате RFC3339
				obj[field.ColumnName] = t.UTC().Format(time.RFC3339)
			}
		case "DATE":
			if val, ok := obj[field.ColumnName]; ok {
				str, ok := val.(string)
				if !ok || str == "" {
					continue
				}
				t, err := time.ParseInLocation("2006-01-02", str, timeZone)
				if err != nil {
					return nil, fmt.Errorf("поле %s имеет неверный формат даты. Ожидается формат YYYY-MM-DD: %v", field.ColumnName, err)
				}
				// Для DATE сохраняем только дату в указанной локации (без смещений)
				d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, timeZone)
				obj[field.ColumnName] = d.Format("2006-01-02")
			}
		}
	}

	return result, nil
}
