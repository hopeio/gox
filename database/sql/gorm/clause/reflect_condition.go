package clause

import (
	"reflect"
	"strings"

	"github.com/hopeio/gox/database/sql"
	"github.com/hopeio/gox/reflect/structtag"
	stringsx "github.com/hopeio/gox/strings"
	"gorm.io/gorm/clause"
)

/*
		type ReportList struct {
		PaginationEmbedded `sqlcond:"-"`
		LoadingTime        *Range[time.Time]
		UserId             int
		CarId              int
		TaskId             int
		RouteId            int
		Diff               float64 `sqlcond:"-"`
		Outlier            bool    `sqlcond:"-"`
	}
*/
var paginationEmbeddedType = reflect.TypeFor[PaginationEmbedded]()
var paginationEmbeddedPtrType = reflect.TypeFor[*PaginationEmbedded]()
var paginationType = reflect.TypeFor[Pagination]()
var paginationPtrType = reflect.TypeFor[*Pagination]()

func AndConditionBy(param any) clause.Expression {
	return andConditionBy(reflect.ValueOf(param))
}

func OrConditionBy(param any) clause.Expression {
	return orConditionBy(reflect.ValueOf(param))
}

func NotConditionBy(param any) clause.Expression {
	return notConditionBy(reflect.ValueOf(param))
}

func andConditionBy(param reflect.Value) clause.Expression {
	conditions := conditionsBy(param)
	if len(conditions) > 0 {
		return clause.AndConditions{Exprs: conditions}
	}
	return nil
}

func orConditionBy(param reflect.Value) clause.Expression {
	conditions := conditionsBy(param)

	if len(conditions) > 0 {
		return clause.OrConditions{Exprs: conditions}
	}
	return nil
}

func notConditionBy(param reflect.Value) clause.Expression {
	conditions := conditionsBy(param)
	if len(conditions) > 0 {
		return clause.NotConditions{Exprs: conditions}
	}
	return nil
}

func ConditionsBy(param any) []clause.Expression {
	return conditionsBy(reflect.ValueOf(param))
}

func conditionsByImpl(param reflect.Value) (clause.Expression, bool) {
	t := param.Type()
	if t.Implements(conditionExprType) {
		if condition := param.Interface().(ConditionExpr).Condition(); condition != nil {
			return condition, true
		}
		return nil, true
	}

	if param.CanAddr() && param.Addr().Type().Implements(conditionExprType) {
		if condition := param.Addr().Interface().(ConditionExpr).Condition(); condition != nil {
			return condition, true
		}
	}
	return nil, false
}
func conditionsBy(param reflect.Value) []clause.Expression {
	condition, impl := conditionsByImpl(param)
	if impl {
		if condition != nil {
			return []clause.Expression{condition}
		}
		return nil
	}
	t := param.Type()
	if t.Kind() == reflect.Ptr && param.IsNil() {
		return nil
	}
	vi := reflect.Indirect(param)
	ti := vi.Type()

	kind := vi.Kind()
	if kind != reflect.Struct && kind != reflect.Map && kind != reflect.Array && kind != reflect.Slice {
		return nil
	}
	if kind != reflect.Array && kind != reflect.Struct && vi.IsNil() {
		return nil
	}
	if ti == paginationEmbeddedType || ti == paginationType ||
		ti == paginationPtrType || ti == paginationEmbeddedPtrType {
		return nil
	}
	if kind == reflect.Map {
		var conditions []clause.Expression
		if ti.Key().Kind() == reflect.String {
			iter := vi.MapRange()
			for iter.Next() {
				conditions = append(conditions, clause.Eq{Column: stringsx.CamelToSnake(iter.Key().String()), Value: iter.Value()})
			}
		}
		return conditions
	}
	var conditions []clause.Expression
	for i := range vi.NumField() {
		field := vi.Field(i)
		fieldKind := field.Kind()
		structField := ti.Field(i)
		empty := field.IsZero()
		tag, ok := structField.Tag.Lookup(sql.CondiTagName)
		if tag == "-" || fieldKind == reflect.Interface {
			continue
		}
		if subConditions, ok := conditionsByImpl(field); ok {
			conditions = append(conditions, subConditions)
			continue
		}

		if !ok && structField.Anonymous && (fieldKind == reflect.Ptr || fieldKind == reflect.Struct) {
			subConditions := conditionsBy(field)
			if subConditions != nil {
				conditions = append(conditions, subConditions...)
			}
		} else {
			if tag == "" && empty {
				continue
			}
			if tag == "embedded" && (fieldKind == reflect.Ptr || fieldKind == reflect.Struct) {
				subConditions := conditionsBy(field)
				if subConditions != nil {
					conditions = append(conditions, subConditions...)
				}
				continue
			}

			if tag == "" {
				if fieldKind == reflect.Ptr || fieldKind == reflect.Struct || fieldKind == reflect.Map {
					subCondition := andConditionBy(field)
					if subCondition != nil {
						conditions = append(conditions, subCondition)
					}
					continue
				}
				conditions = append(conditions, clause.Eq{Column: stringsx.CamelToSnake(structField.Name), Value: vi.Field(i).Interface()})
				continue
			}
			if !strings.Contains(tag, ";") && !strings.Contains(tag, ":") {
				switch tag {
				case "or":
					if fieldKind == reflect.Ptr || fieldKind == reflect.Struct || fieldKind == reflect.Map {
						subCondition := orConditionBy(field)
						if subCondition != nil {
							conditions = append(conditions, subCondition)
						}
					}
					continue
				case "omitempty":
					if empty {
						continue
					}
				case "validempty":
					conditions = append(conditions, clause.Eq{Column: stringsx.CamelToSnake(structField.Name), Value: vi.Field(i).Interface()})
					continue
				default:
					if empty {
						continue
					}
					op := sql.ParseConditionOperation(tag)
					if op == sql.OperationPlace {
						conditions = append(conditions, clause.Expr{SQL: tag, Vars: []any{vi.Field(i).Interface()}})
					} else {
						conditions = append(conditions, NewCondition(stringsx.CamelToSnake(structField.Name), op, vi.Field(i).Interface()))
					}

					continue
				}

			}
			conditionTag, err := structtag.ParseSettingTagToStruct[sql.ConditionTag](tag, ';')
			if err != nil {
				panic(err)
			}
			if !conditionTag.EmptyValid && empty {
				continue
			}
			column := conditionTag.Column
			if column == "" {
				column = stringsx.CamelToSnake(structField.Name)
			}
			if conditionTag.Operate == "" {
				conditions = append(conditions, clause.Eq{Column: column, Value: vi.Field(i).Interface()})
				continue
			}
			op := sql.ParseConditionOperation(conditionTag.Operate)
			if op == sql.OperationPlace {
				conditions = append(conditions, clause.Expr{SQL: conditionTag.Operate, Vars: []any{vi.Field(i).Interface()}})
			} else {
				conditions = append(conditions, NewCondition(stringsx.CamelToSnake(structField.Name), op, vi.Field(i).Interface()))
			}
		}
	}
	return conditions
}
