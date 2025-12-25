package cache

import (
	"container/list"
	"time"
)

// Discards the least recently used items first.
type LRU struct {
	size int
	items     map[any]*list.Element
	evictList *list.List
}

func (c *LRU) init(size int) {
	c.size = size
	c.evictList = list.New()
	c.items = make(map[any]*list.Element, size+1)
}

func (c *LRU) set(key, value any, expiration *time.Time) error {
	// Check for existing item
	var ite *item
	if it, ok := c.items[key]; ok {
		c.evictList.MoveToFront(it)
		ite = it.Value.(*item)
		ite.value = value
	} else {
		// Verify size not exceeded
		if c.evictList.Len() >= c.size {
			c.evict(1)
		}
		ite = &item{
			key:   key,
			value: value,
		}
		c.items[key] = c.evictList.PushFront(ite)
	}

	ite.expiration = expiration

	return nil
}

func (c *LRU) get(key any) (any, error) {
	ite, ok := c.items[key]
	if ok {
		it := ite.Value.(*item)
		if !it.Expired(nil) {
			c.evictList.MoveToFront(ite)
			v := it.value

			return v, nil
		}
		c.removeElement(ite)
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
	ite, ok := c.items[key]
	if !ok {
		return false
	}
	return !ite.Value.(*item).Expired(now)
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
