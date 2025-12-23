package cache

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	TYPE_LRU    = "lru"
	TYPE_LFU    = "lfu"
	TYPE_ARC    = "arc"
	TYPE_Simple = "simple"
)

const (
	// For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the Simple was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
)

var (
	KeyNotFoundError     = errors.New("key not found")
	KeyAlreadyExistError = errors.New("key already exist")
)

type Cache interface {
	Set(key, value any, expiration time.Duration) error
	SetNX(key, value any, expiration time.Duration) error
	Get(key any) (any, error)
	GetALL(checkExpired bool) map[any]any
	get(key any, onLoad bool) (any, error)
	Remove(key any) bool

	Keys(checkExpired bool) []any
	Len(checkExpired bool) int
	Has(key any) bool

	Purge()
	Flush()
	statsAccessor
}

type baseCache struct {
	size             int
	loaderFunc       LoaderFunc
	evictedFunc      EvictedFunc
	purgeVisitorFunc PurgeVisitorFunc
	addedFunc        AddedFunc
	expiration       time.Duration
	mu               sync.RWMutex
	loadGroup        Group
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

func New(size int) *CacheBuilder {
	return &CacheBuilder{
		tp:   TYPE_LRU,
		size: size,
	}
}

// Set a loader function with expiration.
// loaderFunc: create a new value with this function if cached value is expired.
// If nil returned instead of time.Duration from loaderFunc than value will never expire.
func (cb *CacheBuilder) LoaderFunc(loaderFunc LoaderFunc) *CacheBuilder {
	cb.loaderFunc = loaderFunc
	return cb
}

func (cb *CacheBuilder) EvictType(tp string) *CacheBuilder {
	cb.tp = tp
	return cb
}

func (cb *CacheBuilder) LRU() *CacheBuilder {
	return cb.EvictType(TYPE_LRU)
}

func (cb *CacheBuilder) LFU() *CacheBuilder {
	return cb.EvictType(TYPE_LFU)
}

func (cb *CacheBuilder) ARC() *CacheBuilder {
	return cb.EvictType(TYPE_ARC)
}

func (cb *CacheBuilder) Simple() *CacheBuilder {
	return cb.EvictType(TYPE_Simple)
}

func (cb *CacheBuilder) EvictedFunc(evictedFunc EvictedFunc) *CacheBuilder {
	cb.evictedFunc = evictedFunc
	return cb
}

func (cb *CacheBuilder) PurgeVisitorFunc(purgeVisitorFunc PurgeVisitorFunc) *CacheBuilder {
	cb.purgeVisitorFunc = purgeVisitorFunc
	return cb
}

func (cb *CacheBuilder) AddedFunc(addedFunc AddedFunc) *CacheBuilder {
	cb.addedFunc = addedFunc
	return cb
}

func (cb *CacheBuilder) Expiration(expiration time.Duration) *CacheBuilder {
	cb.expiration = expiration
	return cb
}

func (cb *CacheBuilder) Janitor(interval time.Duration) *CacheBuilder {
	cb.janitorInterval = interval
	return cb
}

func (cb *CacheBuilder) Build() Cache {
	if cb.tp != TYPE_Simple && cb.size <= 0 {
		panic("gcache: Cache size <= 0")
	}

	return cb.build()
}

func (cb *CacheBuilder) build() Cache {
	switch cb.tp {
	case TYPE_LRU:
		return newLRUCache(cb)
	case TYPE_LFU:
		return newLFUCache(cb)
	case TYPE_ARC:
		return newARCCache(cb)
	case TYPE_Simple:
		return newSimpleCache(cb)
	default:
		return newSimpleCache(cb)
	}
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

// load a new value using by specified key.
func (c *baseCache) load(key any, cb func(any, time.Duration, error) (any, error), isWait bool) (any, bool, error) {
	v, called, err := c.loadGroup.Do(key, func() (v any, e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("loader panics: %v", r)
			}
		}()
		return cb(c.loaderFunc(key))
	}, isWait)
	if err != nil {
		return nil, called, err
	}
	return v, called, nil
}

type janitor struct {
	interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c Cache) {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			c.Purge()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopJanitor(c *Simple) {
	c.janitor.stop <- true
}
