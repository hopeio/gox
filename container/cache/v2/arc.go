package cache

import (
	"container/list"
	"time"
)

// Constantly balances between LRU and LFU, to improve the combined result.
type ARC struct {
	size int
	items map[any]*item

	part int
	t1   *arcList
	t2   *arcList
	b1   *arcList
	b2   *arcList
}


func (c *ARC) init(size int) {
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
	_, ok := c.items[old]
	if ok {
		delete(c.items, old)
	}
}


func (c *ARC) set(key, value any, expiration *time.Time)  error {
	if c.t1.Has(key) || c.t2.Has(key) {
		return  KeyAlreadyExistError
	}

	it, ok := c.items[key]
	if ok {
		it.value = value
	} else {
		it = &item{
			value: value,
		}
		c.items[key] = it
	}

	it.expiration = expiration

	if elt := c.b1.Lookup(key); elt != nil {
		c.setPart(min(c.size, c.part+max(c.b2.Len()/c.b1.Len(), 1)))
		c.replace(key)
		c.b1.Remove(key, elt)
		c.t2.PushFront(key)
		return  nil
	}

	if elt := c.b2.Lookup(key); elt != nil {
		c.setPart(max(0, c.part-max(c.b1.Len()/c.b2.Len(), 1)))
		c.replace(key)
		c.b2.Remove(key, elt)
		c.t2.PushFront(key)
		return  nil
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
	return  nil
}



func (c *ARC) get(key any,) (any, error) {
	if elt := c.t1.Lookup(key); elt != nil {
		c.t1.Remove(key, elt)
		item := c.items[key]
		if !item.Expired(nil) {
			c.t2.PushFront(key)
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
			return item.value, nil
		}

		delete(c.items, key)
		c.t2.Remove(key, elt)
		c.b2.PushFront(key)
		if c.evictedFunc != nil {
			c.evictedFunc(key, item.value)
		}
	}

	return nil, KeyNotFoundError
}


func (c *ARC) has(key any, now *time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.Expired(now)
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


// Len returns the number of items in the cache.
func (c *ARC) length() int {
	return len(c.items)
}

func (c *ARC) foreach(f func(*item) ) {
	for _, v := range c.items {
		f(v)
	}
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
