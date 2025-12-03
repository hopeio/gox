package clause

import (
	"reflect"

	"github.com/hopeio/gox/database/sql"
	"github.com/hopeio/gox/reflect/structtag"
	stringsx "github.com/hopeio/gox/strings"
	"gorm.io/gorm/clause"
)

/*
	type ReportList struct {
		param.PageSortEmbed `sqlcondi:"-"`
		LoadingTime         *clause.Range[time.Time]
		UserId              int
		CarId               int
		TaskId              int
		RouteId             int
		Diff                float64                  `sqlcondi:"-"`
		Outlier             bool                     `sqlcondi:"-"`
	}
*/
func ConditionByStruct(param any) (clause.Expression, error) {
	v := reflect.ValueOf(param)
	v = reflect.Indirect(v)
	var conds []clause.Expression
	t := v.Type()
	for i := range v.NumField() {
		field := v.Field(i)
		fieldKind := field.Kind()
		structField := t.Field(i)
		empty := field.IsZero()
		tag, ok := structField.Tag.Lookup(sql.CondiTagName)
		if tag == "-" {
			continue
		}
		if !ok && structField.Anonymous && (fieldKind == reflect.Interface || fieldKind == reflect.Ptr || fieldKind == reflect.Struct) {
			subCondition, err := ConditionByStruct(field.Interface())
			if err != nil {
				return nil, err
			}
			if subCondition != nil {
				conds = append(conds, subCondition)
			}
		} else {
			if tag == "" && empty {
				continue
			}
			if structField.Type.Implements(ConditionExprType) {
				if (fieldKind == reflect.Interface || fieldKind == reflect.Ptr) && field.Elem().IsZero() {
					continue
				}
				if cond := field.Interface().(ConditionExpr).Condition(); cond != nil {
					conds = append(conds, cond)
				}
				continue
			}
			if fieldKind == reflect.Struct && field.Addr().Type().Implements(ConditionExprType) {
				conds = append(conds, field.Addr().Interface().(ConditionExpr).Condition())
				continue
			}

			if tag == "" {
				if fieldKind == reflect.Interface || fieldKind == reflect.Ptr || fieldKind == reflect.Struct {
					subCondition, err := ConditionByStruct(field.Interface())
					if err != nil {
						return nil, err
					}
					if subCondition != nil {
						conds = append(conds, subCondition)
					}
					continue
				}
				if fieldKind == reflect.Map {
					if field.Type().Key().Kind() == reflect.String {
						iter := field.MapRange()
						for iter.Next() {
							conds = append(conds, clause.Eq{Column: stringsx.CamelToSnake(iter.Key().String()), Value: iter.Value()})
						}
					}
					continue
				}
				conds = append(conds, clause.Eq{Column: stringsx.CamelToSnake(structField.Name), Value: v.Field(i).Interface()})
				continue
			}
			condition, err := structtag.ParseSettingTagToStruct[sql.ConditionTag](tag, ';')
			if err != nil {
				return nil, err
			}
			if !condition.EmptyValid && empty {
				continue
			}
			if condition.Expr != "" {
				conds = append(conds, clause.Expr{SQL: condition.Expr, Vars: []any{v.Field(i).Interface()}})
			} else {
				column := condition.Column
				if column == "" {
					column = stringsx.CamelToSnake(structField.Name)
				}
				if condition.Op == "" {
					condition.Op = "EQUAL"
				}
				op := sql.ParseConditionOperation(condition.Op)
				conds = append(conds, NewCondition(column, op, v.Field(i).Interface()))
			}
		}
	}
	if len(conds) == 0 {
		return nil, nil
	}
	return clause.AndConditions{Exprs: conds}, nil
}
