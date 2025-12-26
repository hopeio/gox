package cache

import (
	"time"
)

type Simple struct {
	*baseCache
	items map[any]*item

	// If this is confusing, see the comment at the bottom of New()
}

func (c *Simple) init(bc *baseCache) {
	c.baseCache = bc
	c.items = make(map[any]*item, bc.size)
}

func (c *Simple) set(k any, x any, expiration *time.Time) (*item, error) {
	if (len(c.items)) >= c.size {
		c.exvict(1)
	}
	item := &item{
		key:        k,
		value:      x,
		expiration: expiration,
	}
	c.items[k] = item
	if c.addedFunc != nil {
		c.addedFunc(k, x)
	}
	return item, nil
}

func (c *Simple) get(k any, onLoad bool) (*item, error) {
	item, found := c.items[k]
	if found {
		if !item.Expired(nil) {
			if !onLoad {
				c.stats.IncrMissCount()
			}
			return item, nil
		}
		delete(c.items, k)
		if c.evictedFunc != nil {
			c.evictedFunc(item.key, item.value)
		}
	}
	if !onLoad {
		c.stats.IncrMissCount()
	}
	return nil, KeyNotFoundError
}

func (c *Simple) length() int {
	return len(c.items)
}

func (c *Simple) foreach(f func(*item)) {
	for _, item := range c.items {
		f(item)
	}
}

func (c *Simple) remove(k any) bool {
	if it, found := c.items[k]; found {
		delete(c.items, it.key)
		if c.evictedFunc != nil {
			c.evictedFunc(it.key, it.value)
		}
		return true
	}
	return false
}

func (c *Simple) exvict(count int) {
	for k, item := range c.items {
		if item.Expired(nil) {
			count--
			delete(c.items, k)
			if c.evictedFunc != nil {
				c.evictedFunc(item.key, item.value)
			}
			if count <= 0 {
				break
			}
		}
	}
	if count > 0 {
		for k, item := range c.items {
			count--
			delete(c.items, k)
			if c.evictedFunc != nil {
				c.evictedFunc(item.key, item.value)
			}
			if count <= 0 {
				break
			}
		}
	}
}
