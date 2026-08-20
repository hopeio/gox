/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import (
	jsonx "github.com/hopeio/gox/encoding/json"
)

// Returning an Option can copy the value multiple times; decide whether you need it.
type Option[T any] struct {
	value T
	ok    bool
}

// Some creates an Option containing v.
func Some[T any](v T) Option[T] {
	return Option[T]{value: v, ok: true}
}

// None creates an empty Option.
func None[T any]() Option[T] {
	return Option[T]{ok: false}
}

// Nil creates an empty Option.
func Nil[T any]() Option[T] {
	return Option[T]{ok: false}
}

// Val returns the value and whether it is present.
func (opt *Option[T]) Val() (T, bool) {
	return opt.value, opt.ok
}

// Get returns the value and whether it is present.
func (opt *Option[T]) Get() (T, bool) {
	return opt.value, opt.ok
}

// IsNone reports whether the condition holds.
func (opt *Option[T]) IsNone() bool {
	return !opt.ok
}

// IsSome reports whether the condition holds.
func (opt *Option[T]) IsSome() bool {
	return opt.ok
}

// Unwrap returns the value or panics if the Option is empty.
func (opt *Option[T]) Unwrap() T {
	if opt.IsNone() {
		panic("Attempted to unwrap an empty Option.")
	}
	return opt.value
}

// UnwrapOr returns the value or def if the Option is empty.
func (opt *Option[T]) UnwrapOr(def T) T {
	if opt.IsSome() {
		return opt.Unwrap()
	}
	return def
}

// UnwrapOrElse returns the value or calls fn if the Option is empty.
func (opt *Option[T]) UnwrapOrElse(fn func() T) T {
	if opt.IsSome() {
		return opt.Unwrap()
	}
	return fn()
}

// MapOption maps the contained value to another type.
func MapOption[T any, R any](opt Option[T], fn func(T) R) Option[R] {
	if !opt.IsSome() {
		return None[R]()
	}
	return Some(fn(opt.Unwrap()))
}

// IfSome runs action if the Option contains a value.
func (opt *Option[T]) IfSome(action func(value T)) {
	if opt.ok {
		action(opt.value)
	}
}

// IfNone runs action if the Option is empty.
func (opt *Option[T]) IfNone(action func()) {
	if !opt.ok {
		action()
	}
}

// MarshalJSON encodes the value, or null if the Option is empty.
func (opt *Option[T]) MarshalJSON() ([]byte, error) {
	if opt.ok {
		return jsonx.Marshal(opt.value)
	}
	return []byte("null"), nil
}

// UnmarshalJSON decodes the value, or leaves the Option empty for null.
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

// SomePtr creates an OptionPtr containing v.
func SomePtr[T any](v *T) OptionPtr[T] {
	return OptionPtr[T]{value: v}
}

// NonePtr creates an empty OptionPtr.
func NonePtr[T any]() OptionPtr[T] {
	return OptionPtr[T]{}
}

// NilPtr creates an empty OptionPtr.
func NilPtr[T any]() OptionPtr[T] {
	return OptionPtr[T]{}
}

// Val returns the value and whether it is present.
func (opt OptionPtr[T]) Val() (*T, bool) {
	if opt.value == nil {
		return nil, false
	}
	return opt.value, true
}

// Get returns the value and whether it is present.
func (opt OptionPtr[T]) Get() (*T, bool) {
	if opt.value == nil {
		return nil, false
	}
	return opt.value, true
}

// IsNone reports whether the condition holds.
func (opt OptionPtr[T]) IsNone() bool {
	return opt.value == nil
}

// IsSome reports whether the condition holds.
func (opt OptionPtr[T]) IsSome() bool {
	return opt.value != nil
}

// Unwrap returns the value or panics if the OptionPtr is empty.
func (opt OptionPtr[T]) Unwrap() *T {
	if opt.IsNone() {
		panic("Attempted to unwrap an empty OptionPtr.")
	}
	return opt.value
}

// UnwrapOr returns the value or def if the OptionPtr is empty.
func (opt OptionPtr[T]) UnwrapOr(def *T) *T {
	if opt.IsSome() {
		return opt.Unwrap()
	}
	return def
}

// UnwrapOrElse returns the value or calls fn if the OptionPtr is empty.
func (opt OptionPtr[T]) UnwrapOrElse(fn func() *T) *T {
	if opt.IsSome() {
		return opt.Unwrap()
	}
	return fn()
}

// Map maps the contained pointer value to another pointer type using a
// method-level generic (Go 1.27): the result type R is independent of T.
func (opt OptionPtr[T]) Map[R any](fn func(*T) *R) OptionPtr[R] {
	if !opt.IsSome() {
		return NonePtr[R]()
	}
	return SomePtr(fn(opt.Unwrap()))
}

// MapOptionPtr maps the contained value to another pointer type.
// Deprecated: use OptionPtr[T].Map instead.
func MapOptionPtr[T any, R any](opt OptionPtr[T], fn func(*T) *R) OptionPtr[R] {
	return opt.Map[R](fn)
}

// IfSome runs action if the OptionPtr contains a value.
func (opt OptionPtr[T]) IfSome(action func(value *T)) {
	if opt.IsSome() {
		action(opt.value)
	}
}

// IfNone runs action if the OptionPtr is empty.
func (opt OptionPtr[T]) IfNone(action func()) {
	if opt.IsNone() {
		action()
	}
}

// MarshalJSON encodes the value, or null if the OptionPtr is empty.
func (opt OptionPtr[T]) MarshalJSON() ([]byte, error) {
	if opt.IsSome() {
		return jsonx.Marshal(opt.value)
	}
	return []byte("null"), nil
}

// UnmarshalJSON decodes the value, or leaves the OptionPtr empty for null.
func (opt *OptionPtr[T]) UnmarshalJSON(data []byte) error {
	if len(data) < 5 && string(data) == "null" {
		return nil
	}
	return jsonx.Unmarshal(data, &opt.value)
}
