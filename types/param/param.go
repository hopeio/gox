/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package param

type Pageable interface {
	PageNo() int
	PageSize() int
	Sort() []Sort
}

type SortType int

const (
	_ SortType = iota
	SortTypeAsc
	SortTypeDesc
)

type PaginationEmbedded struct {
	PageNo   int    `json:"pageNo"`
	PageSize int    `json:"pageSize"`
	Sort     []Sort `json:"sort"`
}

type Pagination struct {
	No   int    `json:"no"`
	Size int    `json:"size"`
	Sort []Sort `json:"sort"`
}

type Sort struct {
	Field string   `json:"field"`
	Type  SortType `json:"type,omitempty"`
}

type Range[T any] struct {
	Field string    `json:"field,omitempty"`
	Begin T         `json:"begin"`
	End   T         `json:"end"`
	Type  RangeType `json:"type,omitempty"`
}

type Id struct {
	Id uint `json:"id"`
}

type RangeType int8

func (r RangeType) ContainsBegin() bool {
	return r&RangeTypeContainsBegin != 0
}

func (r RangeType) ContainsEnd() bool {
	return r&RangeTypeContainsEnd != 0
}

const (
	RangeTypeContainsBegin RangeType = 1 << iota
	RangeTypeContainsEnd
)

type FilterType int8

func (f FilterType) RangeType() RangeType {
	switch f {
	case FilterTypeRange:
		return RangeTypeContainsBegin | RangeTypeContainsEnd
	case FilterTypeRangeContainsBegin:
		return RangeTypeContainsBegin
	case FilterTypeRangeContainsEnd:
		return RangeTypeContainsEnd
	default:
		return 0
	}
}

const (
	FilterTypeEqual FilterType = iota
	FilterTypeNotEqual
	FilterTypeFuzzy
	FilterTypeIn
	FilterTypeNotIn
	FilterTypeIsNull
	FilterTypeIsNotNull
	FilterTypeRange
	FilterTypeRangeContainsBegin
	FilterTypeRangeContainsEnd
	FilterTypeOr
)

type Cursor[T any] struct {
	Field string `json:"field,omitempty"`
	Prev  T      `json:"prev,omitempty"`
	Size  int    `json:"size,omitempty"`
}

type CursorAny = Cursor[any]

type RangeAny = Range[any]

type List struct {
	PaginationEmbedded
	Filters Filters `json:"filters,omitempty"`
}

type Filters []Filter
type FilterMap map[string]Filter

type Filter struct {
	Field  string     `json:"field,omitempty"`
	Type   FilterType `json:"type,omitempty"`
	Value  any        `json:"value,omitempty"`
	Values []any      `json:"values,omitempty"`
}
