package v2

import (
	"sync"
	"time"
)

type Cache[T Store] struct {
	baseCache
	store T
}

type baseCache struct {
	size             int
	loaderFunc       LoaderFunc
	evictedFunc      EvictedFunc
	purgeVisitorFunc PurgeVisitorFunc
	addedFunc        AddedFunc
	expiration       time.Duration
	mu               sync.RWMutex
	janitor          *janitor
	*stats
}
type (
	LoaderFunc       func(any) (any, time.Duration, error)
	EvictedFunc      func(any, any)
	PurgeVisitorFunc func(any, any)
	AddedFunc        func(any, any)
	DeserializeFunc  func(any, any) (any, error)
	SerializeFunc    func(any, any) (any, error)
)

type Store interface {
	init(size int)
	set(key, value any, expiration time.Duration) error
	get(key any, onLoad bool) (any, error)
	has(key any, now *time.Time) bool
	remove(key any) bool
}

func NewCache[T Store](size int, cb *CacheBuilder, store T) *Cache[T] {
	var cache Cache[T]
	buildCache(&cache.baseCache, cb)
	cache.store = store
	cache.store.init(size)
	return &cache
}

type CacheBuilder struct {
	tp               string
	size             int
	loaderFunc       LoaderFunc
	evictedFunc      EvictedFunc
	purgeVisitorFunc PurgeVisitorFunc
	addedFunc        AddedFunc
	expiration       time.Duration
	janitorInterval  time.Duration
}

type stats struct {
	hitCount  uint64
	missCount uint64
}

type janitor struct {
	interval time.Duration
	stop     chan bool
}

func buildCache(c *baseCache, cb *CacheBuilder) {
	c.size = cb.size
	c.loaderFunc = cb.loaderFunc
	c.expiration = cb.expiration
	c.addedFunc = cb.addedFunc
	c.evictedFunc = cb.evictedFunc
	c.purgeVisitorFunc = cb.purgeVisitorFunc
	c.stats = &stats{}
	if cb.janitorInterval > 0 {
		c.janitor = &janitor{interval: cb.janitorInterval, stop: make(chan bool)}
	}
}
