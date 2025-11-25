package services

import (
	"fmt"
	"httpServer/models"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func SchemeHelperGetByName(db *gorm.DB, schemeName string) (*models.DynamicScheme, error) {
	var scheme models.DynamicScheme
	if err := db.Preload("Columns").Where("name = ?", schemeName).First(&scheme).Error; err != nil {
		return nil, err
	}
	return &scheme, nil
}

func GenerateUpdateTableSQL(columnsUpdate []*models.DynamicColumns, currentScheme *models.DynamicScheme) (string, []*models.DynamicColumns, []*models.DynamicColumns, error) {
	var newColumns []*models.DynamicColumns
	var oldColumns []*models.DynamicColumns
	var deleteColumns []*models.DynamicColumns
	SQLAlert := fmt.Sprintf("ALTER TABLE \"%s\" ", currentScheme.Name)
	SQLUpdate := fmt.Sprintf("UPDATE \"%s\" ", currentScheme.Name)
	isUpdate := false
	var resultSQL string
	for _, column := range columnsUpdate {
		if column.ID == 0 {
			column.DynamicTableID = currentScheme.ID
			newColumns = append(newColumns, column)
			continue
		}
		for _, currentCol := range currentScheme.Columns {
			if column.ID == currentCol.ID {
				column.DynamicTableID = currentScheme.ID
				oldColumns = append(oldColumns, column)
				if column.ColumnName != currentCol.ColumnName {
					SQLAlert += fmt.Sprintf("RENAME COLUMN \"%s\" TO \"%s\", ", currentCol.ColumnName, column.ColumnName)
				}
				if column.DataType != currentCol.DataType {
					var dataType string
					if column.DataType == "ref" && column.ReferencedScheme != "" {
						dataType = "INT"
					} else {
						dataType = column.DataType
					}
					SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" TYPE %s USING \"%s\"::%s, ", column.ColumnName, dataType, column.ColumnName, dataType)
				}
				if column.DefaultValue != currentCol.DefaultValue {
					if column.DefaultValue != "" {
						switch column.DataType {
						case "TEXT", "TIMESTAMP", "DATE", "JSON":
							SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" SET DEFAULT '%s', ", column.ColumnName, column.DefaultValue)
						case "INT", "BIGINT", "BOOLEAN", "ref":
							isInt, err := strconv.ParseInt(column.DefaultValue, 10, 64)
							if err != nil {
								return "", nil, nil, fmt.Errorf("не верный тип данных DEFAULT для типа %s: %s", column.DataType, column.DefaultValue)
							}
							SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" SET DEFAULT %d, ", column.ColumnName, isInt)
						default:
							return "", nil, nil, fmt.Errorf("не верный тип данных %s", column.DataType)
						}
					} else {
						SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" DROP DEFAULT, ", column.ColumnName)
					}
				}
				if *column.NotNull != *currentCol.NotNull {
					if column.NotNull != nil && *column.NotNull {
						SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" SET NOT NULL, ", column.ColumnName)
					} else {
						SQLAlert += fmt.Sprintf("ALTER COLUMN \"%s\" DROP NOT NULL, ", column.ColumnName)
					}
				}
				if *column.IsUnique != *currentCol.IsUnique {
					unName := fmt.Sprintf("uniq_%s_%s", currentScheme.Name, column.ColumnName)
					if column.IsUnique != nil && *column.IsUnique {
						SQLAlert += fmt.Sprintf("ADD CONSTRAINT \"%s\" UNIQUE (\"%s\"), ", unName, column.ColumnName)
					} else {

						SQLAlert += fmt.Sprintf("DROP CONSTRAINT IF EXISTS \"%s\", ", unName)
					}
				}
				if column.DataType == "ref" && column.ReferencedScheme != currentCol.ReferencedScheme {
					if currentCol.DataType == "ref" && currentCol.ReferencedScheme != "" {
						// Drop existing foreign key constraint
						SQLAlert += fmt.Sprintf("DROP CONSTRAINT IF EXISTS \"%s\", ", fmt.Sprintf("fk_%s_%s", currentScheme.Name, column.ReferencedScheme))
					}
					if column.ReferencedScheme != "" {
						// Add new foreign key constraint
						SQLUpdate += fmt.Sprintf("SET \"%s\" = NULL WHERE \"%s\" NOT IN (SELECT ID FROM \"%s\") AND \"%s\" IS NOT NULL, ", column.ColumnName, column.ColumnName, column.ReferencedScheme, column.ColumnName)
						isUpdate = true
						SQLAlert += fmt.Sprintf("ADD CONSTRAINT \"%s\" FOREIGN KEY (\"%s\") REFERENCES \"%s\" (ID) ON DELETE SET NULL NOT VALID, ", fmt.Sprintf("fk_%s_%s", currentScheme.Name, column.ReferencedScheme), column.ColumnName, column.ReferencedScheme)
					} else {
						return "", nil, nil, fmt.Errorf("пустая ссылка на коллекцию")
					}
				}
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
		if !found {
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
	if len(resultSQL) == len(SQLAlert) {
		resultSQL = ""
	}
	return resultSQL, newColumns, deleteColumns, nil
}

