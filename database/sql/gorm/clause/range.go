/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	sqlx "github.com/hopeio/gox/database/sql"
	"github.com/hopeio/gox/types/request"
	"gorm.io/gorm/clause"
)

type Range[T request.Ordered] request.Range[T]

func (req *Range[T]) Condition() clause.Expression {
	if req == nil || req.Field == "" {
		return nil
	}
	if req.Type == 0 {
		return NewCondition(req.Field, sqlx.Between, req.Begin, req.End)
	}
	if req.Type.ContainsBegin() && req.Type.ContainsBegin() {
		return NewCondition(req.Field, sqlx.Between, req.Begin, req.End)
	}
	leftOp, rightOp := sqlx.Greater, sqlx.Less
	if req.Type.ContainsBegin() {
		leftOp = sqlx.GreaterOrEqual
	}
	if req.Type.ContainsEnd() {
		rightOp = sqlx.LessOrEqual
	}
	return clause.AndConditions{Exprs: []clause.Expression{NewCondition(req.Field, leftOp, req.Begin), NewCondition(req.Field, rightOp, req.End)}}
}
