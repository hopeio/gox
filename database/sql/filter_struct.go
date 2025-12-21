package sql

import (
	"time"

	"golang.org/x/exp/constraints"
)

type Ordered interface {
	constraints.Ordered | time.Time | ~*time.Time
}

type SortType uint8

const (
	SortTypeDesc SortType = iota
	SortTypeAsc
)

type PaginationEmbedded struct {
	PageNo   uint32 `json:"pageNo"`
	PageSize uint32 `json:"pageSize"`
}

type Pagination struct {
	No   uint32 `json:"no"`
	Size uint32 `json:"size"`
}

type Sort struct {
	Field string   `json:"field"`
	Type  SortType `json:"type,omitempty"`
}

type SortExpr struct {
	Sql                string `json:"sql"`
	Var                any    `json:"var"`
	WithoutParenthesis bool   `json:"withoutParenthesis"`
}

type Sorts []Sort

type Range[T any] struct {
	Field string    `json:"field,omitempty"`
	Begin T         `json:"begin"`
	End   T         `json:"end"`
	Type  RangeMode `json:"type,omitempty"`
}

type Id struct {
	Id uint64 `json:"id"`
}

type RangeMode int8

func (r RangeMode) HasBegin() bool {
	return r&RangeModeHasBegin != 0
}

func (r RangeMode) HasEnd() bool {
	return r&RangeModeHasEnd != 0
}

func (r RangeMode) ContainsBegin() bool {
	return r&RangeModeContainsBegin != 0
}

func (r RangeMode) ContainsEnd() bool {
	return r&RangeModeContainsEnd != 0
}

const (
	RangeModeContainsEnd RangeMode = 1 << iota
	RangeModeContainsBegin
	RangeModeHasEnd
	RangeModeHasBegin
)

type Cursor[T any] struct {
	Field string `json:"field,omitempty"`
	Prev  T      `json:"prev,omitempty"`
	Size  int    `json:"size,omitempty"`
}

type List struct {
	Pagination Pagination  `json:"page,omitempty"`
	Sort       Sorts       `json:"sort,omitempty"`
	Filters    FilterExprs `json:"filters,omitempty"`
}

type FilterExprMap map[string]FilterExpr
