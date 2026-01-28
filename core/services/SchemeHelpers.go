package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/UserId56/httpServer/core/models"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func SchemeHelperGetByName(db *gorm.DB, schemeName string) (*models.DynamicScheme, error) {
	var scheme models.DynamicScheme
	if err := db.Preload("Columns").Where("name = ?", schemeName).First(&scheme).Error; err != nil {
		return nil, err
	}
	return &scheme, nil
}

func formatUsingAndDefault(column *models.DynamicColumns, dataType string) (usingExpr string, defaultExpr string, err error) {
	if column.DefaultValue == "" {
		return "NULL", "NULL", nil // USING NULL для существующих значений, DROP DEFAULT будет сформирован отдельно
	}
	switch column.DataType {
	case "TEXT", "TIMESTAMP", "DATE", "JSON":
		escaped := strings.ReplaceAll(column.DefaultValue, "'", "''")
		using := fmt.Sprintf("'%s'::%s", escaped, dataType)
		return using, using, nil
	case "BOOLEAN":
		val := strings.ToLower(column.DefaultValue)
		if val != "true" && val != "false" {
			val = "false"
		}
		using := fmt.Sprintf("%s::%s", val, dataType)
		return using, using, nil
	case "INT", "BIGINT", "ref":
		isInt, e := strconv.ParseInt(column.DefaultValue, 10, 64)
		if e != nil {
			return "", "", fmt.Errorf("не верный тип данных DEFAULT для %s: %s", column.DataType, column.DefaultValue)
		}
		using := fmt.Sprintf("%d::%s", isInt, dataType)
		return using, using, nil
	case "FLOAT", "MONEY":
		isFloat, e := strconv.ParseFloat(column.DefaultValue, 64)
		if e != nil {
			return "", "", fmt.Errorf("не верный тип данных DEFAULT для %s: %s", column.DataType, column.DefaultValue)
		}
		using := fmt.Sprintf("%f::%s", isFloat, dataType)
		return using, using, nil
	default:
		return "", "", fmt.Errorf("не поддерживаемый тип %s", column.DataType)
	}
}

