package cache

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

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

type Simple struct {
	baseCache
	items map[any]item

	// If this is confusing, see the comment at the bottom of New()
}

// SetNX an item to the Simple, replacing any existing item. If the duration is 0
// (DefaultExpiration), the Simple's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *Simple) Set(k any, x any, d time.Duration) error {
	c.mu.Lock()
	c.set(k, x, d)
	c.mu.Unlock()
	return nil
}

func (c *Simple) set(k any, x any, expiration time.Duration) (*item, error) {
	var e *time.Time
	if expiration > 0 {
		now := time.Now().Add(expiration)
		e = &now
	} else if expiration == DefaultExpiration && c.expiration > 0 {
		t := time.Now().Add(c.expiration)
		e = &t
	}
	item := item{
		value:      x,
		expiration: e,
	}
	c.items[k] = item
	if c.addedFunc != nil {
		c.addedFunc(k, x)
	}
	return &item, nil
}

// SetNX an item to the Simple only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *Simple) SetNX(k any, x any, d time.Duration) error {
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

// Set a new value for the Simple key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (c *Simple) Replace(k any, x any, d time.Duration) error {
	c.mu.Lock()
	_, err := c.get(k, true)
	if err != nil {
		c.mu.Unlock()
		return KeyNotFoundError
	}
	c.set(k, x, d)
	c.mu.Unlock()
	return nil
}

// Get an item from the Simple. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *Simple) Get(k any) (any, error) {
	c.mu.RLock()
	item, found := c.items[k]
	if !found || item.Expired(nil) {
		c.mu.RUnlock()
		return nil, KeyNotFoundError
	}
	c.mu.RUnlock()
	return item.value, KeyNotFoundError
}

// GetWithExpiration returns an item and its expiration time from the Simple.
// It returns the item or nil, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (c *Simple) GetWithExpiration(k any) (any, time.Time, error) {
	c.mu.RLock()
	// "Inlining" of get and Expired
	item, found := c.items[k]
	if !found {
		c.mu.RUnlock()
		return nil, time.Time{}, KeyNotFoundError
	}

	if item.expiration != nil {
		if item.expiration.Before(time.Now()) {
			c.mu.RUnlock()
			return nil, time.Time{}, KeyNotFoundError
		}

		// Return the item and the expiration time
		c.mu.RUnlock()
		return item.value, *item.expiration, KeyNotFoundError
	}

	// If expiration <= 0 (i.e. no expiration time set) then return the item
	// and a zeroed time.Time
	c.mu.RUnlock()
	return item.value, time.Time{}, nil
}

func (c *Simple) get(k any, onLoad bool) (any, error) {
	item, found := c.items[k]
	if !found || item.Expired(nil) {
		if !onLoad {
			c.stats.IncrMissCount()
		}
		return nil, KeyNotFoundError
	}
	if !onLoad {
		c.stats.IncrHitCount()
	}
	return item.value, nil
}

