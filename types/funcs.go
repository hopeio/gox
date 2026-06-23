/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

import "context"

// Supplier 产生一个元素
type Supplier[T any] func() T

// Consumer 消费一个元素
type Consumer[T any] func(T)

// UnaryFunction 将一个类型转为另一个类型
type UnaryFunction[T, R any] func(T) R

// Predicate 断言是否满足指定条件
type Predicate[T any] func(T) bool

// UnaryOperator 对输入进行一元运算返回相同类型的结果
type UnaryOperator[T any] func(T) T

// BinaryFunction 将两个类型转为第三个类型
type BinaryFunction[T, R, U any] func(T, R) U

// BinaryOperator 输入两个相同类型的参数，对其做二元运算，返回相同类型的结果
type BinaryOperator[T any] func(T, T) T

// Comparator 比较两个元素.
// 第一个元素大于第二个元素时，返回正数;
// 第一个元素小于第二个元素时，返回负数;
// 否则返回 0.
type Comparator[T any] func(T, T) int

type Less[T any] func(T, T) bool

// SupplierKV 产生一个KV
type SupplierKV[K, V any] func() (K, V)

// UnaryKVFunction 将一个类型转为另一个类型
type UnaryKVFunction[K, V, R any] func(K, V) R
type UnaryKVFunction2[K, V, RK, RV any] func(K, V) (RK, RV)

// Predicate 断言是否满足指定条件
type PredicateKV[K, V any] func(K, V) bool

type UnaryKVOperator[K, V any] func(K, V) (K, V)

type BinaryKVFunction[K, V, R, U any] func(K, V, R) U
type BinaryKVFunction2[K, V, RK, RV, UK, UV any] func(K, V, RK, RV) (UK, UV)

type BinaryKVOperator[K, V any] func(K, V, K, V) (K, V)

// Comparator 比较两个元素.
// 第一个元素大于第二个元素时，返回正数;
// 第一个元素小于第二个元素时，返回负数;
// 否则返回 0.
type ComparatorKV[K, V any] func(K, V, K, V) int
type LessKV[K, V any] func(K, V, K, V) bool

// ConsumerKV 消费一个KV
type ConsumerKV[K, V any] func(K, V)


type Service[REQ, RESP any] func(context.Context, REQ) (RESP, error)

type Func func()
type FuncReturnErr func() error
type FuncReturnDataOrErr[T any] func() (T, error)
type FuncRetry func(times uint) (retry bool)

func (f FuncRetry) Do(times uint) (retry bool) {
	return f(times)
}

type Task func(context.Context)
type TaskReturnErr func(context.Context) error