func GenerateUpdateTableSQL(columnsUpdate []*models.DynamicColumns, currentScheme *models.DynamicScheme) (string, []*models.DynamicColumns, []*models.DynamicColumns, error) {
	var newColumns []*models.DynamicColumns
	var oldColumns []*models.DynamicColumns
	var deleteColumns []*models.DynamicColumns
	SQLAlert := fmt.Sprintf("ALTER TABLE \"%s\" ", currentScheme.Name)
	defStringLen := len(SQLAlert)
	SQLUpdate := fmt.Sprintf("UPDATE \"%s\" ", currentScheme.Name)
	isUpdate := false
	var resultSQL string
	for _, column := range columnsUpdate {
		if column.ColumnName == "id" || column.ColumnName == "created_at" || column.ColumnName == "updated_at" || column.ColumnName == "deleted_at" || column.ColumnName == "owner_id" {
			continue
		}
		if column.ID == 0 {
			column.DynamicTableID = currentScheme.ID
			newColumns = append(newColumns, column)
			continue
		}
		for _, currentCol := range currentScheme.Columns {
			if column.ID == currentCol.ID {
				column.DynamicTableID = currentScheme.ID
				if column.ColumnName != currentCol.ColumnName {
					//SQLAlert += fmt.Sprintf("RENAME COLUMN \"%s\" TO \"%s\"; ALTER TABLE \"%s\" ", currentCol.ColumnName, column.ColumnName, currentScheme.Name)
					SQLAlert += fmt.Sprintf("RENAME COLUMN \"%s\" TO \"%s\"; ALTER TABLE \"%s\" ", currentCol.ColumnName, column.ColumnName, currentScheme.Name)
				}
				if column.DataType != currentCol.DataType {
					validate := validator.New()
					err := validate.Var(column.DataType, "oneof=TEXT STRING INT BIGINT BOOLEAN TIMESTAMP DATE JSON FLOAT MONEY ref")
					if err != nil {
						return "", nil, nil, fmt.Errorf("не верный тип данных %s", column.DataType)
					}
					var dataType string
					if column.DataType == "ref" && column.ReferencedScheme != "" {
						dataType = "BIGINT"
						if column.IsMultiple != nil && *column.IsMultiple {
							dataType = "BIGINT[]"
						}
						if column.ReferencedScheme == "files" {
							dataType = "TEXT"
							if column.IsMultiple != nil && *column.IsMultiple {
								dataType = "JSONB"
							}
						}
					} else {
						dataType = column.DataType
					}
					//Пустая ссылка если меняется тип данных
					//if column.DataType != "ref" && currentCol.DataType == "ref" {
					//	SQLAlert += fmt.Sprintf("DROP CONSTRAINT IF EXISTS \"%s\", ", fmt.Sprintf("fk_%s_%s_%s", currentScheme.Name, currentCol.ReferencedScheme, currentCol.ColumnName))
					//	column.ReferencedScheme = ""
					//}
					if column.DataType == "FLOAT" {
						dataType = "DOUBLE PRECISION"
					}
					if column.DataType == "MONEY" {
						dataType = "NUMERIC(19,4)"
					}
					if column.DataType == "STRING" {
						dataType = "TEXT"
					}
					usingExpr, defaultExpr, err := formatUsingAndDefault(column, dataType)
					if err != nil {
						return "", nil, nil, err
					}
					SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" TYPE %s USING %s, ALTER COLUMN \"%s\" SET DEFAULT %s, ", column.ColumnName, dataType, usingExpr, column.ColumnName, defaultExpr)
				}
				if column.DefaultValue != currentCol.DefaultValue {
					if column.DefaultValue != "" {
						switch column.DataType {
						case "TEXT", "STRING", "TIMESTAMPTZ", "DATE", "JSON":
							SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" SET DEFAULT '%s', ", column.ColumnName, column.DefaultValue)
						case "INT", "BIGINT", "BOOLEAN", "ref":
							if column.ReferencedScheme != "files" {
								isInt, err := strconv.ParseInt(column.DefaultValue, 10, 64)
								if err != nil {
									return "", nil, nil, fmt.Errorf("не верный тип данных DEFAULT для типа %s: %s", column.DataType, column.DefaultValue)
								}
								SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" SET DEFAULT %d, ", column.ColumnName, isInt)
							} else {
								SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" SET DEFAULT '%s', ", column.ColumnName, column.DefaultValue)
							}
						case "FLOAT", "MONEY":
							isFloat, err := strconv.ParseFloat(column.DefaultValue, 64)
							if err != nil {
								return "", nil, nil, fmt.Errorf("не верный тип данных DEFAULT для типа %s: %s", column.DataType, column.DefaultValue)
							}
							SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" SET DEFAULT %f, ", column.ColumnName, isFloat)

						default:
							return "", nil, nil, fmt.Errorf("не верный тип данных %s", column.DataType)
						}
					} else {
						SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" DROP DEFAULT, ", column.ColumnName)
					}
				}
				if column.NotNull != nil && *column.NotNull {
					SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" SET NOT NULL, ", column.ColumnName)
				} else {
					boolFalse := false
					column.NotNull = &boolFalse
					SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" DROP NOT NULL, ", column.ColumnName)
				}

				unName := fmt.Sprintf("uniq_%s_%s", currentScheme.Name, column.ColumnName)
				unNameOld := fmt.Sprintf("uniq_%s_%s", currentScheme.Name, currentCol.ColumnName)
				if column.IsUnique != nil && *column.IsUnique {
					SQLAlert += fmt.Sprintf("ADD CONSTRAINT \"%s\" UNIQUE (\"%s\"), ", unName, column.ColumnName)
				} else {
					boolFalse := false
					column.IsUnique = &boolFalse
					SQLAlert += fmt.Sprintf("DROP CONSTRAINT IF EXISTS \"%s\", ", unNameOld)
				}
				if column.DataType == "ref" && *column.IsMultiple != *currentCol.IsMultiple {
					if column.IsMultiple != nil && *column.IsMultiple {
						if column.ReferencedScheme == "files" {
							SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" TYPE JSONB USING (CASE WHEN \"%s\" IS NULL OR \"%s\" = 'null' THEN '[]'::jsonb ELSE to_jsonb(ARRAY[\"%s\"]) END), ", column.ColumnName, column.ColumnName, column.ColumnName, column.ColumnName)
						} else {
							SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" TYPE BIGINT[] USING ARRAY[\"%s\"]::BIGINT[], ", column.ColumnName, column.ColumnName)
						}
					} else {
						if column.ReferencedScheme == "files" {
							SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" TYPE TEXT USING (NULLIF((\"%s\"::JSONB->>0), 'null')), ", column.ColumnName, column.ColumnName)
						} else {
							SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" TYPE INT USING (CASE WHEN cardinality(\"%s\") >= 1 THEN (\"%s\")[1] ELSE NULL END), ", column.ColumnName, column.ColumnName, column.ColumnName)
						}
					}
				}
				if column.DataType == "ref" && column.ReferencedScheme != currentCol.ReferencedScheme {
					//if currentCol.DataType == "ref" && currentCol.ReferencedScheme != "" {
					//	// Drop existing foreign key constraint
					//	SQLAlert += fmt.Sprintf("DROP CONSTRAINT IF EXISTS \"%s\", ", fmt.Sprintf("fk_%s_%s_%s", currentScheme.Name, column.ReferencedScheme, currentCol.ColumnName))
					//}
					if column.ReferencedScheme != "" {
						// Add new foreign key constraint
						SQLUpdate += fmt.Sprintf("SET \"%s\" = NULL, ", column.ColumnName)
						isUpdate = true
						//SQLAlert += fmt.Sprintf("ADD CONSTRAINT \"%s\" FOREIGN KEY (\"%s\") REFERENCES \"%s\" (ID) ON DELETE SET NULL NOT VALID, ", fmt.Sprintf("fk_%s_%s_%s", currentScheme.Name, column.ReferencedScheme, column.ColumnName), column.ColumnName, column.ReferencedScheme)
					} else {
						return "", nil, nil, fmt.Errorf("пустая ссылка на коллекцию")
					}
				}
				oldColumns = append(oldColumns, column)
			}
		}
	}
	for _, currentCol := range currentScheme.Columns {
		found := false
		for _, column := range columnsUpdate {
			if column.ID == currentCol.ID {
				found = true
				break
			}
		}
		if !found && currentCol.ColumnName != "id" && currentCol.ColumnName != "created_at" && currentCol.ColumnName != "updated_at" && currentCol.ColumnName != "deleted_at" && currentCol.ColumnName != "owner_id" {
			deleteColumns = append(deleteColumns, currentCol)
			SQLAlert += fmt.Sprintf("DROP COLUMN \"%s\", ", currentCol.ColumnName)
		}
	}
	SQLAlert = strings.TrimSuffix(SQLAlert, ", ") + ";"
	SQLUpdate = strings.TrimSuffix(SQLUpdate, ", ") + ";"
	if isUpdate {
		resultSQL = SQLAlert + SQLUpdate
	} else {
		resultSQL = SQLAlert
	}
	currentScheme.Columns = oldColumns
	if len(resultSQL) == defStringLen {
		resultSQL = ""
	}
	return resultSQL, newColumns, deleteColumns, nil
}

