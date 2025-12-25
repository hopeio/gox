package cache

import (
	"errors"
	"fmt"
	"runtime"
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

type Cache[T Store] struct {
	baseCache
	store T
}

type Store interface {
	init(size int)
	set(key, value any, expiration *time.Time) error
	get(key any) (any, error)
	has(key any, now *time.Time) bool
	remove(key any) bool
	length() int
	foreach(func(*item))
}

type baseCache struct {
	size             int
	loaderFunc       LoaderFunc
	evictedFunc      EvictedFunc
	purgeVisitorFunc PurgeVisitorFunc
	addedFunc        AddedFunc
	expiration       time.Duration
	mu               sync.RWMutex
	m     map[any]*call // lazily initialized
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

func (cb *CacheBuilder) LRU() *Cache[*LRU] {
	c := &Cache[*LRU]{
		store: &LRU{},
	}
	buildCache(&c.baseCache, cb)
	c.store.init(cb.size)
	if c.janitor != nil {
		go c.janitor.Run(c)
		runtime.SetFinalizer(c, stopJanitor)
	}
	return c
}

func (cb *CacheBuilder) LFU() *Cache[*LFU] {
	c := &Cache[*LFU]{
		store: &LFU{},
	}
	buildCache(&c.baseCache, cb)
	c.store.init(cb.size)
	if c.janitor != nil {
		go c.janitor.Run(c)
		runtime.SetFinalizer(c, stopJanitor)
	}
	return c
}

func (cb *CacheBuilder) ARC() *Cache[*ARC] {
	c := &Cache[*ARC]{
		store: &ARC{},
	}
	buildCache(&c.baseCache, cb)
	c.store.init(cb.size)
	if c.janitor != nil {
		go c.janitor.Run(c)
		runtime.SetFinalizer(c, stopJanitor)
	}
	return c
}

func (cb *CacheBuilder) Simple() *CacheBuilder {
	return cb.EvictType(TYPE_Simple)
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
func (c *Cache[T]) load(key any, cb func(any, time.Duration, error) (any, error), isWait bool) (any, bool, error) {
	v, called, err := c.Do(key, func() (v any, e error) {
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

func (j *janitor) Run(c interface{ Purge() }) {
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

// Set a new key-value pair with an expiration time
func (c *Cache[T]) Set(key, value any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var t *time.Time
	if expiration == 0 && c.expiration > 0 {
		exp := time.Now().Add(c.expiration)
		t = &exp
	} else if expiration > 0 {
		exp := time.Now().Add(expiration)
		t = &exp
	}
	return c.store.set(key, value, t)
}

// SetNX an item to the Cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *Cache[T]) SetNX(k any, x any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.store.get(k)
	if err == nil {
		return KeyAlreadyExistError
	}
	var t *time.Time
	if expiration == 0 && c.expiration > 0 {
		exp := time.Now().Add(c.expiration)
		t = &exp
	} else if expiration > 0 {
		exp := time.Now().Add(expiration)
		t = &exp
	}
	return c.store.set(k, x, t)
}

// Get a value from cache pool using key if it exists. If not exists and it has LoaderFunc, it will generate the value using you have specified LoaderFunc method returns value.
func (c *Cache[T]) Get(key any) (any, error) {
	c.mu.Lock()
	v, err := c.store.get(key)
	c.mu.Unlock()
	if err == KeyNotFoundError {
		return c.store.getWithLoader(key, true)
	}
	return v, err
}

// Has checks if key exists in cache
func (c *Cache[T]) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	return c.store.has(key, &now)
}

// Remove removes the provided key from the cache.
func (c *Cache[T]) Remove(key any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.store.remove(key)
}

// GetALL returns all key-value pairs in the cache.
func (c *Cache[T]) GetALL(checkExpired bool) map[any]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make(map[any]any, c.store.length())
	now := time.Now()
	c.store.foreach(func(item *item) {
		if !checkExpired || !item.Expired(&now) {
			items[item.key] = item.value
		}
	})
	return items
}

func (c *Cache[T]) getWithLoader(key any, isWait bool) (any, error) {
	if c.loaderFunc == nil {
		return nil, KeyNotFoundError
	}
	value, _, err := c.load(key, func(v any, expiration time.Duration, e error) (any, error) {
		if e != nil {
			return nil, e
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		var t *time.Time
		if expiration == 0 && c.expiration > 0 {
			exp := time.Now().Add(c.expiration)
			t = &exp
		} else if expiration > 0 {
			exp := time.Now().Add(expiration)
			t = &exp
		}
		err := c.store.set(key, v, t)
		if err != nil {
			return nil, err
		}
		return v, nil
	}, isWait)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// GetIFPresent gets a value from cache pool using key if it exists.
// If it dose not exists key, returns KeyNotFoundError.
// And send a request which refresh value for specified key if cache object has LoaderFunc.
func (c *Cache[T]) GetIFPresent(key any) (any, error) {
	c.mu.Lock()
	v, err := c.store.get(key)
	c.mu.Unlock()
	if err == KeyNotFoundError {
		return c.store.getWithLoader(key, false)
	}
	return v, err
}

// Keys returns a slice of the keys in the cache.
func (c *Cache[T]) Keys(checkExpired bool) []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]any, 0, c.store.length())
	now := time.Now()
	c.store.foreach(func(item *item) {
		if !checkExpired || !item.Expired(&now) {
			keys = append(keys, item.key)
		}
	})
	return keys
}

// Len returns the number of items in the cache.
func (c *Cache[T]) Len(checkExpired bool) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !checkExpired {
		return c.store.length()
	}
	var length int
	now := time.Now()
	c.store.foreach(func(item *item) {
		if !item.Expired(&now) {
			length++
		}
	})
	return length
}

// Purge is used to completely clear the cache
func (c *Cache[T]) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	c.store.foreach(func(item *item) {
		if item.Expired(&now) {
			c.store.remove(item.key)
			if c.evictedFunc != nil {
				c.evictedFunc(item.key, item.value)
			}
		}
	})
}

// Flush is used to completely clear the cache
func (c *Cache[T]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store.init(c.size)
}
