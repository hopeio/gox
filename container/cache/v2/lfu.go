package cache

import (
	"container/list"
	"time"
)

// Discards the least frequently used items first.
type LFU struct {
	size     int
	items    map[any]*lfuItem
	freqList *list.List // list for freqEntry
}

func (c *LFU) init(size int) {
	c.size = size
	c.freqList = list.New()
	c.items = make(map[any]*lfuItem, size+1)
	c.freqList.PushFront(&freqEntry{
		freq:  0,
		items: make(map[*lfuItem]struct{}),
	})
}

func (c *LFU) set(key, value any, expiration *time.Time) error {
	// Check for existing item
	ite, ok := c.items[key]
	if ok {
		ite.value = value
	} else {
		// Verify size not exceeded
		if len(c.items) >= c.size {
			c.evict(1)
		}
		ite = &lfuItem{
			item: item{
				key:   key,
				value: value},
			freqElement: nil,
		}
		el := c.freqList.Front()
		fe := el.Value.(*freqEntry)
		fe.items[ite] = struct{}{}

		ite.freqElement = el
		c.items[key] = ite
	}

	ite.expiration = expiration

	return nil
}

func (c *LFU) get(key any) (any, error) {
	item, ok := c.items[key]
	if ok {
		if !item.Expired(nil) {
			c.increment(item)
			v := item.value
			return v, nil
		}
		c.removeItem(item)
	}
	return nil, KeyNotFoundError
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

func (c *LFU) has(key any, now *time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.Expired(now)
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
}

func (c *LFU) length() int {
	return len(c.items)
}

func (c *LFU) foreach(f func(*item)) {
	for _, item := range c.items {
		f(&item.item)
	}
}

type freqEntry struct {
	freq  uint
	items map[*lfuItem]struct{}
}

type lfuItem struct {
	item
	freqElement *list.Element
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
