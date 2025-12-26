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

type Cache struct {
	baseCache
	mu          sync.RWMutex
	group       Group
	janitorstop chan bool
	store       Store
}

type item struct {
	key        any
	value      any
	expiration *time.Time
}

// Returns true if the item has expired.
func (it item) Expired(now *time.Time) bool {
	if it.expiration == nil {
		return false
	}
	if now == nil {
		return it.expiration.Before(time.Now())
	}
	return it.expiration.Before(*now)
}

type Store interface {
	init(cb *baseCache)
	set(key, value any, expiration *time.Time) (*item, error)
	get(key any, onLoad bool) (*item, error)
	has(key any, now *time.Time) bool
	remove(key any) bool
	length() int
	foreach(func(*item))
}

type baseCache struct {
	size             int
	loaderFunc       LoaderFunc
	evictedFunc      EvictedFunc
	clearVisitorFunc ClearVisitorFunc
	addedFunc        AddedFunc
	expiration       time.Duration
	*stats
}

type (
	LoaderFunc       func(any) (any, time.Duration, error)
	EvictedFunc      func(any, any)
	ClearVisitorFunc func(any, any)
	AddedFunc        func(any, any)
	DeserializeFunc  func(any, any) (any, error)
	SerializeFunc    func(any, any) (any, error)
)

type CacheBuilder struct {
	size             int
	loaderFunc       LoaderFunc
	evictedFunc      EvictedFunc
	clearVisitorFunc ClearVisitorFunc
	addedFunc        AddedFunc
	expiration       time.Duration
	janitorInterval  time.Duration
}

func New(size int) *CacheBuilder {
	return &CacheBuilder{
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

func (cb *CacheBuilder) EvictedFunc(evictedFunc EvictedFunc) *CacheBuilder {
	cb.evictedFunc = evictedFunc
	return cb
}

func (cb *CacheBuilder) ClearVisitorFunc(purgeVisitorFunc ClearVisitorFunc) *CacheBuilder {
	cb.clearVisitorFunc = purgeVisitorFunc
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

func (cb *CacheBuilder) LRU() *Cache {
	c := &Cache{store: &LRU{}}
	c.init(cb)
	return c
}

func (cb *CacheBuilder) LFU() *Cache {
	c := &Cache{store: &LFU{}}
	c.init(cb)
	return c
}

func (cb *CacheBuilder) ARC() *Cache {
	c := &Cache{store: &ARC{}}
	c.init(cb)
	return c
}

func (cb *CacheBuilder) Simple() *Cache {
	c := &Cache{store: &Simple{}}
	c.init(cb)
	return c
}

func (c *Cache) init(cb *CacheBuilder) {
	c.size = cb.size
	c.loaderFunc = cb.loaderFunc
	c.expiration = cb.expiration
	c.addedFunc = cb.addedFunc
	c.evictedFunc = cb.evictedFunc
	c.clearVisitorFunc = cb.clearVisitorFunc
	c.stats = &stats{}
	c.store.init(&c.baseCache)
	if cb.janitorInterval > 0 {
		c.startJanitor(cb.janitorInterval)
	}
}

// load a new value using by specified key.
func (c *Cache) load(key any, cb func(any, time.Duration, error) (*item, error), isWait bool) (*item, bool, error) {
	v, called, err := c.group.Do(key, func() (v any, e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("loader panics: %v", r)
			}
		}()
		c.mu.Lock()
		it, err := c.store.get(key, true)
		c.mu.Unlock()
		if err == nil {
			return it, nil
		}
		return cb(c.loaderFunc(key))
	}, isWait)
	if err != nil {
		return nil, called, err
	}
	return v.(*item), called, nil
}

func (c *Cache) startJanitor(interval time.Duration) {
	stop := make(chan bool)
	c.janitorstop = stop
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				c.Purge()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
	runtime.SetFinalizer(c, func(*Cache) {
		close(stop)
		stop <- true
	})
}

// Set a new key-value pair with an expiration time
func (c *Cache) Set(key, value any, expiration time.Duration) error {
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
	_, err := c.store.set(key, value, t)
	return err
}

// SetNX an item to the Cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *Cache) SetNX(k any, x any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.store.get(k, false)
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
	_, err = c.store.set(k, x, t)
	return err
}

// Get a value from cache pool using key if it exists. If not exists and it has LoaderFunc, it will generate the value using you have specified LoaderFunc method returns value.
func (c *Cache) Get(key any) (any, error) {
	c.mu.Lock()
	v, err := c.store.get(key, false)
	c.mu.Unlock()
	if err == nil {
		return v.value, err
	}
	if err == KeyNotFoundError {
		v, err = c.getWithLoader(key, true)
		if err != nil {
			return nil, err
		}
		return v.value, err
	}
	return nil, err
}

