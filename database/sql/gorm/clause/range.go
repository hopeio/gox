/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	sqlx "github.com/hopeio/gox/database/sql"
	"gorm.io/gorm/clause"
)

type Range[T any] sqlx.Range[T]

func (req *Range[T]) Condition() clause.Expression {
	if req == nil || req.Field == "" {
		return nil
	}
	if req.Type == 0 {
		return NewCondition(req.Field, sqlx.Between, []any{req.Begin, req.End})
	}
	if req.Type.ContainsBegin() && req.Type.ContainsEnd() {
		return NewCondition(req.Field, sqlx.Between, []any{req.Begin, req.End})
	}
	if req.Type.HasBegin() && req.Type.HasEnd() {
		if req.Type.ContainsBegin() && req.Type.ContainsEnd() {
			return NewCondition(req.Field, sqlx.Between, []any{req.Begin, req.End})
		}
		leftOp, rightOp := sqlx.Greater, sqlx.Less
		if req.Type.ContainsBegin() {
			leftOp = sqlx.GreaterOrEqual
		}
		if req.Type.ContainsEnd() {
			leftOp = sqlx.LessOrEqual
		}
		return clause.AndConditions{Exprs: []clause.Expression{NewCondition(req.Field, leftOp, req.Begin), NewCondition(req.Field, rightOp, req.End)}}
	}
	if req.Type.HasBegin() {
		operation := sqlx.Greater
		if req.Type.ContainsEnd() {
			operation = sqlx.GreaterOrEqual
		}
		return NewCondition(req.Field, operation, req.Begin)
	}
	if req.Type.HasEnd() {
		operation := sqlx.Less
		if req.Type.ContainsEnd() {
			operation = sqlx.LessOrEqual
		}
		return NewCondition(req.Field, operation, req.End)
	}
	return nil
}
