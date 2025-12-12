//go:build go1.18

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	"reflect"

	sqlx "github.com/hopeio/gox/database/sql"
	"github.com/hopeio/gox/types/request"
	"gorm.io/gorm/clause"
)

type ConditionExpr interface {
	Condition() clause.Expression
}

var ConditionExprType = reflect.TypeOf((*ConditionExpr)(nil)).Elem()

func NewCondition(field string, op sqlx.ConditionOperation, args ...any) clause.Expression {
	if field == "" {
		return nil
	}
	switch op {
	case sqlx.Equal:
		if len(args) == 0 {
			return nil
		}
		return clause.Eq{
			Column: field,
			Value:  args[0],
		}
	case sqlx.In:
		return clause.IN{
			Column: field,
			Values: args,
		}
	case sqlx.Between:
		if len(args) != 2 {
			return nil
		}
		return Between{
			Column: field,
			Begin:  args[0],
			End:    args[1],
		}
	case sqlx.Greater:
		if len(args) == 0 {
			return nil
		}
		return clause.Gt{
			Column: field,
			Value:  args[0],
		}
	case sqlx.Less:
		if len(args) == 0 {
			return nil
		}
		return clause.Lt{
			Column: field,
			Value:  args[0],
		}
	case sqlx.Like:
		if len(args) == 0 {
			return nil
		}
		return clause.Like{
			Column: field,
			Value:  args[0],
		}
	case sqlx.GreaterOrEqual:
		if len(args) == 0 {
			return nil
		}
		return clause.Gte{
			Column: field,
			Value:  args[0],
		}
	case sqlx.LessOrEqual:
		if len(args) == 0 {
			return nil
		}
		return clause.Lte{
			Column: field,
			Value:  args[0],
		}
	case sqlx.NotIn:
		return Not{Expr: clause.IN{
			Column: field,
			Values: args,
		}}
	case sqlx.NotEqual:
		if len(args) == 0 {
			return nil
		}
		return clause.Neq{
			Column: field,
			Value:  args[0],
		}
	case sqlx.IsNull:
		return IsNull{
			Column: field,
		}
	case sqlx.IsNotNull:
		return IsNotNull{
			Column: field,
		}
	}
	return clause.Expr{
		SQL:  field,
		Vars: args,
	}
}

type Sorts []request.Sort

func (o Sorts) Condition() clause.Expression {
	if len(o) == 0 {
		return nil
	}
	return SortExpr(o...)
}

func SortExpr(sorts ...request.Sort) clause.Expression {
	if len(sorts) == 0 {
		return nil
	}
	var orders []clause.OrderByColumn
	for _, sort := range sorts {
		orders = append(orders, clause.OrderByColumn{
			Column: clause.Column{
				Name: sort.Field,
				Raw:  true,
			},
			Desc: sort.Type == request.SortTypeDesc,
		})
	}
	return clause.OrderBy{Columns: orders}
}

func ByPrimaryKey(v any) clause.Expression {
	return clause.Eq{
		Column: clause.PrimaryColumn,
		Value:  v,
	}
}

type Between struct {
	Column     any
	Begin, End any
}

func (gt Between) Build(builder clause.Builder) {
	builder.WriteQuoted(gt.Column)
	builder.WriteString(" BETWEEN ")
	builder.AddVar(builder, gt.Begin)
	builder.WriteString(" AND ")
	builder.AddVar(builder, gt.End)
}

func (gt Between) NegationBuild(builder clause.Builder) {
	builder.WriteQuoted(gt.Column)
	builder.WriteString(" < ")
	builder.AddVar(builder, gt.Begin)
	builder.WriteString(" OR ")
	builder.WriteQuoted(gt.Column)
	builder.WriteString(" > ")
	builder.AddVar(builder, gt.End)
}

type Not struct {
	Expr clause.NegationExpressionBuilder
}

func (n Not) Build(builder clause.Builder) {
	n.Expr.NegationBuild(builder)
}

type IsNull struct {
	Column any
}

func (in IsNull) Build(builder clause.Builder) {
	builder.WriteQuoted(in.Column)
	builder.WriteString(" IS NULL")
}

func (in IsNull) NegationBuild(builder clause.Builder) {
	IsNotNull(in).Build(builder)
}

type IsNotNull IsNull

func (inn IsNotNull) Build(builder clause.Builder) {
	builder.WriteQuoted(inn.Column)
	builder.WriteString(" IS NOT NULL")
}

func (inn IsNotNull) NegationBuild(builder clause.Builder) {
	IsNull(inn).Build(builder)
}