func GenerateCreateTableSQL(req models.CreateSchemeRequest, isAdd bool) (string, bool, error) {
	var cols []string
	if !isAdd {
		cols = append(cols, "id SERIAL PRIMARY KEY")
	}
	for _, col := range req.Columns {
		fmt.Printf("-!!!!!!%+v!!!!!-", col)
		var colString string
		var updateStr string
		if isAdd {
			updateStr = "ADD COLUMN "
		}
		if col.DataType == "ref" && col.ReferencedScheme != "" {
			colString += fmt.Sprintf(`%s"%s" %s`, updateStr, col.ColumnName, "INT")
		} else {
			colString += fmt.Sprintf(`%s"%s" %s`, updateStr, col.ColumnName, col.DataType)
		}
		//fmt.Printf("%+v\n", col)
		if col.DefaultValue != "" {
			switch col.DataType {
			case "TEXT":
				colString += fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue)
			case "INT", "BIGINT":
				isInt, err := strconv.ParseInt(col.DefaultValue, 10, 64)
				if err != nil {
					return "", false, fmt.Errorf("не верный тип данных дефолтного значения для типа %s: %s", col.DataType, col.DefaultValue)
				}
				colString += fmt.Sprintf(" DEFAULT %d", isInt)
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
				return "", false, fmt.Errorf("не верный тип данных %s", col.DataType)
			}

		}
		if col.NotNull != nil && *col.NotNull {
			colString += " NOT NULL"
		}
		if col.IsUnique != nil && *col.IsUnique {
			unName := fmt.Sprintf("uniq_%s_%s", req.Name, col.ColumnName)
			colString += fmt.Sprintf(" CONSTRAINT \"%s\" UNIQUE", unName)
		}
		if col.DataType == "ref" {
			if col.ReferencedScheme != "" {
				colString += fmt.Sprintf(` REFERENCES "%s" (ID) ON DELETE SET NULL`, col.ReferencedScheme)
			} else {
				return "", false, fmt.Errorf("пустая ссылка на коллекцию")
			}
		}
		cols = append(cols, colString)
	}
	if isAdd {
		sql := fmt.Sprintf(`ALTER TABLE "%s" %s;`, req.Name, strings.Join(cols, ", "))
		return sql, true, nil
	}
	sql := fmt.Sprintf(`CREATE TABLE "%s" (%s);`, req.Name, strings.Join(cols, ", "))

	return sql, false, nil
}

func CheckRefTables(columns []*models.DynamicColumns, db *gorm.DB) error {
	for _, col := range columns {
		if col.DataType == "ref" {
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
