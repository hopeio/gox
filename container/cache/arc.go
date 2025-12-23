package cache

import (
	"container/list"
	"runtime"
	"time"
)

// Constantly balances between LRU and LFU, to improve the combined result.
type ARC struct {
	baseCache
	items map[any]*item

	part int
	t1   *arcList
	t2   *arcList
	b1   *arcList
	b2   *arcList
}

func newARCCache(cb *CacheBuilder) *ARC {
	c := &ARC{}
	buildCache(&c.baseCache, cb)

	c.init()
	c.loadGroup.cache = c
	if c.janitor != nil {
		go c.janitor.Run(c)
		runtime.SetFinalizer(c, stopJanitor)
	}
	return c
}

func (c *ARC) init() {
	c.items = make(map[any]*item)
	c.t1 = newARCList()
	c.t2 = newARCList()
	c.b1 = newARCList()
	c.b2 = newARCList()
}

func (c *ARC) replace(key any) {
	if !c.isCacheFull() {
		return
	}
	var old any
	if c.t1.Len() > 0 && ((c.b2.Has(key) && c.t1.Len() == c.part) || (c.t1.Len() > c.part)) {
		old = c.t1.RemoveTail()
		c.b1.PushFront(old)
	} else if c.t2.Len() > 0 {
		old = c.t2.RemoveTail()
		c.b2.PushFront(old)
	} else {
		old = c.t1.RemoveTail()
		c.b1.PushFront(old)
	}
	item, ok := c.items[old]
	if ok {
		delete(c.items, old)
		if c.evictedFunc != nil {
			c.evictedFunc(old, item.value)
		}
	}
}

// Set a new key-value pair with an expiration time
func (c *ARC) Set(key, value any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.set(key, value, expiration)
	return err
}

// SetNX an item to the Simple only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *ARC) SetNX(k any, x any, d time.Duration) error {
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

func (c *ARC) set(key, value any, expiration time.Duration) (*item, error) {
	it, ok := c.items[key]
	if ok {
		it.value = value
	} else {
		it = &item{
			value: value,
		}
		c.items[key] = it
	}

	if expiration == DefaultExpiration && c.expiration > 0 {
		t := time.Now().Add(c.expiration)
		it.expiration = &t
	} else if expiration > 0 {
		t := time.Now().Add(expiration)
		it.expiration = &t
	}

	defer func() {
		if c.addedFunc != nil {
			c.addedFunc(key, value)
		}
	}()

	if c.t1.Has(key) || c.t2.Has(key) {
		return it, nil
	}

	if elt := c.b1.Lookup(key); elt != nil {
		c.setPart(min(c.size, c.part+max(c.b2.Len()/c.b1.Len(), 1)))
		c.replace(key)
		c.b1.Remove(key, elt)
		c.t2.PushFront(key)
		return it, nil
	}

	if elt := c.b2.Lookup(key); elt != nil {
		c.setPart(max(0, c.part-max(c.b1.Len()/c.b2.Len(), 1)))
		c.replace(key)
		c.b2.Remove(key, elt)
		c.t2.PushFront(key)
		return it, nil
	}

	if c.isCacheFull() && c.t1.Len()+c.b1.Len() == c.size {
		if c.t1.Len() < c.size {
			c.b1.RemoveTail()
			c.replace(key)
		} else {
			pop := c.t1.RemoveTail()
			item, ok := c.items[pop]
			if ok {
				delete(c.items, pop)
				if c.evictedFunc != nil {
					c.evictedFunc(pop, item.value)
				}
			}
		}
	} else {
		total := c.t1.Len() + c.b1.Len() + c.t2.Len() + c.b2.Len()
		if total >= c.size {
			if total == (2 * c.size) {
				if c.b2.Len() > 0 {
					c.b2.RemoveTail()
				} else {
					c.b1.RemoveTail()
				}
			}
			c.replace(key)
		}
	}
	c.t1.PushFront(key)
	return it, nil
}

