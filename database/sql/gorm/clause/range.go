/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	dbi "github.com/hopeio/gox/database/sql"
	"github.com/hopeio/gox/types/param"
	"gorm.io/gorm/clause"
)

type Range[T param.Ordered] param.Range[T]

func (req *Range[T]) Condition() clause.Expression {
	if req == nil || req.Field == "" {
		return nil
	}
	if req.Type == 0 {
		return NewCondition(req.Field, dbi.Between, req.Begin, req.End)
	}
	if req.Type.HasBegin() && req.Type.HasEnd() {
		if req.Type.ContainsBegin() && req.Type.ContainsEnd() {
			return NewCondition(req.Field, dbi.Between, req.Begin, req.End)
		} else {
			leftOp, rightOp := dbi.Greater, dbi.Less
			if req.Type.ContainsBegin() {
				leftOp = dbi.GreaterOrEqual
			}
			if req.Type.ContainsEnd() {
				leftOp = dbi.LessOrEqual
			}
			return clause.AndConditions{Exprs: []clause.Expression{NewCondition(req.Field, leftOp, req.Begin), NewCondition(req.Field, rightOp, req.End)}}
		}
	}

	if req.Type.HasBegin() {
		operation := dbi.Greater
		if req.Type.ContainsEnd() {
			operation = dbi.GreaterOrEqual
		}
		return NewCondition(req.Field, operation, req.Begin)
	}
	if req.Type.HasEnd() {
		operation := dbi.Less
		if req.Type.ContainsEnd() {
			operation = dbi.LessOrEqual
		}
		return NewCondition(req.Field, operation, req.End)
	}
	return nil
}

type RangeInTwoField[T param.Ordered] param.RangeInTwoField[T]

func (req *RangeInTwoField[T]) Condition() clause.Expression {
	if req == nil || req.BeginField == "" || req.EndField == "" {
		return nil
	}
	if req.Type == 0 {
		return clause.AndConditions{Exprs: []clause.Expression{clause.Or(Between{Column: req.BeginField, Begin: req.Begin, End: req.End}, Between{Column: req.EndField, Begin: req.Begin, End: req.End})}}
	}
	if req.Type.HasBegin() && req.Type.HasEnd() {
		if req.Type.ContainsBegin() && req.Type.ContainsEnd() {
			return clause.AndConditions{Exprs: []clause.Expression{clause.Or(Between{Column: req.BeginField, Begin: req.Begin, End: req.End}, Between{Column: req.EndField, Begin: req.Begin, End: req.End})}}
		} else {
			if req.Type.ContainsBegin() {
				return clause.AndConditions{Exprs: []clause.Expression{clause.Or(Between{Column: req.BeginField, Begin: req.Begin, End: req.End}, clause.And(clause.Gte{Column: req.EndField, Value: req.Begin}, clause.Lt{Column: req.EndField, Value: req.End}))}}
			}
			if req.Type.ContainsEnd() {
				return clause.AndConditions{Exprs: []clause.Expression{clause.And(clause.Gt{Column: req.BeginField, Value: req.Begin}, clause.Lte{Column: req.EndField, Value: req.End}), Between{Column: req.EndField, Begin: req.Begin, End: req.End}}}
			}

		}
	}

	if req.Type.HasBegin() {
		operation := dbi.Greater
		if req.Type.ContainsEnd() {
			operation = dbi.GreaterOrEqual
		}
		return NewCondition(req.BeginField, operation, req.Begin)
	}
	if req.Type.HasEnd() {
		operation := dbi.Less
		if req.Type.ContainsEnd() {
			operation = dbi.LessOrEqual
		}
		return NewCondition(req.EndField, operation, req.End)
	}
	return nil
}
