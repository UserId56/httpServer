package models

type WhereParamField struct {
	Field    string      `json:"field"`
	Value    interface{} `json:"value"`
	Operator string      `json:"operator"`
}

type AndWhere struct {
	And []interface{} `json:"and"`
}

type OrWhere struct {
	Or []interface{} `json:"or"`
}

type Order struct {
	Field     string `json:"field" binding:"required"`
	Direction string `json:"direction" binding:"required,oneof=asc desc"`
}

type Search struct {
	Fields []string `json:"fields" binding:"required"`
	Query  string   `json:"query" binding:"required"`
}

type Query struct {
	Include []string      `json:"include,omitempty"`
	Where   []interface{} `json:"where,omitempty"`
	Order   []Order       `json:"order,omitempty"`
	Take    int           `json:"take,omitempty"`
	Skip    int           `json:"skip,omitempty"`
	Count   bool          `json:"count,omitempty"`
}

func (q *Query) IncludeBox() {
	for i, field := range q.Include {
		q.Include[i] = "\"" + field + "\""
	}
}
