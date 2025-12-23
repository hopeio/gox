package cache

import (
	"container/list"
	"runtime"
	"time"
)

// Discards the least frequently used items first.
type LFU struct {
	baseCache
	items    map[any]*lfuItem
	freqList *list.List // list for freqEntry
}

func newLFUCache(cb *CacheBuilder) *LFU {
	c := &LFU{}
	buildCache(&c.baseCache, cb)

	c.init()
	c.loadGroup.cache = c
	if c.janitor != nil {
		go c.janitor.Run(c)
		runtime.SetFinalizer(c, stopJanitor)
	}
	return c
}

func (c *LFU) init() {
	c.freqList = list.New()
	c.items = make(map[any]*lfuItem, c.size+1)
	c.freqList.PushFront(&freqEntry{
		freq:  0,
		items: make(map[*lfuItem]struct{}),
	})
}

// Set a new key-value pair with an expiration time
func (c *LFU) Set(key, value any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.set(key, value, expiration)
	return err
}

// SetNX an item to the Simple only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *LFU) SetNX(k any, x any, d time.Duration) error {
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

func (c *LFU) set(key, value any, expiration time.Duration) (*lfuItem, error) {
	// Check for existing item
	item, ok := c.items[key]
	if ok {
		item.value = value
	} else {
		// Verify size not exceeded
		if len(c.items) >= c.size {
			c.evict(1)
		}
		item = &lfuItem{
			key:         key,
			value:       value,
			freqElement: nil,
		}
		el := c.freqList.Front()
		fe := el.Value.(*freqEntry)
		fe.items[item] = struct{}{}

		item.freqElement = el
		c.items[key] = item
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

// Get a value from cache pool using key if it exists.
// If it dose not exists key and has LoaderFunc,
// generate a value using `LoaderFunc` method returns value.
func (c *LFU) Get(key any) (any, error) {
	c.mu.Lock()
	v, err := c.get(key, false)
	c.mu.Unlock()
	if err == KeyNotFoundError {
		return c.getWithLoader(key, true)
	}
	return v, err
}

// GetIFPresent gets a value from cache pool using key if it exists.
// If it dose not exists key, returns KeyNotFoundError.
// And send a request which refresh value for specified key if cache object has LoaderFunc.
func (c *LFU) GetIFPresent(key any) (any, error) {
	c.mu.Lock()
	v, err := c.get(key, false)
	c.mu.Unlock()
	if err == KeyNotFoundError {
		return c.getWithLoader(key, false)
	}
	return v, err
}

func (c *LFU) get(key any, onLoad bool) (any, error) {
	item, ok := c.items[key]
	if ok {
		if !item.Expired(nil) {
			c.increment(item)
			v := item.value
			if !onLoad {
				c.stats.IncrHitCount()
			}
			return v, nil
		}
		c.removeItem(item)
	}
	if !onLoad {
		c.stats.IncrMissCount()
	}
	return nil, KeyNotFoundError
}

func (c *LFU) getWithLoader(key any, isWait bool) (any, error) {
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

func (c *LFU) increment(item *lfuItem) {
	currentFreqElement := item.freqElement
	currentFreqEntry := currentFreqElement.Value.(*freqEntry)
	nextFreq := currentFreqEntry.freq + 1
	delete(currentFreqEntry.items, item)

	nextFreqElement := currentFreqElement.Next()
	if nextFreqElement == nil {
		nextFreqElement = c.freqList.InsertAfter(&freqEntry{
			freq:  nextFreq,
			items: make(map[*lfuItem]struct{}),
		}, currentFreqElement)
	}
	nextFreqElement.Value.(*freqEntry).items[item] = struct{}{}
	item.freqElement = nextFreqElement
}

// evict removes the least frequence item from the cache.
func (c *LFU) evict(count int) {
	entry := c.freqList.Front()
	for i := 0; i < count; {
		if entry == nil {
			return
		} else {
			for item := range entry.Value.(*freqEntry).items {
				if i >= count {
					return
				}
				c.removeItem(item)
				i++
			}
			entry = entry.Next()
		}
	}
}

// Has checks if key exists in cache
func (c *LFU) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	return c.has(key, &now)
}

func (c *LFU) has(key any, now *time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.Expired(now)
}

// Remove removes the provided key from the cache.
func (c *LFU) Remove(key any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *LFU) remove(key any) bool {
	if item, ok := c.items[key]; ok {
		c.removeItem(item)
		return true
	}
	return false
}

// removeElement is used to remove a given list element from the cache
func (c *LFU) removeItem(item *lfuItem) {
	delete(c.items, item.key)
	delete(item.freqElement.Value.(*freqEntry).items, item)
	if c.evictedFunc != nil {
		c.evictedFunc(item.key, item.value)
	}
}

func (c *LFU) keys() []any {
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
func (c *LFU) GetALL(checkExpired bool) map[any]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make(map[any]any, len(c.items))
	now := time.Now()
	for k, item := range c.items {
		if !checkExpired || !item.Expired(&now) {
			items[k] = item.value
		}
	}
	return items
}

// Keys returns a slice of the keys in the cache.
func (c *LFU) Keys(checkExpired bool) []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]any, 0, len(c.items))
	now := time.Now()
	for k, item := range c.items {
		if !checkExpired || !item.Expired(&now) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LFU) Len(checkExpired bool) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !checkExpired {
		return len(c.items)
	}
	var length int
	now := time.Now()
	for _, item := range c.items {
		if !item.Expired(&now) {
			length++
		}
	}
	return length
}

// Completely clear the cache
func (c *LFU) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	if c.purgeVisitorFunc != nil {
		for key, item := range c.items {
			if item.Expired(&now) {
				c.removeItem(item)
				c.purgeVisitorFunc(key, item.value)
			}
		}
	}
}

// Flush is used to completely clear the cache
func (c *LFU) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.init()
}

type freqEntry struct {
	freq  uint
	items map[*lfuItem]struct{}
}

type lfuItem struct {
	key         any
	value       any
	freqElement *list.Element
	expiration  *time.Time
}

// Expired returns boolean value whether this item is expired or not.
func (it *lfuItem) Expired(now *time.Time) bool {
	if it.expiration == nil {
		return false
	}
	if now == nil {
		return it.expiration.Before(time.Now())
	}
	return it.expiration.Before(*now)
}