func GenerateCreateTableSQL(req models.CreateSchemeRequest, isAdd bool) (string, bool, error) {
	var cols []string
	var colRef []string
	if !isAdd {
		cols = append(cols, "id SERIAL PRIMARY KEY", `"created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP`, `"updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP`, `"deleted_at" TIMESTAMPTZ`, `"owner_id" BIGINT`)
	}
	for _, col := range req.Columns {
		var colString string
		var updateStr string
		if isAdd {
			updateStr = "ADD COLUMN "
		}
		if col.DataType == "ref" && col.ReferencedScheme != "" {
			typeRef := "BIGINT"
			if col.IsMultiple != nil && *col.IsMultiple {
				typeRef = "BIGINT[]"
			}
			if col.ReferencedScheme == "files" {
				typeRef = "TEXT"
				if col.IsMultiple != nil && *col.IsMultiple {
					typeRef = "JSONB"
				}
			}
			colString += fmt.Sprintf(`%s"%s" %s`, updateStr, col.ColumnName, typeRef)
		} else {
			dataType := col.DataType
			if col.DataType == "FLOAT" {
				dataType = "DOUBLE PRECISION"
			}
			if col.DataType == "MONEY" {
				dataType = "NUMERIC(19,4)"
			}
			if col.DataType == "STRING" {
				dataType = "TEXT"
			}
			colString += fmt.Sprintf(`%s"%s" %s`, updateStr, col.ColumnName, dataType)
		}
		if col.DefaultValue != "" {
			switch col.DataType {
			case "TEXT", "STRING":
				colString += fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue)
			case "INT", "BIGINT":
				isInt, err := strconv.ParseInt(col.DefaultValue, 10, 64)
				if err != nil {
					return "", false, fmt.Errorf("не верный тип данных дефолтного значения для типа %s: %s", col.DataType, col.DefaultValue)
				}
				colString += fmt.Sprintf(" DEFAULT %d", isInt)
			case "FLOAT", "MONEY":
				isFloat, err := strconv.ParseFloat(col.DefaultValue, 64)
				if err != nil {
					return "", false, fmt.Errorf("не верный тип данных дефолтного значения для типа %s: %s", col.DataType, col.DefaultValue)
				}
				colString += fmt.Sprintf(" DEFAULT %f", isFloat)
			case "BOOLEAN":
				colString += fmt.Sprintf(" DEFAULT %s", col.DefaultValue)
			case "TIMESTAMPTZ", "DATE":
				colString += fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue)
			case "JSON":
				colString += fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue)
			case "ref":
				//colString += fmt.Sprintf(\" DEFAULT %s", col.DefaultValue)
				return "", false, fmt.Errorf("для поля ref не поддерживается дефолтное значение")
			default:
				// Unsupported data type for default value
				return "", false, fmt.Errorf("не верный тип данных дефолтного значения для типа %s: %s", col.DataType, col.DefaultValue)
			}

		}
		if col.NotNull != nil && *col.NotNull {
			colString += " NOT NULL"
		}
		if col.IsUnique != nil && *col.IsUnique {
			unName := fmt.Sprintf("uniq_%s_%s", req.Name, col.ColumnName)
			colString += fmt.Sprintf(" CONSTRAINT \"%s\" UNIQUE", unName)
		}
		//if col.DataType == "ref" {
		//	if col.ReferencedScheme != "" {
		//		colRef = append(colRef, fmt.Sprintf(`CONSTRAINT "%s" FOREIGN KEY ("%s") REFERENCES "%s" (ID) ON DELETE SET NULL`, fmt.Sprintf("fk_%s_%s_%s", req.Name, col.ReferencedScheme, col.ColumnName), col.ColumnName, col.ReferencedScheme))
		//	} else {
		//		return "", false, fmt.Errorf("пустая ссылка на коллекцию")
		//	}
		//}
		cols = append(cols, colString)
	}
	cols = append(cols, colRef...)
	if isAdd {
		sql := fmt.Sprintf(`ALTER TABLE "%s" %s;`, req.Name, strings.Join(cols, ", "))
		return sql, true, nil
	}
	sql := fmt.Sprintf(`CREATE TABLE "%s" (%s);`, req.Name, strings.Join(cols, ", "))

	return sql, false, nil
}

func CheckRefTables(columns []*models.DynamicColumns, db *gorm.DB, currentTableName string) error {
	for _, col := range columns {
		if col.DataType == "ref" {
			if col.ReferencedScheme == currentTableName {
				continue
			}
			var count int64
			if err := db.Model(&models.DynamicScheme{}).Where("name = ?", col.ReferencedScheme).Count(&count).Error; err != nil {
				return fmt.Errorf("ошибка проверки ссылки на таблицу %s: %v", col.ReferencedScheme, err)
			}
			if count == 0 {
				return fmt.Errorf("поле %s ссылается на несуществующую таблицу %s", col.ColumnName, col.ReferencedScheme)
			}
		}
	}
	return nil
}