// Get a value from cache pool using key if it exists. If not exists and it has LoaderFunc, it will generate the value using you have specified LoaderFunc method returns value.
func (c *ARC) Get(key any) (any, error) {
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
func (c *ARC) GetIFPresent(key any) (any, error) {
	c.mu.Lock()
	v, err := c.get(key, false)
	c.mu.Unlock()
	if err == KeyNotFoundError {
		return c.getWithLoader(key, false)
	}
	return v, err
}

func (c *ARC) get(key any, onLoad bool) (any, error) {
	if elt := c.t1.Lookup(key); elt != nil {
		c.t1.Remove(key, elt)
		item := c.items[key]
		if !item.Expired(nil) {
			c.t2.PushFront(key)
			if !onLoad {
				c.stats.IncrHitCount()
			}
			return item.value, nil
		}

		delete(c.items, key)
		c.b1.PushFront(key)
		if c.evictedFunc != nil {
			c.evictedFunc(key, item.value)
		}
	}
	if elt := c.t2.Lookup(key); elt != nil {
		item := c.items[key]
		if !item.Expired(nil) {
			c.t2.MoveToFront(elt)
			if !onLoad {
				c.stats.IncrHitCount()
			}
			return item.value, nil
		}

		delete(c.items, key)
		c.t2.Remove(key, elt)
		c.b2.PushFront(key)
		if c.evictedFunc != nil {
			c.evictedFunc(key, item.value)
		}
	}

	if !onLoad {
		c.stats.IncrMissCount()
	}
	return nil, KeyNotFoundError
}

func (c *ARC) getWithLoader(key any, isWait bool) (any, error) {
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

// Has checks if key exists in cache
func (c *ARC) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	return c.has(key, &now)
}

func (c *ARC) has(key any, now *time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.Expired(now)
}

// Remove removes the provided key from the cache.
func (c *ARC) Remove(key any) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *ARC) remove(key any) bool {
	if elt := c.t1.Lookup(key); elt != nil {
		c.t1.Remove(key, elt)
		item := c.items[key]
		delete(c.items, key)
		c.b1.PushFront(key)
		if c.evictedFunc != nil {
			c.evictedFunc(key, item.value)
		}
		return true
	}

	if elt := c.t2.Lookup(key); elt != nil {
		c.t2.Remove(key, elt)
		item := c.items[key]
		delete(c.items, key)
		c.b2.PushFront(key)
		if c.evictedFunc != nil {
			c.evictedFunc(key, item.value)
		}
		return true
	}

	return false
}

// GetALL returns all key-value pairs in the cache.
func (c *ARC) GetALL(checkExpired bool) map[any]any {
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
func (c *ARC) Keys(checkExpired bool) []any {
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
func (c *ARC) Len(checkExpired bool) int {
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

// Purge is used to completely clear the cache
func (c *ARC) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	if c.purgeVisitorFunc != nil {
		for key, item := range c.items {
			if item.Expired(&now) {
				c.remove(key)
				c.purgeVisitorFunc(key, item.value)
			}
		}
	}
}

// Flush is used to completely clear the cache
func (c *ARC) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.init()
}

func (c *ARC) setPart(p int) {
	if c.isCacheFull() {
		c.part = p
	}
}

func (c *ARC) isCacheFull() bool {
	return (c.t1.Len() + c.t2.Len()) == c.size
}

type arcList struct {
	l    *list.List
	keys map[any]*list.Element
}

func newARCList() *arcList {
	return &arcList{
		l:    list.New(),
		keys: make(map[any]*list.Element),
	}
}

func (al *arcList) Has(key any) bool {
	_, ok := al.keys[key]
	return ok
}

func (al *arcList) Lookup(key any) *list.Element {
	elt := al.keys[key]
	return elt
}

func (al *arcList) MoveToFront(elt *list.Element) {
	al.l.MoveToFront(elt)
}

func (al *arcList) PushFront(key any) {
	if elt, ok := al.keys[key]; ok {
		al.l.MoveToFront(elt)
		return
	}
	elt := al.l.PushFront(key)
	al.keys[key] = elt
}

func (al *arcList) Remove(key any, elt *list.Element) {
	delete(al.keys, key)
	al.l.Remove(elt)
}

func (al *arcList) RemoveTail() any {
	elt := al.l.Back()
	if elt == nil {
		return nil
	}
	al.l.Remove(elt)

	key := elt.Value
	delete(al.keys, key)

	return key
}

func (al *arcList) Len() int {
	return al.l.Len()
}
