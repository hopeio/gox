/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package container

import (
	"time"

	"github.com/hopeio/gox/types/constraints"
)

type Get[K constraints.Key, V any] interface {
	Get(key K) V
}

type Set[K constraints.Key, V any] interface {
	Set(key K, v V)
}

type SetWithExpire[K constraints.Key, V any] interface {
	SetWithExpire(key K, v V, expire time.Duration)
}

type Delete[K constraints.Key, V any] interface {
	Delete(key K)
}

type StoreWithExpire[K constraints.Key, V any] interface {
	SetWithExpire[K, V]
	Get[K, V]
	Delete[K, V]
}

type Store[K constraints.Key, V any] interface {
	Set[K, V]
	Get[K, V]
	Delete[K, V]
}
