/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package sql

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	strconv2 "github.com/hopeio/gox/encoding/text"
)

type ConditionOperation int

const (
	OperationPlace ConditionOperation = iota
	Equal
	NotEqual
	Greater
	Less
	Between
	GreaterOrEqual
	LessOrEqual
	IsNotNull
	IsNull
	In
	NotIn
	Like
	NotLike
)

func ParseConditionOperation(op string) ConditionOperation {
	op = strings.ToUpper(op)
	switch op {
	case "=", " = ", "EQUAL", "1":
		return Equal
	case ">", " > ", "GREATER", "3":
		return Greater
	case "<", " < ", "LESS", "4":
		return Less
	case ">=", " >= ", "GREATER OR EQUAL", "6":
		return GreaterOrEqual
	case "<=", " <= ", "LESS OR EQUAL", "LESSOREQUAL", "7":
		return LessOrEqual
	case "!=", " != ", "NOT EQUAL", "NOTEQUAL", "2":
		return NotEqual
	case " IN ", "10":
		return In
	case "NOT IN", "NOTIN", "11":
		return NotIn
	case "LIKE", "12":
		return Like
	case "NOT LIKE", "NOTLIKE", "13":
		return NotLike
	case "NULL", "IS NULL", "ISNULL", "9":
		return IsNull
	case "NOT NULL", "NOTNULL", "IS NOT NULL", "ISNOTNULL", "8":
		return IsNotNull
	}
	return OperationPlace
}

func (m ConditionOperation) SQL() string {
	switch m {
	case Equal:
		return "= ?"
	case NotEqual:
		return "!= ?"
	case Greater:
		return "> ?"
	case Less:
		return "< ?"
	case Between:
		return "BETWEEN ? AND ?"
	case GreaterOrEqual:
		return ">= ?"
	case LessOrEqual:
		return "<= ?"
	case IsNull:
		return "IS NULL"
	case IsNotNull:
		return "IS NOT NULL"
	case In:
		return "IN (?)"
	case NotIn:
		return "NOT IN (?)"
	case Like:
		return "LIKE ?"
	case NotLike:
		return "NOT LIKE ?"
	default:
		return ""
	}
}

func (m ConditionOperation) String() string {
	switch m {
	case Equal:
		return " = "
	case In:
		return " IN "
	case Between:
		return " BETWEEN "
	case Greater:
		return " > "
	case Less:
		return " < "
	case NotEqual:
		return " != "
	case GreaterOrEqual:
		return " >= "
	case LessOrEqual:
		return " <= "
	case IsNull:
		return " IS NULL"
	case IsNotNull:
		return " IS NOT NULL"
	case NotIn:
		return " NOT IN "
	case Like:
		return " LIKE "
	case NotLike:
		return " NOT LIKE "
	default:
		return "="
	}
}

type FilterExpr struct {
	Field     string             `json:"field"`
	Operation ConditionOperation `json:"op"`
	Value     []any              `json:"value"`
}

func (filter *FilterExpr) Build() string {
	filter.Field = strings.TrimSpace(filter.Field)

	if filter.Field == "" {
		return ""
	}
	switch filter.Operation {
	case Greater, Less, Equal, NotEqual, GreaterOrEqual, LessOrEqual:
		return filter.Field + filter.Operation.String() + ConvertParams(filter.Value[0], "'")
	case In, NotIn:
		var vars = make([]string, len(filter.Value))
		for idx, v := range filter.Value {
			vars[idx] = ConvertParams(v, "'")
		}
		return filter.Field + filter.Operation.String() + "(" + strings.Join(vars, ",") + ")"
	case Between:
		if len(filter.Value) < 2 {
			return ""
		}
		var vars = make([]string, len(filter.Value))
		for idx, v := range filter.Value {
			vars[idx] = ConvertParams(v, "'")
		}
		return filter.Field + filter.Operation.String() + vars[0] + " AND " + vars[1]
	case Like, NotLike:
		return filter.Field + filter.Operation.String() + ConvertParams(filter.Value[0], "'")
	case IsNull, IsNotNull:
		return filter.Field + filter.Operation.String()
	}
	return filter.Field
}

type FilterExprs []FilterExpr

func (f FilterExprs) Build() string {
	var conditions []string
	for _, filter := range f {
		filter.Field = strings.TrimSpace(filter.Field)
		condition := filter.Build()
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return ""
	}
	return strings.Join(conditions, " AND ")
}

func ConvertParams(v interface{}, escaper string) string {
	switch v := v.(type) {
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return escaper + v.Format(TmFmtWithMS) + escaper
	case *time.Time:
		if v != nil {
			return escaper + v.Format(TmFmtWithMS) + escaper
		}

		return NullStr
	case driver.Valuer:
		reflectValue := reflect.ValueOf(v)
		if v != nil && reflectValue.IsValid() && ((reflectValue.Kind() == reflect.Ptr && !reflectValue.IsNil()) || reflectValue.Kind() != reflect.Ptr) {
			r, _ := v.Value()
			ConvertParams(r, escaper)
		} else {
			return NullStr
		}
	case fmt.Stringer:
		reflectValue := reflect.ValueOf(v)
		if v != nil && reflectValue.IsValid() && ((reflectValue.Kind() == reflect.Ptr && !reflectValue.IsNil()) || reflectValue.Kind() != reflect.Ptr) {
			return escaper + strings.ReplaceAll(v.String(), escaper, "\\"+escaper) + escaper
		}

		return NullStr
	case []byte:
		if isPrintable(v) {
			return escaper + strings.ReplaceAll(string(v), escaper, "\\"+escaper) + escaper
		} else {
			return escaper + "<binary>" + escaper
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return strconv2.FormatInteger(v)
	case float64, float32:
		return fmt.Sprintf("%.6f", v)
	case string:
		return escaper + strings.ReplaceAll(v, escaper, "\\"+escaper) + escaper
	default:
		rv := reflect.ValueOf(v)
		if v == nil || !rv.IsValid() || rv.Kind() == reflect.Ptr && rv.IsNil() {
			return NullStr
		} else if valuer, ok := v.(driver.Valuer); ok {
			v, _ = valuer.Value()
			ConvertParams(v, escaper)
		} else if rv.Kind() == reflect.Ptr && !rv.IsZero() {
			ConvertParams(reflect.Indirect(rv).Interface(), escaper)
		} else {
			for _, t := range convertableTypes {
				if rv.Type().ConvertibleTo(t) {
					return ConvertParams(rv.Convert(t).Interface(), escaper)
				}
			}
			return escaper + strings.ReplaceAll(fmt.Sprint(v), escaper, "\\"+escaper) + escaper
		}
	}
	return ""
}

var convertableTypes = []reflect.Type{reflect.TypeOf(time.Time{}), reflect.TypeOf(false), reflect.TypeOf([]byte{})}

func isPrintable(s []byte) bool {
	for _, r := range s {
		if !unicode.IsPrint(rune(r)) {
			return false
		}
	}
	return true
}

func (f FilterExprs) BuildSQL() (string, []interface{}) {
	var builder strings.Builder
	var vars []interface{}
	for i, filter := range f {
		if filter.Field == "" || filter.Operation == 0 || len(filter.Value) == 0 {
			continue
		}
		builder.WriteString(filter.Field)
		builder.WriteByte(' ')
		builder.WriteString(filter.Operation.SQL())
		if i < len(f) {
			builder.WriteString(" AND")
		}
		vars = append(vars, filter.Value...)
	}
	return "", nil
}
