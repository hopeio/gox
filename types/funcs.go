/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import "context"

// Supplier produces an element.
type Supplier[T any] func() T

// Consumer consumes an element.
type Consumer[T any] func(T)

// UnaryFunction converts a value from one type to another.
type UnaryFunction[T, R any] func(T) R

// Predicate checks whether a value satisfies a condition.
type Predicate[T any] func(T) bool

// UnaryOperator applies a unary operation and returns a value of the same type.
type UnaryOperator[T any] func(T) T

// BinaryFunction converts two inputs to a third type.
type BinaryFunction[T, R, U any] func(T, R) U

// BinaryOperator applies a binary operation to two values of the same type and returns the same type.
type BinaryOperator[T any] func(T, T) T

// Comparator compares two elements.
// If the first element is greater than the second, it returns a positive number;
// if the first element is less than the second, it returns a negative number;
// otherwise it returns 0.
type Comparator[T any] func(T, T) int

type Less[T any] func(T, T) bool

// SupplierKV produces a key/value pair.
type SupplierKV[K, V any] func() (K, V)

// UnaryKVFunction converts a key/value pair to another value.
type UnaryKVFunction[K, V, R any] func(K, V) R
type UnaryKVFunction2[K, V, RK, RV any] func(K, V) (RK, RV)

// Predicate checks whether a key/value pair satisfies a condition.
type PredicateKV[K, V any] func(K, V) bool

type UnaryKVOperator[K, V any] func(K, V) (K, V)

type BinaryKVFunction[K, V, R, U any] func(K, V, R) U
type BinaryKVFunction2[K, V, RK, RV, UK, UV any] func(K, V, RK, RV) (UK, UV)

type BinaryKVOperator[K, V any] func(K, V, K, V) (K, V)

// Comparator compares two elements.
// If the first element is greater than the second, it returns a positive number;
// if the first element is less than the second, it returns a negative number;
// otherwise it returns 0.
type ComparatorKV[K, V any] func(K, V, K, V) int
type LessKV[K, V any] func(K, V, K, V) bool

// ConsumerKV consumes a key/value pair.
type ConsumerKV[K, V any] func(K, V)

type Service[REQ, RESP any] func(context.Context, REQ) (RESP, error)

type Func func()
type FuncReturnErr func() error
type FuncReturnDataOrErr[T any] func() (T, error)
type FuncRetry func(times uint) (retry bool)

// Do performs the retry logic.
func (f FuncRetry) Do(times uint) (retry bool) {
	return f(times)
}

type Task func(context.Context)
type TaskReturnErr func(context.Context) error