func (c *Cache) GetWithExpiration(key any) (any, time.Duration, error) {
	c.mu.Lock()
	v, err := c.store.get(key, false)
	c.mu.Unlock()
	if err == nil {
		expiration := NoExpiration
		if v.expiration != nil {
			expiration = v.expiration.Sub(time.Now())
		}
		return v.value, expiration, err
	}
	if err == KeyNotFoundError {
		value, err := c.getWithLoader(key, true)
		if err != nil {
			return nil, NoExpiration, err
		}
		expiration := NoExpiration
		if v.expiration != nil {
			expiration = v.expiration.Sub(time.Now())
		}
		return value.value, expiration, nil
	}
	return nil, NoExpiration, err
}

// Has checks if key exists in cache
func (c *Cache) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	return c.store.has(key, &now)
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.store.remove(key)
}

// GetALL returns all key-value pairs in the cache.
func (c *Cache) GetALL(checkExpired bool) map[any]any {
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

func (c *Cache) getWithLoader(key any, isWait bool) (*item, error) {
	if c.loaderFunc == nil {
		return nil, KeyNotFoundError
	}
	value, _, err := c.load(key, func(v any, expiration time.Duration, e error) (*item, error) {
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
		it, err := c.store.set(key, v, t)
		if err != nil {
			return nil, err
		}
		return it, nil
	}, isWait)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// GetIFPresent gets a value from cache pool using key if it exists.
// If it dose not exists key, returns KeyNotFoundError.
// And send a request which refresh value for specified key if cache object has LoaderFunc.
func (c *Cache) GetIFPresent(key any) (any, error) {
	c.mu.Lock()
	v, err := c.store.get(key, false)
	c.mu.Unlock()
	if err == KeyNotFoundError {
		return c.getWithLoader(key, false)
	}
	return v, err
}

// Keys returns a slice of the keys in the cache.
func (c *Cache) Keys(checkExpired bool) []any {
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
func (c *Cache) Len(checkExpired bool) int {
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
func (c *Cache) Purge() {
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

// Clear is used to completely clear the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.clearVisitorFunc != nil {
		c.store.foreach(func(item *item) {
			c.clearVisitorFunc(item.key, item.value)
		})
	}
	c.store.init(&c.baseCache)
}

// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n. To retrieve the incremented value, use one
// of the specialized methods, e.g. IncrementInt64.
func (c *Cache) Increment(k any, n int64) error {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	switch v.value.(type) {
	case int:
		v.value = v.value.(int) + int(n)
	case int8:
		v.value = v.value.(int8) + int8(n)
	case int16:
		v.value = v.value.(int16) + int16(n)
	case int32:
		v.value = v.value.(int32) + int32(n)
	case int64:
		v.value = v.value.(int64) + n
	case uint:
		v.value = v.value.(uint) + uint(n)
	case uintptr:
		v.value = v.value.(uintptr) + uintptr(n)
	case uint8:
		v.value = v.value.(uint8) + uint8(n)
	case uint16:
		v.value = v.value.(uint16) + uint16(n)
	case uint32:
		v.value = v.value.(uint32) + uint32(n)
	case uint64:
		v.value = v.value.(uint64) + uint64(n)
	case float32:
		v.value = v.value.(float32) + float32(n)
	case float64:
		v.value = v.value.(float64) + float64(n)
	default:
		c.mu.Unlock()
		return fmt.Errorf("the value for %v is not an integer", k)
	}
	c.mu.Unlock()
	return nil
}

// Increment an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to increment it by n. Pass a negative number to decrement the
// value. To retrieve the incremented value, use one of the specialized methods,
// e.g. IncrementFloat64.
func (c *Cache) IncrementFloat(k any, n float64) error {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	switch v.value.(type) {
	case float32:
		v.value = v.value.(float32) + float32(n)
	case float64:
		v.value = v.value.(float64) + n
	default:
		c.mu.Unlock()
		return fmt.Errorf("the value for %v does not have type float32 or float64", k)
	}
	c.mu.Unlock()
	return nil
}

// Increment an item of type int by n. Returns an error if the item's value is
// not an int, or if it was not found. If there is no error, the incremented
// value is returned.
func (c *Cache) IncrementInt(k any, n int) (int, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, fmt.Errorf("item %v not found", k)
	}
	rv, ok := v.value.(int)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an int", k)
	}
	nv := rv + n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type int32 by n. Returns an error if the item's value is
// not an int32, or if it was not found. If there is no error, the incremented
// value is returned.
func (c *Cache) IncrementInt32(k any, n int32) (int32, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(int32)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an int32", v.value)
	}
	nv := rv + n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type int64 by n. Returns an error if the item's value is
// not an int64, or if it was not found. If there is no error, the incremented
// value is returned.
func (c *Cache) IncrementInt64(k any, n int64) (int64, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(int64)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an int64", v.value)
	}
	nv := rv + n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint by n. Returns an error if the item's value is
// not an uint, or if it was not found. If there is no error, the incremented
// value is returned.
func (c *Cache) IncrementUint(k any, n uint) (uint, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(uint)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an uint", v.value)
	}
	nv := rv + n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uintptr by n. Returns an error if the item's value
// is not an uintptr, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Cache) IncrementUintptr(k any, n uintptr) (uintptr, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(uintptr)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an uintptr", v.value)
	}
	nv := rv + n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint32 by n. Returns an error if the item's value
// is not an uint32, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Cache) IncrementUint32(k any, n uint32) (uint32, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(uint32)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an uint32", v.value)
	}
	nv := rv + n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint64 by n. Returns an error if the item's value
