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

	"github.com/hopeio/gox/types/request"
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

type Sorts []request.Sort

func (o Sorts) Clauses() []clause.Expression {
	if len(o) == 0 {
		return nil
	}
	return []clause.Expression{
		SortExpr(o...),
	}
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

type Pagination request.Pagination

func (req *Pagination) Clauses() []clause.Expression {
	if req.No == 0 && req.Size == 0 {
		return nil
	}
	if len(req.Sort) > 0 {
		return []clause.Expression{SortExpr(req.Sort...), PaginationExpr(req.No, req.Size)}
	}

	return []clause.Expression{PaginationExpr(req.No, req.Size)}
}

func (req *Pagination) Apply(db *gorm.DB) *gorm.DB {
	return db.Clauses(req.Clauses()...)
}

func PaginationExpr(pageNo, pageSize uint32) clause.Expression {
	if pageNo == 0 || pageSize == 0 {
		return nil
	}
	if pageNo > 1 {
		return Limit{Offset: uint64(pageNo-1) * uint64(pageSize), Limit: pageSize}
	}
	return Limit{Limit: pageSize}
}

func FindPagination[T any](db *gorm.DB, req *request.Pagination, clauses ...clause.Expression) ([]T, int64, error) {
	var models []T

	if len(clauses) > 0 {
		db = db.Clauses(clauses...)
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
	pageClauses := (*Pagination)(req).Clauses()
	err = db.Clauses(pageClauses...).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	return models, count, nil
}

type PaginationEmbedded request.PaginationEmbedded

func (req *PaginationEmbedded) ToPagination() *Pagination {
	return (*Pagination)(unsafe.Pointer(req))
}
