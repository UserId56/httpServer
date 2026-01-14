package services

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/UserId56/httpServer/core/models"
)

func getOperator(operator string) (string, error) {
	switch operator {
	case "eq":
		return "=", nil
	case "ne":
		return "<>", nil
	case "gt":
		return ">", nil
	case "lt":
		return "<", nil
	case "gte":
		return ">=", nil
	case "lte":
		return "<=", nil
	case "in":
		return "IN", nil
	case "notin":
		return "NOT IN", nil
	case "null":
		return "IS", nil
	case "notNull":
		return "IS NOT", nil

	default:
		return "", fmt.Errorf("неизвестный оператор: %s", operator)
	}
}

//"where": [
//        {
//            "or": [
//                {
//                    "field": "deleted_at",
//                    "operator": "notNull"
//                },
//                {
//                    "field": "id",
//                    "operator": "eq",
//                    "value": "2"
//                }
//            ]
//        },
//        {
//            "query": "bor"
//        }
//    ]

func OrderGeneration(listOrder []models.Order, fields []models.DynamicColumns) (string, error) {
	if len(listOrder) == 0 {
		return "", nil
	}
	defaultFields := []models.DynamicColumns{
		{ColumnName: "id", DataType: "INT"},
		{ColumnName: "created_at", DataType: "TIMESTAMP"},
		{ColumnName: "updated_at", DataType: "TIMESTAMP"},
		{ColumnName: "deleted_at", DataType: "TIMESTAMP"},
	}
	fields = append(fields, defaultFields...)
	result := ""
	for index, order := range listOrder {
		if order.Field == "" {
			return "", fmt.Errorf("поле для сортировки не указано")
		}
		fieldFound := false
		for _, field := range fields {
			if field.ColumnName == order.Field {
				fieldFound = true
				break
			}
		}
		if !fieldFound {
			return "", fmt.Errorf("поле для сортировки %s не найдено в схеме", order.Field)
		}
		switch order.Direction {
		case "asc", "desc":
			if index == 0 {
				result += fmt.Sprintf(`"%s" %s`, order.Field, order.Direction)
			} else {
				result += fmt.Sprintf(`, "%s" %s`, order.Field, order.Direction)
			}
		default:
			return "", fmt.Errorf("неизвестное направление сортировки: %s", order.Direction)
		}
	}
	return result, nil
}

