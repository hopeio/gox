/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import (
	jsonx "github.com/hopeio/gox/encoding/json"
)

// 返回option 返回时会有两次复制value,后续使用还有可能更多次,自行选择用不用
type Option[T any] struct {
	value T
	ok    bool
}

func Some[T any](v T) Option[T] {
	return Option[T]{value: v, ok: true}
}

func None[T any]() Option[T] {
	return Option[T]{ok: false}
}
func Nil[T any]() Option[T] {
	return Option[T]{ok: false}
}

func (opt *Option[T]) Val() (T, bool) {
	return opt.value, opt.ok
}

func (opt *Option[T]) Get() (T, bool) {
	return opt.value, opt.ok
}

func (opt *Option[T]) IsNone() bool {
	return !opt.ok
}

func (opt *Option[T]) IsSome() bool {
	return opt.ok
}

func (opt *Option[T]) Unwrap() T {
	if opt.IsNone() {
		panic("Attempted to unwrap an empty Option.")
	}
	return opt.value
}

func (opt *Option[T]) UnwrapOr(def T) T {
	if opt.IsSome() {
		return opt.Unwrap()
	}
	return def
}

func (opt *Option[T]) UnwrapOrElse(fn func() T) T {
	if opt.IsSome() {
		return opt.Unwrap()
	}
	return fn()
}

func MapOption[T any, R any](opt Option[T], fn func(T) R) Option[R] {
	if !opt.IsSome() {
		return None[R]()
	}
	return Some(fn(opt.Unwrap()))
}

func (opt *Option[T]) IfSome(action func(value T)) {
	if opt.ok {
		action(opt.value)
	}
}

func (opt *Option[T]) IfNone(action func()) {
	if !opt.ok {
		action()
	}
}

func (opt *Option[T]) MarshalJSON() ([]byte, error) {
	if opt.ok {
		return jsonx.Marshal(opt.value)
	}
	return []byte("null"), nil
}

func (opt *Option[T]) UnmarshalJSON(data []byte) error {
	if len(data) < 5 && string(data) == "null" {
		opt.ok = false
		return nil
	}
	opt.ok = true
	return jsonx.Unmarshal(data, &opt.value)
}

type OptionPtr[T any] struct {
	value *T
}

func SomePtr[T any](v *T) OptionPtr[T] {
	return OptionPtr[T]{value: v}
}

func NonePtr[T any]() OptionPtr[T] {
	return OptionPtr[T]{}
}
func NilPtr[T any]() OptionPtr[T] {
	return OptionPtr[T]{}
}

func (opt OptionPtr[T]) Val() (*T, bool) {
	if opt.value == nil {
		return nil, false
	}
	return opt.value, true
}

func (opt OptionPtr[T]) Get() (*T, bool) {
	if opt.value == nil {
		return nil, false
	}
	return opt.value, true
}

func (opt OptionPtr[T]) IsNone() bool {
	return opt.value == nil
}

func (opt OptionPtr[T]) IsSome() bool {
	return opt.value != nil
}

func (opt OptionPtr[T]) Unwrap() *T {
	if opt.IsNone() {
		panic("Attempted to unwrap an empty OptionPtr.")
	}
	return opt.value
}

func (opt OptionPtr[T]) UnwrapOr(def *T) *T {
	if opt.IsSome() {
		return opt.Unwrap()
	}
	return def
}

func (opt OptionPtr[T]) UnwrapOrElse(fn func() *T) *T {
	if opt.IsSome() {
		return opt.Unwrap()
	}
	return fn()
}

func MapOptionPtr[T any, R any](opt OptionPtr[T], fn func(*T) *R) OptionPtr[R] {
	if !opt.IsSome() {
		return NonePtr[R]()
	}
	return SomePtr(fn(opt.Unwrap()))
}

func (opt OptionPtr[T]) IfSome(action func(value *T)) {
	if opt.IsSome() {
		action(opt.value)
	}
}

func (opt OptionPtr[T]) IfNone(action func()) {
	if opt.IsNone() {
		action()
	}
}

func (opt OptionPtr[T]) MarshalJSON() ([]byte, error) {
	if opt.IsSome() {
		return jsonx.Marshal(opt.value)
	}
	return []byte("null"), nil
}

func (opt *OptionPtr[T]) UnmarshalJSON(data []byte) error {
	if len(data) < 5 && string(data) == "null" {
		return nil
	}
	return jsonx.Unmarshal(data, &opt.value)
}
