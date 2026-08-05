package gorm

import (
	"strings"

	sqlx "github.com/hopeio/gox/database/sql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FilterExpr sqlx.FilterExpr

// Condition ...
func (f *FilterExpr) Condition() clause.Expression {
	f.Field = strings.TrimSpace(f.Field)
	return NewCondition(f.Field, f.Operation, f.Value)
}

type FilterExprs sqlx.FilterExprs

// Conditions ...
func (f FilterExprs) Conditions() []clause.Expression {
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
		return exprs
	}
	return nil
}

// Condition ...
func (f FilterExprs) Condition() clause.Expression {
	exprs := f.Conditions()
	if len(exprs) > 0 {
		return clause.AndConditions{Exprs: exprs}
	}
	return nil
}

// Apply ...
func (f FilterExprs) Apply(db *gorm.DB) *gorm.DB {
	for _, filter := range f {
		filter.Field = strings.TrimSpace(filter.Field)

		if filter.Field == "" {
			continue
		}

		db = db.Where(filter.Field+" "+filter.Operation.SQL(), sqlx.AnyToAnys(filter.Value))
	}
	return db
}
