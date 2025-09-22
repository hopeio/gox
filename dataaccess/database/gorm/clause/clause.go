//go:build go1.18

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	dbi "github.com/hopeio/gox/dataaccess/database/sql"
	"github.com/hopeio/gox/types/param"
	"gorm.io/gorm/clause"
	"reflect"
)

type ConditionExpr interface {
	Condition() clause.Expression
}

var ConditionExprType = reflect.TypeOf((*ConditionExpr)(nil)).Elem()

func NewCondition(field string, op dbi.ConditionOperation, args ...any) clause.Expression {
	if field == "" {
		return nil
	}
	switch op {
	case dbi.Equal:
		if len(args) == 0 {
			return nil
		}
		return clause.Eq{
			Column: field,
			Value:  args[0],
		}
	case dbi.In:
		return clause.IN{
			Column: field,
			Values: args,
		}
	case dbi.Between:
		if len(args) != 2 {
			return nil
		}
		return Between{
			Column: field,
			Begin:  args[0],
			End:    args[1],
		}
	case dbi.Greater:
		if len(args) == 0 {
			return nil
		}
		return clause.Gt{
			Column: field,
			Value:  args[0],
		}
	case dbi.Less:
		if len(args) == 0 {
			return nil
		}
		return clause.Lt{
			Column: field,
			Value:  args[0],
		}
	case dbi.Like:
		if len(args) == 0 {
			return nil
		}
		return clause.Like{
			Column: field,
			Value:  args[0],
		}
	case dbi.GreaterOrEqual:
		if len(args) == 0 {
			return nil
		}
		return clause.Gte{
			Column: field,
			Value:  args[0],
		}
	case dbi.LessOrEqual:
		if len(args) == 0 {
			return nil
		}
		return clause.Lte{
			Column: field,
			Value:  args[0],
		}
	case dbi.NotIn:
		return Not{Expr: clause.IN{
			Column: field,
			Values: args,
		}}
	case dbi.NotEqual:
		if len(args) == 0 {
			return nil
		}
		return clause.Neq{
			Column: field,
			Value:  args[0],
		}
	case dbi.IsNull:
		return IsNull{
			Column: field,
		}
	case dbi.IsNotNull:
		return IsNotNull{
			Column: field,
		}
	}
	return clause.Expr{
		SQL:  field,
		Vars: args,
	}
}

func SortExpr(column string, typ param.SortType) clause.Expression {
	var desc bool
	if typ == param.SortTypeDesc {
		desc = true
	}
	return clause.OrderBy{Columns: []clause.OrderByColumn{{Column: clause.Column{Name: column, Raw: true}, Desc: desc}}}
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