func (c *Simple) getWithLoader(key any, isWait bool) (any, error) {
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

func (c *Simple) getItem(k any) (*item, error) {
	item, found := c.items[k]
	if !found {
		return nil, KeyNotFoundError
	}
	// "Inlining" of Expired
	if item.Expired(nil) {
		return nil, KeyNotFoundError
	}
	return &item, nil
}

// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n. To retrieve the incremented value, use one
// of the specialized methods, e.g. IncrementInt64.
func (c *Simple) Increment(k any, n int64) error {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nil
}

// Increment an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to increment it by n. Pass a negative number to decrement the
// value. To retrieve the incremented value, use one of the specialized methods,
// e.g. IncrementFloat64.
func (c *Simple) IncrementFloat(k any, n float64) error {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nil
}

// Increment an item of type int by n. Returns an error if the item's value is
// not an int, or if it was not found. If there is no error, the incremented
// value is returned.
func (c *Simple) IncrementInt(k any, n int) (int, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type int32 by n. Returns an error if the item's value is
// not an int32, or if it was not found. If there is no error, the incremented
// value is returned.
func (c *Simple) IncrementInt32(k any, n int32) (int32, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type int64 by n. Returns an error if the item's value is
// not an int64, or if it was not found. If there is no error, the incremented
// value is returned.
func (c *Simple) IncrementInt64(k any, n int64) (int64, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint by n. Returns an error if the item's value is
// not an uint, or if it was not found. If there is no error, the incremented
// value is returned.
func (c *Simple) IncrementUint(k any, n uint) (uint, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uintptr by n. Returns an error if the item's value
// is not an uintptr, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Simple) IncrementUintptr(k any, n uintptr) (uintptr, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint32 by n. Returns an error if the item's value
// is not an uint32, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Simple) IncrementUint32(k any, n uint32) (uint32, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint64 by n. Returns an error if the item's value
// is not an uint64, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Simple) IncrementUint64(k any, n uint64) (uint64, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type float32 by n. Returns an error if the item's value
// is not an float32, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Simple) IncrementFloat32(k any, n float32) (float32, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type float64 by n. Returns an error if the item's value
// is not an float64, or if it was not found. If there is no error, the
// incremented value is returned.
func (c *Simple) IncrementFloat64(k any, n float64) (float64, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n. To retrieve the decremented value, use one
// of the specialized methods, e.g. DecrementInt64.
func (c *Simple) Decrement(k any, n int64) error {
	// TODO: Implement Increment and Decrement more cleanly.
	// (Cannot do Increment(k, n*-1) for uints.)
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nil
}

// Decrement an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to decrement it by n. Pass a negative number to decrement the
// value. To retrieve the decremented value, use one of the specialized methods,
// e.g. DecrementFloat64.
func (c *Simple) DecrementFloat(k any, n float64) error {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nil
}

// Decrement an item of type int by n. Returns an error if the item's value is
// not an int, or if it was not found. If there is no error, the decremented
// value is returned.
func (c *Simple) DecrementInt(k any, n int) (int, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int32 by n. Returns an error if the item's value is
// not an int32, or if it was not found. If there is no error, the decremented
// value is returned.
func (c *Simple) DecrementInt32(k any, n int32) (int32, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int64 by n. Returns an error if the item's value is
// not an int64, or if it was not found. If there is no error, the decremented
// value is returned.
func (c *Simple) DecrementInt64(k any, n int64) (int64, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint by n. Returns an error if the item's value is
// not an uint, or if it was not found. If there is no error, the decremented
// value is returned.
func (c *Simple) DecrementUint(k any, n uint) (uint, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uintptr by n. Returns an error if the item's value
// is not an uintptr, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Simple) DecrementUintptr(k any, n uintptr) (uintptr, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint32 by n. Returns an error if the item's value
// is not an uint32, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Simple) DecrementUint32(k any, n uint32) (uint32, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint64 by n. Returns an error if the item's value
// is not an uint64, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Simple) DecrementUint64(k any, n uint64) (uint64, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type float32 by n. Returns an error if the item's value
// is not an float32, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Simple) DecrementFloat32(k any, n float32) (float32, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type float64 by n. Returns an error if the item's value
// is not an float64, or if it was not found. If there is no error, the
// decremented value is returned.
func (c *Simple) DecrementFloat64(k any, n float64) (float64, error) {
	c.mu.Lock()
	v, err := c.getItem(k)
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
	c.items[k] = *v
	c.mu.Unlock()
	return nv, nil
}

// Remove an item from the Simple. Does nothing if the key is not in the Simple.
func (c *Simple) Remove(k any) bool {
	c.mu.Lock()
	v, evicted := c.remove(k)
	c.mu.Unlock()
	if evicted {
		c.evictedFunc(k, v)
	}
	return true
}

func (c *Simple) remove(k any) (any, bool) {
	if c.evictedFunc != nil {
		if v, found := c.items[k]; found {
			delete(c.items, k)
			return v.value, true
		}
	}
	delete(c.items, k)
	return nil, false
}

type keyAndValue struct {
	key   any
	value any
}

// Purge all expired items from the Simple.
func (c *Simple) Purge() {
	var evictedItems []keyAndValue
	now := time.Now()
	c.mu.Lock()
	for k, v := range c.items {
		if c.purgeVisitorFunc != nil {
			c.purgeVisitorFunc(k, v.value)
		}

		if v.Expired(&now) {
			ov, evicted := c.remove(k)
			if evicted {
				evictedItems = append(evictedItems, keyAndValue{k, ov})
			}
		}
	}
	c.mu.Unlock()
	for _, v := range evictedItems {
		c.evictedFunc(v.key, v.value)
	}
}

// Sets an (optional) function that is called with the key and value when an
// item is evicted from the Simple. (Including when it is deleted manually, but
// not when it is overwritten.) Set to nil to disable.
func (c *Simple) OnEvicted(f func(any, any)) {
	c.mu.Lock()
	c.evictedFunc = f
	c.mu.Unlock()
}

// Write the Simple's items (using Gob) to an io.Writer.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *Simple) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("error registering item types with Gob library")
		}
	}()
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.items {
		gob.Register(v.value)
	}
	err = enc.Encode(&c.items)
	return
}

// Save the Simple's items to the given filename, creating the file if it
// doesn't exist, and overwriting it if it does.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *Simple) SaveFile(fname string) error {
	fp, err := os.Create(fname)
	if err != nil {
		return err
	}
	err = c.Save(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// SetNX (Gob-serialized) Simple items from an io.Reader, excluding any items with
// keys that already exist (and haven't expired) in the current Simple.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *Simple) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]item{}
	err := dec.Decode(&items)
	if err == nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		for k, v := range items {
			ov, found := c.items[k]
			if !found || ov.Expired(nil) {
				c.items[k] = v
			}
		}
	}
	return err
}

// Load and add Simple items from the given filename, excluding any items with
// keys that already exist in the current Simple.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *Simple) LoadFile(fname string) error {
	fp, err := os.Open(fname)
	if err != nil {
		return err
	}
	err = c.Load(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// Copies all unexpired items in the Simple into a new map and returns it.
func (c *Simple) Items() map[any]item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := make(map[any]item, len(c.items))
	now := time.Now()
	for k, v := range c.items {
		if v.Expired(&now) {
			continue
		}
		m[k] = v
	}
	return m
}

// Remove all items from the Simple.
func (c *Simple) Flush() {
	c.mu.Lock()
	c.items = map[any]item{}
	c.mu.Unlock()
}

// Has checks if key exists in Simple
func (c *Simple) Has(key any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.has(key, time.Now())
}

func (c *Simple) has(key any, now time.Time) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.Expired(&now)
}

// Returns a slice of the keys in the Simple.
func (c *Simple) keys(checkExpired bool) []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]any, 0, len(c.items))
	now := time.Now()
	for k, item := range c.items {
		if checkExpired && item.Expired(&now) {
			continue
		}
		keys = append(keys, k)
	}
	return keys
}

// GetALL returns all key-value pairs in the Simple.
func (c *Simple) GetALL(checkExpired bool) map[any]any {
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

// Keys returns a slice of the keys in the Simple.
func (c *Simple) Keys(checkExpired bool) []any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]interface{}, 0, len(c.items))
	now := time.Now()
	for k, item := range c.items {
		if !checkExpired || !item.Expired(&now) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Len returns the number of items in the Simple.
func (c *Simple) Len(checkExpired bool) int {
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

func newSimpleCache(cb *CacheBuilder) *Simple {
	c := &Simple{items: make(map[any]item, cb.size)}
	buildCache(&c.baseCache, cb)

	c.loadGroup.cache = c
	if c.janitor != nil {
		go c.janitor.Run(c)
		runtime.SetFinalizer(c, stopJanitor)
	}
	return c
}
