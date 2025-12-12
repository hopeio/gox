//go:build go1.18

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	"strconv"

	"github.com/hopeio/gox/types/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Limit limit clause
type Limit struct {
	Limit  int
	Offset int
}

// Name where clause name
func (limit Limit) Name() string {
	return "LIMIT"
}

// Build build where clause
func (limit Limit) Build(builder clause.Builder) {
	if limit.Limit > 0 {
		builder.WriteString("LIMIT ")
		builder.WriteString(strconv.Itoa(limit.Limit))
	}
	if limit.Offset > 0 {
		if limit.Limit > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString("OFFSET ")
		builder.WriteString(strconv.Itoa(limit.Offset))
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

type PaginationEmbedded request.PaginationEmbedded

func (req *PaginationEmbedded) Clause() []clause.Expression {
	if req.PageNo == 0 && req.PageSize == 0 {
		return nil
	}
	if len(req.Sort) > 0 {
		return []clause.Expression{PaginationExpr(req.PageNo, req.PageSize)}
	}

	return []clause.Expression{SortExpr(req.Sort...), PaginationExpr(req.PageNo, req.PageSize)}
}

func FindPaginationEmbedded[T any](db *gorm.DB, req *request.PaginationEmbedded, clauses ...clause.Expression) ([]T, int64, error) {
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
	pageClauses := (*PaginationEmbedded)(req).Clause()
	err = db.Clauses(pageClauses...).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	return models, count, nil
}

type Pagination request.Pagination

func (req *Pagination) Clause() []clause.Expression {
	if req.No == 0 && req.Size == 0 {
		return nil
	}
	if len(req.Sort) == 0 {
		return []clause.Expression{PaginationExpr(req.No, req.Size)}
	}

	return []clause.Expression{SortExpr(req.Sort...), PaginationExpr(req.No, req.Size)}
}

func (req *Pagination) Apply(db *gorm.DB) *gorm.DB {
	return db.Clauses(req.Clause()...)
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
	pageClauses := (*Pagination)(req).Clause()
	err = db.Clauses(pageClauses...).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	return models, count, nil
}

func PaginationExpr(pageNo, pageSize int, sort ...request.Sort) clause.Limit {
	if pageSize == 0 {
		pageSize = 100
	}
	if pageNo > 1 {
		return clause.Limit{Offset: (pageNo - 1) * pageSize, Limit: &pageSize}
	}
	return clause.Limit{Limit: &pageSize}
}

func (req *PaginationEmbedded) Apply(db *gorm.DB) *gorm.DB {
	return db.Clauses(req.Clause()...)
}
