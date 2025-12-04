package clause

import (
	"strings"

	dbi "github.com/hopeio/gox/database/sql"
	"gorm.io/gorm/clause"
)

type FilterExpr dbi.FilterExpr

func (f *FilterExpr) Condition() clause.Expression {
	f.Field = strings.TrimSpace(f.Field)

	return NewCondition(f.Field, f.Operation, f.Value)
}

type FilterExprs dbi.FilterExprs

func (f FilterExprs) Condition() clause.Expression {
	var exprs []clause.Expression
	for _, filter := range f {
		filter.Field = strings.TrimSpace(filter.Field)

		filterExpr := (FilterExpr)(filter)
		expr := filterExpr.Condition()
		if expr != nil {
			exprs = append(exprs, expr)
		}
	}
	if len(exprs) > 0 {
		return clause.AndConditions{Exprs: exprs}
	}
	return nil
}
