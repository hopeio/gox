//go:build go1.18

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	"strconv"
	"unsafe"

	sqlx "github.com/hopeio/gox/database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Limit limit clause
type Limit struct {
	Limit  uint32
	Offset uint64
}

// Name where clause name
func (limit Limit) Name() string {
	return "LIMIT"
}

// Build build where clause
func (limit Limit) Build(builder clause.Builder) {
	if limit.Limit > 0 {
		builder.WriteString("LIMIT ")
		builder.WriteString(strconv.Itoa(int(limit.Limit)))
	}
	if limit.Offset > 0 {
		if limit.Limit > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString("OFFSET ")
		builder.WriteString(strconv.Itoa(int(limit.Offset)))
	}
}

// MergeClause merge order by clauses
func (limit Limit) MergeClause(clause *clause.Clause) {
	clause.Name = ""

	if v, ok := clause.Expression.(Limit); ok {
		if limit.Limit == 0 && v.Limit != 0 {
			limit.Limit = v.Limit
		}

		if limit.Offset == 0 && v.Offset > 0 {
			limit.Offset = v.Offset
		} else if limit.Offset < 0 {
			limit.Offset = 0
		}
	}

	clause.Expression = limit
}

type Sort sqlx.Sort

func (o *Sort) Clause() clause.Expression {
	if o.Field == "" {
		return nil
	}
	return SingleSortExpr(o.Field, o.Type)
}
func SingleSortExpr(field string, sortType sqlx.SortType) clause.Expression {
	return clause.OrderBy{Columns: []clause.OrderByColumn{{Column: clause.Column{Name: field}, Desc: sortType == sqlx.SortTypeDesc}}}
}

type Sorts []sqlx.Sort

func (o Sorts) Clause() clause.Expression {
	if len(o) == 0 {
		return nil
	}
	return SortExpr(nil, o...)
}

func SortExpr(expr clause.Expression, sorts ...sqlx.Sort) clause.Expression {
	if expr == nil && len(sorts) == 0 {
		return nil
	}
	var orders []clause.OrderByColumn
	for _, sort := range sorts {
		orders = append(orders, clause.OrderByColumn{
			Column: clause.Column{
				Name: sort.Field,
				Raw:  true,
			},
			Desc: sort.Type == sqlx.SortTypeDesc,
		})
	}
	return clause.OrderBy{Columns: orders, Expression: expr}
}

type Pagination sqlx.Pagination

func (req *Pagination) Clause() clause.Expression {
	if req.No == 0 && req.Size == 0 {
		return nil
	}
	return PaginationExpr(req.No, req.Size)
}

func (req *Pagination) Apply(db *gorm.DB) *gorm.DB {
	return db.Clauses(req.Clause())
}

func PaginationExpr(pageNo, pageSize uint32) clause.Expression {
	if pageNo == 0 || pageSize == 0 {
		return nil
	}
	limit := Limit{Limit: pageSize}
	if pageNo > 1 {
		limit.Offset = uint64(pageNo-1) * uint64(pageSize)
	}

	return limit
}

func FindPagination[T any](db *gorm.DB, page *Pagination, sort Sorts, conds ...clause.Expression) ([]T, int64, error) {
	var models []T

	if len(conds) > 0 {
		db = db.Clauses(conds...)
	}
	var count int64
	var t T
	err := db.Model(&t).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return nil, 0, nil
	}
	pageClauses := page.Clause()
	sortClauses := sort.Clause()
	err = db.Clauses(pageClauses, sortClauses).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	return models, count, nil
}

type PaginationEmbedded sqlx.PaginationEmbedded

func (req *PaginationEmbedded) ToPagination() *Pagination {
	return (*Pagination)(unsafe.Pointer(req))
}
