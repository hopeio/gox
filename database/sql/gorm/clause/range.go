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
	if req.Type.ContainsBegin() && req.Type.ContainsBegin() {
		return NewCondition(req.Field, dbi.Between, req.Begin, req.End)
	}
	leftOp, rightOp := dbi.Greater, dbi.Less
	if req.Type.ContainsBegin() {
		leftOp = dbi.GreaterOrEqual
	}
	if req.Type.ContainsEnd() {
		rightOp = dbi.LessOrEqual
	}
	return clause.AndConditions{Exprs: []clause.Expression{NewCondition(req.Field, leftOp, req.Begin), NewCondition(req.Field, rightOp, req.End)}}
}
