/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import (
	jsonx "github.com/hopeio/gox/encoding/json"
)

type Result[T any] struct {
	value T
	err   error
}

func Ok[T any](a T) Result[T] {
	return Result[T]{value: a}
}

func Err[T any](a error) Result[T] {
	return Result[T]{err: a}
}

func (a Result[T]) Val() (value T, err error) {
	return a.value, a.err
}

func (a Result[T]) OrPanic() T {
	if a.err != nil {
		panic("error of result")
	}
	return a.value
}

func (a Result[T]) Or(value T) T {
	if a.err != nil {
		return value
	}
	return a.value
}

func (a Result[T]) OrDefault() (v T) {
	if a.err != nil {
		return
	}
	return a.value
}

func (a Result[T]) IsOk() bool {
	return a.err == nil
}

func (a Result[T]) IsOkAnd(f func(T) bool) bool {
	if a.err != nil {
		return false
	}
	return f(a.value)
}

func (a Result[T]) IsErr() bool {
	return a.err != nil
}

func (a Result[T]) IsErrAnd(f func(error) bool) bool {
	if a.err == nil {
		return false
	}
	return f(a.err)
}

func (a Result[T]) IfOk(action func(value T)) {
	if a.err == nil {
		action(a.value)
	}
}

func (a Result[T]) IfErr(action func(err error)) {
	if a.err != nil {
		action(a.err)
	}
}

func (a *Result[T]) MarshalJSON() ([]byte, error) {
	if a.err == nil {
		return jsonx.Marshal(a.value)
	}
	return []byte("null"), a.err
}

func (a *Result[T]) UnmarshalJSON(data []byte) error {
	if len(data) < 5 && string(data) == "null" {
		return nil
	}
	return jsonx.Unmarshal(data, &a.value)
}

func ResultVal[T any](v T, err error) T {
	return v
}
