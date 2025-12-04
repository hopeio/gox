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

func (r RangeType) HasBegin() bool {
	return r&RangeTypeHasBegin != 0
}

func (r RangeType) HasEnd() bool {
	return r&RangeTypeHasEnd != 0
}

func (r RangeType) ContainsBegin() bool {
	return r&RangeTypeContainsBegin != 0
}

func (r RangeType) ContainsEnd() bool {
	return r&RangeTypeContainsEnd != 0
}

const (
	RangeTypeContainsEnd RangeType = 1 << iota
	RangeTypeContainsBegin
	RangeTypeHasEnd
	RangeTypeHasBegin
)

type FilterType int8

func (f FilterType) RangeType() RangeType {
	if f == FilterTypeRange {
		return RangeType(f) & RangeTypeHasBegin
	}
	return RangeType(f)
}

const (
	FilterTypeEqual FilterType = iota
	FilterTypeNotEqual
	FilterTypeFuzzy
	FilterTypeIn
	FilterTypeNotIn
	FilterTypeIsNull
	FilterTypeIsNotNull
	FilterTypeRange = 16
)

type Cursor[T any] struct {
	Field string `json:"field,omitempty"`
	Prev  T      `json:"prev,omitempty"`
	Size  int    `json:"size,omitempty"`
}

type RangeInTwoField[T any] struct {
	BeginField string    `json:"beginField,omitempty"`
	EndField   string    `json:"endField,omitempty"`
	Begin      T         `json:"begin"`
	End        T         `json:"end"`
	Type       RangeType `json:"type,omitempty"`
}

type CursorAny = Cursor[any]

type RangeAny = Range[any]

type RangeInTwoFieldAny = RangeInTwoField[any]

type List struct {
	PaginationEmbedded
	Filter map[string]Field[any] `json:"filter,omitempty"`
}

type Field[T any] struct {
	Field  string    `json:"field,omitempty"`
	Type   RangeType `json:"type,omitempty"`
	Value  T         `json:"value,omitempty"`
	Values []T       `json:"values,omitempty"`
}

type Equal[T any] struct {
	Field string `json:"field,omitempty"`
	Value T      `json:"value,omitempty"`
}

type In[T any] struct {
	Field  string `json:"field,omitempty"`
	Values []T    `json:"values,omitempty"`
}
