package cache

import (
	"container/list"
	"time"
)

// Discards the least recently used items first.
type LRU struct {
	*baseCache
	items     map[any]*list.Element
	evictList *list.List
}

func (c *LRU) init(bc *baseCache) {
	c.baseCache = bc
	c.evictList = list.New()
	c.items = make(map[any]*list.Element, c.size+1)
}

func (c *LRU) set(key, value any, expiration *time.Time) (*item, error) {
	// Check for existing item
	var it *item
	if ite, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ite)
		it = ite.Value.(*item)
		it.value = value
	} else {
		// Verify size not exceeded
		if c.evictList.Len() >= c.size {
			c.evict(1)
		}
		it = &item{
			key:   key,
			value: value,
		}
		c.items[key] = c.evictList.PushFront(it)
	}

	it.expiration = expiration

	if c.addedFunc != nil {
		c.addedFunc(key, value)
	}

	return it, nil
}

func (c *LRU) get(key any, onLoad bool) (*item, error) {
	ite, ok := c.items[key]
	if ok {
		it := ite.Value.(*item)
		if !it.Expired(nil) {
			c.evictList.MoveToFront(ite)

			if !onLoad {
				c.stats.IncrHitCount()
			}
			return it, nil
		}
		c.removeElement(ite)
	}
	if !onLoad {
		c.stats.IncrMissCount()
	}
	return nil, KeyNotFoundError
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

func (c *LRU) has(key any, now *time.Time) bool {
	it, ok := c.items[key]
	if !ok {
		return false
	}
	return !it.Value.(*item).Expired(now)
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
	entry := e.Value.(*item)
	delete(c.items, entry.key)
	if c.evictedFunc != nil {
		entry := e.Value.(*item)
		c.evictedFunc(entry.key, entry.value)
	}
}

func (c *LRU) length() int {
	return len(c.items)
}

func (c *LRU) foreach(f func(*item)) {
	for _, e := range c.items {
		f(e.Value.(*item))
	}
}
