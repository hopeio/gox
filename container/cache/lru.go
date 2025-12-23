package cache

import (
	"container/list"
	"runtime"
	"time"
)

// Discards the least recently used items first.
type LRU struct {
	baseCache
	items     map[any]*list.Element
	evictList *list.List
}

func newLRUCache(cb *CacheBuilder) *LRU {
	c := &LRU{}
	buildCache(&c.baseCache, cb)

	c.init()
	c.loadGroup.cache = c
	if c.janitor != nil {
		go c.janitor.Run(c)
		runtime.SetFinalizer(c, stopJanitor)
	}
	return c
}

func (c *LRU) init() {
	c.evictList = list.New()
	c.items = make(map[any]*list.Element, c.size+1)
}

func (c *LRU) set(key, value any, expiration time.Duration) (*lruItem, error) {
	// Check for existing item
	var item *lruItem
	if it, ok := c.items[key]; ok {
		c.evictList.MoveToFront(it)
		item = it.Value.(*lruItem)
		item.value = value
	} else {
		// Verify size not exceeded
		if c.evictList.Len() >= c.size {
			c.evict(1)
		}
		item = &lruItem{
			key:   key,
			value: value,
		}
		c.items[key] = c.evictList.PushFront(item)
	}

	if expiration == 0 && c.expiration > 0 {
		t := time.Now().Add(c.expiration)
		item.expiration = &t
	} else if expiration > 0 {
		t := time.Now().Add(expiration)
		item.expiration = &t
	}

	if c.addedFunc != nil {
		c.addedFunc(key, value)
	}

	return item, nil
}

// Set a new key-value pair with an expiration time
func (c *LRU) Set(key, value any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.set(key, value, expiration)
	return err
}

// SetNX an item to the Simple only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *LRU) SetNX(k any, x any, d time.Duration) error {
	c.mu.Lock()
	_, err := c.get(k, false)
	if err == nil {
		c.mu.Unlock()
		return KeyAlreadyExistError
	}
	_, err = c.set(k, x, d)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()
	return nil
}

// Get a value from cache pool using key if it exists.
// If it dose not exists key and has LoaderFunc,
// generate a value using `LoaderFunc` method returns value.
func (c *LRU) Get(key any) (any, error) {
	c.mu.Lock()
	v, err := c.get(key, false)
	if err == KeyNotFoundError {
		c.mu.Unlock()
		return c.getWithLoader(key, true)
	}
	return v, err
}

// GetIFPresent gets a value from cache pool using key if it exists.
// If it dose not exists key, returns KeyNotFoundError.
// And send a request which refresh value for specified key if cache object has LoaderFunc.
func (c *LRU) GetIFPresent(key any) (any, error) {
	c.mu.Lock()
	v, err := c.get(key, false)
	if err == KeyNotFoundError {
		c.mu.Unlock()
		return c.getWithLoader(key, false)
	}
	return v, err
}

func (c *LRU) get(key any, onLoad bool) (any, error) {
	item, ok := c.items[key]
	if ok {
		it := item.Value.(*lruItem)
		if !it.Expired(nil) {
			c.evictList.MoveToFront(item)
			v := it.value

			if !onLoad {
				c.stats.IncrHitCount()
			}
			return v, nil
		}
		c.removeElement(item)
	}
	if !onLoad {
		c.stats.IncrMissCount()
	}
	return nil, KeyNotFoundError
}

func (c *LRU) getWithLoader(key any, isWait bool) (any, error) {
	if c.loaderFunc == nil {
		return nil, KeyNotFoundError
	}
	value, _, err := c.load(key, func(v any, expiration time.Duration, e error) (any, error) {
		if e != nil {
			return nil, e
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		_, err := c.set(key, v, expiration)
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

// evict removes the oldest item from the cache.
func (c *LRU) evict(count int) {
	for i := 0; i < count; i++ {
		ent := c.evictList.Back()
		if ent == nil {
			return
		}

		c.removeElement(ent)
	}
}

// Has checks if key exists in cache
func (c *LRU) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	return c.has(key, &now)
}

func (c *LRU) has(key any, now *time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.Value.(*lruItem).Expired(now)
}

// Remove removes the provided key from the cache.
func (c *LRU) Remove(key any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *LRU) remove(key any) bool {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

func (c *LRU) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	entry := e.Value.(*lruItem)
	delete(c.items, entry.key)
	if c.evictedFunc != nil {
		entry := e.Value.(*lruItem)
		c.evictedFunc(entry.key, entry.value)
	}
}

func (c *LRU) keys() []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]any, len(c.items))
	var i = 0
	for k := range c.items {
		keys[i] = k
		i++
	}
	return keys
}

// GetALL returns all key-value pairs in the cache.
func (c *LRU) GetALL(checkExpired bool) map[any]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make(map[any]any, len(c.items))
	now := time.Now()
	for k, item := range c.items {
		if !checkExpired || !item.Value.(*lruItem).Expired(&now) {
			items[k] = item.Value.(*lruItem).value
		}
	}
	return items
}

// Keys returns a slice of the keys in the cache.
func (c *LRU) Keys(checkExpired bool) []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]any, 0, len(c.items))
	now := time.Now()
	for k, item := range c.items {
		if !checkExpired || !item.Value.(*lruItem).Expired(&now) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRU) Len(checkExpired bool) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !checkExpired {
		return len(c.items)
	}
	var length int
	now := time.Now()
	for _, item := range c.items {
		if !item.Value.(*lruItem).Expired(&now) {
			length++
		}
	}
	return length
}

// Purge all expired items from the Simple.
func (c *LRU) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	if c.purgeVisitorFunc != nil {
		for key, item := range c.items {
			it := item.Value.(*lruItem)
			v := it.value
			if it.Expired(&now) {
				c.removeElement(item)
				c.purgeVisitorFunc(key, v)
			}

		}
	}
}

// Flush is used to completely clear the cache
func (c *LRU) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.init()
}

type lruItem struct {
	key        any
	value      any
	expiration *time.Time
}

// Expired returns boolean value whether this item is expired or not.
func (it *lruItem) Expired(now *time.Time) bool {
	if it.expiration == nil {
		return false
	}
	if now == nil {
		return it.expiration.Before(time.Now())
	}
	return it.expiration.Before(*now)
}
