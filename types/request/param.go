/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package request

import (
	sqlx "github.com/hopeio/gox/database/sql"
	gormx "github.com/hopeio/gox/database/sql/gorm"
)

type PaginationEmbedded = gormx.PaginationEmbedded

type Pagination = gormx.Pagination
type Sorts = gormx.Sorts
type Sort = sqlx.Sort
type Range[T any] = gormx.Range[T]

type Id struct {
	Id uint64 `json:"id"`
}

type Cursor[T any] = sqlx.Cursor[T]
type CursorAny = Cursor[any]

type RangeAny = Range[any]

type List = sqlx.List

type Filters = sqlx.FilterExprs
type FilterMap = sqlx.FilterExprMap