// is not an uint64, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Cache) IncrementUint64(k any, n uint64) (uint64, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(uint64)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an uint64", v.value)
	}
	nv := rv + n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type float32 by n. Returns an error if the item's value
// is not an float32, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Cache) IncrementFloat32(k any, n float32) (float32, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(float32)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an float32", v.value)
	}
	nv := rv + n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type float64 by n. Returns an error if the item's value
// is not an float64, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Cache) IncrementFloat64(k any, n float64) (float64, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(float64)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an float64", v.value)
	}
	nv := rv + n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n. To retrieve the decremented value, use one
// of the specialized methods, e.g. DecrementInt64.
func (c *Cache) Decrement(k any, n int64) error {
	// TODO: Implement Increment and Decrement more cleanly.
	// (Cannot do Increment(k, n*-1) for uints.)
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	switch v.value.(type) {
	case int:
		v.value = v.value.(int) - int(n)
	case int8:
		v.value = v.value.(int8) - int8(n)
	case int16:
		v.value = v.value.(int16) - int16(n)
	case int32:
		v.value = v.value.(int32) - int32(n)
	case int64:
		v.value = v.value.(int64) - n
	case uint:
		v.value = v.value.(uint) - uint(n)
	case uintptr:
		v.value = v.value.(uintptr) - uintptr(n)
	case uint8:
		v.value = v.value.(uint8) - uint8(n)
	case uint16:
		v.value = v.value.(uint16) - uint16(n)
	case uint32:
		v.value = v.value.(uint32) - uint32(n)
	case uint64:
		v.value = v.value.(uint64) - uint64(n)
	case float32:
		v.value = v.value.(float32) - float32(n)
	case float64:
		v.value = v.value.(float64) - float64(n)
	default:
		c.mu.Unlock()
		return fmt.Errorf("the value for %v is not an integer", v.value)
	}
	c.mu.Unlock()
	return nil
}

// Decrement an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to decrement it by n. Pass a negative number to decrement the
// value. To retrieve the decremented value, use one of the specialized methods,
// e.g. DecrementFloat64.
func (c *Cache) DecrementFloat(k any, n float64) error {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	switch v.value.(type) {
	case float32:
		v.value = v.value.(float32) - float32(n)
	case float64:
		v.value = v.value.(float64) - n
	default:
		c.mu.Unlock()
		return fmt.Errorf("the value for %v does not have type float32 or float64", v.value)
	}
	c.mu.Unlock()
	return nil
}

// Decrement an item of type int by n. Returns an error if the item's value is
// not an int, or if it was not found. If there is no error, the decremented
// value is returned.
func (c *Cache) DecrementInt(k any, n int) (int, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(int)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an int", v.value)
	}
	nv := rv - n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int32 by n. Returns an error if the item's value is
// not an int32, or if it was not found. If there is no error, the decremented
// value is returned.
func (c *Cache) DecrementInt32(k any, n int32) (int32, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(int32)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an int32", v.value)
	}
	nv := rv - n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int64 by n. Returns an error if the item's value is
// not an int64, or if it was not found. If there is no error, the decremented
// value is returned.
func (c *Cache) DecrementInt64(k any, n int64) (int64, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(int64)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an int64", v.value)
	}
	nv := rv - n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint by n. Returns an error if the item's value is
// not an uint, or if it was not found. If there is no error, the decremented
// value is returned.
func (c *Cache) DecrementUint(k any, n uint) (uint, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(uint)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an uint", v.value)
	}
	nv := rv - n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uintptr by n. Returns an error if the item's value
// is not an uintptr, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Cache) DecrementUintptr(k any, n uintptr) (uintptr, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(uintptr)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an uintptr", v.value)
	}
	nv := rv - n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint32 by n. Returns an error if the item's value
// is not an uint32, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Cache) DecrementUint32(k any, n uint32) (uint32, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(uint32)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an uint32", v.value)
	}
	nv := rv - n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint64 by n. Returns an error if the item's value
// is not an uint64, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Cache) DecrementUint64(k any, n uint64) (uint64, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(uint64)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an uint64", v.value)
	}
	nv := rv - n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type float32 by n. Returns an error if the item's value
// is not an float32, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Cache) DecrementFloat32(k any, n float32) (float32, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(float32)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an float32", v.value)
	}
	nv := rv - n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type float64 by n. Returns an error if the item's value
// is not an float64, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Cache) DecrementFloat64(k any, n float64) (float64, error) {
	c.mu.Lock()
	v, err := c.store.get(k, false)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	rv, ok := v.value.(float64)
	if !ok {
		c.mu.Unlock()
		return 0, fmt.Errorf("the value for %v is not an float64", v.value)
	}
	nv := rv - n
	v.value = nv
	c.mu.Unlock()
	return nv, nil
}

func (c *Simple) has(key any, now *time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.Expired(now)
}
