//go:build go1.18

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package clause

import (
	"github.com/hopeio/gox/types/param"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
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

type PageSortEmbed param.PageSortEmbed

func (req *PageSortEmbed) Clause() []clause.Expression {
	if req.PageNo == 0 && req.PageSize == 0 {
		return nil
	}
	if req.SortEmbed == nil || req.SortEmbed.SortField == "" {
		return []clause.Expression{PageExpr(req.PageNo, req.PageSize)}
	}

	return []clause.Expression{SortExpr(req.SortField, req.SortType), PageExpr(req.PageNo, req.PageSize)}
}

func FindByPageSortEmbed[T any](db *gorm.DB, req *param.PageSortEmbed, clauses ...clause.Expression) ([]T, int64, error) {
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
	pageSortClauses := (*PageSortEmbed)(req).Clause()
	err = db.Clauses(pageSortClauses...).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	return models, count, nil
}

type PageSort param.PageSort

func (req *PageSort) Clause() []clause.Expression {
	if req.Page.No == 0 && req.Page.Size == 0 {
		return nil
	}
	if req.Sort == nil || req.Sort.Field == "" {
		return []clause.Expression{PageExpr(req.Page.No, req.Page.Size)}
	}

	return []clause.Expression{SortExpr(req.Sort.Field, req.Sort.Type), PageExpr(req.Page.No, req.Page.Size)}
}

func (req *PageSort) Apply(db *gorm.DB) *gorm.DB {
	return db.Clauses(req.Clause()...)
}

func FindByPageSort[T any](db *gorm.DB, req *param.PageSort, clauses ...clause.Expression) ([]T, int64, error) {
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
	pageSortClauses := (*PageSort)(req).Clause()
	err = db.Clauses(pageSortClauses...).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	return models, count, nil
}

func PageExpr(pageNo, pageSize int) clause.Limit {
	if pageSize == 0 {
		pageSize = 100
	}
	if pageNo > 1 {
		return clause.Limit{Offset: (pageNo - 1) * pageSize, Limit: &pageSize}
	}
	return clause.Limit{Limit: &pageSize}
}

type PageEmbed param.PageEmbed

func (req *PageEmbed) Clause() clause.Expression {
	if req.PageNo == 0 && req.PageSize == 0 {
		return nil
	}
	return PageExpr(req.PageNo, req.PageSize)
}

func (req *PageEmbed) Apply(db *gorm.DB) *gorm.DB {
	return db.Clauses(req.Clause())
}