func WhereGeneration(dataWhere []interface{}, fields []models.DynamicColumns, operator string) (string, []interface{}, error) {
	if len(dataWhere) == 0 {
		return "", nil, nil
	}
	defaultFields := []models.DynamicColumns{
		{ColumnName: "id", DataType: "INT"},
		{ColumnName: "created_at", DataType: "TIMESTAMP"},
		{ColumnName: "updated_at", DataType: "TIMESTAMP"},
		{ColumnName: "deleted_at", DataType: "TIMESTAMP"},
	}
	fields = append(fields, defaultFields...)
	result := "( "
	var arg []interface{}
	for index, value := range dataWhere {
		var dataWhereParamField models.WhereParamField
		jsonData, err := json.Marshal(value)
		if err != nil {
			return "", nil, fmt.Errorf("ошибка маршалинга данных условия")
		}
		err = json.Unmarshal(jsonData, &dataWhereParamField)
		if err == nil && dataWhereParamField.Field != "" {
			//fmt.Println(dataWhereParamField)
			fieldFound := false
			for _, field := range fields {
				if field.ColumnName == dataWhereParamField.Field {
					fieldFound = true
					operatorField, err := getOperator(dataWhereParamField.Operator)
					if err != nil {
						return "", nil, err
					}
					switch field.DataType {
					case "INT", "BIGINT", "ref":
						if dataWhereParamField.Value != nil {
							isInt, err := strconv.ParseInt(fmt.Sprintf("%v", dataWhereParamField.Value), 10, 64)
							if err != nil {
								return "", nil, fmt.Errorf("не верный тип данных для поля %s, ожидается INT", dataWhereParamField.Field)
							}
							result += fmt.Sprintf(`"%s" %s ?`, dataWhereParamField.Field, operatorField)
							arg = append(arg, isInt)
						} else {
							result += fmt.Sprintf(`"%s" %s NULL`, dataWhereParamField.Field, operatorField)
						}
					case "FLOAT", "DOUBLE":
						if dataWhereParamField.Value != nil {
							isFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", dataWhereParamField.Value), 64)
							if err != nil {
								return "", nil, fmt.Errorf("не верный тип данных для поля %s, ожидается FLOAT", dataWhereParamField.Field)
							}
							result += fmt.Sprintf(`"%s" %s ?`, dataWhereParamField.Field, operatorField)
							arg = append(arg, isFloat)
						} else {
							result += fmt.Sprintf(`"%s" %s NULL`, dataWhereParamField.Field, operatorField)
						}
					case "BOOLEAN":
						if dataWhereParamField.Value != nil {
							isBool, ok := dataWhereParamField.Value.(bool)
							if !ok {
								return "", nil, fmt.Errorf("не верный тип данных для поля %s, ожидается BOOLEAN", dataWhereParamField.Field)
							}
							result += fmt.Sprintf(`"%s" %s ?`, dataWhereParamField.Field, operatorField)
							arg = append(arg, isBool)
						} else {
							result += fmt.Sprintf(`"%s" %s NULL`, dataWhereParamField.Field, operatorField)
						}
					case "TIMESTAMP", "DATE":
						if dataWhereParamField.Value != nil {
							result += fmt.Sprintf(`"%s" %s ?`, dataWhereParamField.Field, operatorField)
							arg = append(arg, dataWhereParamField.Value)
						} else {
							result += fmt.Sprintf(`"%s" %s NULL`, dataWhereParamField.Field, operatorField)
						}
					case "TEXT", "JSON":
						if dataWhereParamField.Value != nil {
							result += fmt.Sprintf(`"%s" %s ?`, dataWhereParamField.Field, operatorField)
							arg = append(arg, fmt.Sprintf("%v", dataWhereParamField.Value))
						} else {
							result += fmt.Sprintf(`"%s" %s NULL`, dataWhereParamField.Field, operatorField)
						}
					default:
						return "", nil, fmt.Errorf("неизвестный тип данных поля %s: %s", value.(models.WhereParamField).Field, field.DataType)
					}
					if index < len(dataWhere)-1 {
						result += fmt.Sprintf(" %s ", operator)
					} else {
						result += " )"
					}
					break
				}
			}
			if !fieldFound {
				return "", nil, fmt.Errorf("поле %s не найдено в схеме", dataWhereParamField.Field)
			}
			continue
		}
		//dataWhereParamField, ok := value.(models.WhereParamField)
		//if ok {}
		var dataAndWhere models.AndWhere
		//dataAndWhere, ok := value.(models.AndWhere)
		err = json.Unmarshal(jsonData, &dataAndWhere)
		//fmt.Println(dataAndWhere)
		if err == nil && dataAndWhere.And != nil {
			subResult, subArg, err := WhereGeneration(dataAndWhere.And, fields, "AND")
			if err != nil {
				return "", nil, err
			}
			arg = append(arg, subArg...)
			if index < len(dataWhere)-1 {
				result += subResult + " " + operator + " "
			} else {
				result += subResult + " )"
			}
			continue
		}
		var dataOrWhere models.OrWhere
		//dataOrWhere, ok := value.(models.OrWhere)
		err = json.Unmarshal(jsonData, &dataOrWhere)
		if err == nil && dataOrWhere.Or != nil {
			subResult, subArg, err := WhereGeneration(dataOrWhere.Or, fields, "OR")
			if err != nil {
				return "", nil, err
			}
			arg = append(arg, subArg...)
			if index < len(dataWhere)-1 {
				result += subResult + " " + operator + " "
			} else {
				result += subResult + " )"
			}
			continue
		}
		var dataSearch models.Search
		err = json.Unmarshal(jsonData, &dataSearch)
		//fmt.Printf("%+v\n", dataSearch)
		//fmt.Printf("%+v\n", dataSearch.Fields != nil)
		//fmt.Printf("%+v\n", dataSearch.Query != "")
		//fmt.Printf("%+v\n", err)
		if err == nil && dataSearch.Query != "" {
			subResult := "( "
			if len(dataSearch.Fields) == 0 {
				for _, field := range fields {
					if field.DataType == "TEXT" || field.DataType == "STRING" {
						dataSearch.Fields = append(dataSearch.Fields, field.ColumnName)
					}
				}
			}
			for fIndex, fieldName := range dataSearch.Fields {
				fieldFound := false
				for _, field := range fields {
					if field.ColumnName == fieldName {
						fieldFound = true
						subResult += fmt.Sprintf(`"%s" ILIKE ?`, fieldName)
						arg = append(arg, "%"+dataSearch.Query+"%")
						if fIndex < len(dataSearch.Fields)-1 {
							subResult += " OR "
						} else {
							subResult += " )"
						}
						break
					}
				}
				if !fieldFound {
					return "", nil, fmt.Errorf("поле для поиска %s не найдено в схеме", fieldName)
				}
			}
			if index < len(dataWhere)-1 {
				result += subResult + " " + operator + " "
			} else {
				result += subResult + " )"
			}
			continue
		}
		return "", nil, fmt.Errorf("неизвестный тип условия в where")
	}
	return result, arg, nil
}
