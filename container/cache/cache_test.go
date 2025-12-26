package cache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func getCaches(builder *CacheBuilder) []*Cache {
	return []*Cache{
		builder.LRU(),
		builder.LFU(),
		builder.ARC(),
		builder.Simple(),
	}
}

func getCache(i int, builder *CacheBuilder) *Cache {
	switch i {
	case 0:
		return builder.LRU()
	case 1:
		return builder.LFU()
	case 2:
		return builder.ARC()
	case 3:
		return builder.Simple()
	}
	return nil
}

func getName(i int) string {
	switch i {
	case 0:
		return "LRU"
	case 1:
		return "LFU"
	case 2:
		return "ARC"
	case 3:
		return "Simple"
	}
	return ""
}

func TestLoaderFunc(t *testing.T) {

	for i := range 4 {
		var testCounter int64
		counter := 1000
		builder := New(2).
			LoaderFunc(func(key interface{}) (interface{}, time.Duration, error) {
				time.Sleep(10 * time.Millisecond)
				return atomic.AddInt64(&testCounter, 1), 0, nil
			}).
			EvictedFunc(func(key, value interface{}) {
				panic(key)
			})

		cache := getCache(i, builder)
		var wg sync.WaitGroup
		for i := 0; i < counter; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(0)
				if err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()

		if testCounter != 1 {
			t.Errorf("testCounter != %v", testCounter)
		}

	}
}

func TestLoaderExpireFuncWithoutExpire(t *testing.T) {
	for i := range 4 {
		var testCounter int64
		counter := 1000
		builder := New(2).
			LoaderFunc(func(key interface{}) (interface{}, time.Duration, error) {
				return atomic.AddInt64(&testCounter, 1), NoExpiration, nil
			}).
			EvictedFunc(func(key, value interface{}) {
				panic(key)
			})
		cache := getCache(i, builder)

		var wg sync.WaitGroup
		for i := 0; i < counter; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(0)
				if err != nil {
					t.Error(err)
				}
			}()
		}

		wg.Wait()

		if testCounter != 1 {
			t.Errorf("testCounter != %v", testCounter)
		}
	}
}

func TestLoaderExpireFuncWithExpire(t *testing.T) {

	for i := range 4 {
		var testCounter int64
		counter := 1000
		expire := 200 * time.Millisecond
		builder := New(2).
			LoaderFunc(func(key interface{}) (any, time.Duration, error) {
				return atomic.AddInt64(&testCounter, 1), expire, nil
			})
		cache := getCache(i, builder)
		var wg sync.WaitGroup
		for i := 0; i < counter; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(0)
				if err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()

		if testCounter != 1 {
			t.Errorf("%s: testCounter %v != 1", getName(i), testCounter)
		}
		time.Sleep(expire) // Waiting for key expiration

		for i := 0; i < counter; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(0)
				if err != nil {
					t.Error(err)
				}
			}()
		}

		wg.Wait()
	
		if testCounter != 2 {
			t.Errorf("%s: testCounter %v != 2", getName(i), testCounter)
		}
	}
}

func TestLoaderClearVisitorFunc(t *testing.T) {
	for i := range 4 {
		var clearCounter, evictCounter, loaderCounter int64
		counter := 1000
		builder := New(7).
			LoaderFunc(func(key interface{}) (interface{}, time.Duration, error) {
				return atomic.AddInt64(&loaderCounter, 1), NoExpiration, nil
			}).
			EvictedFunc(func(key, value interface{}) {
				atomic.AddInt64(&evictCounter, 1)
			}).
			ClearVisitorFunc(func(k, v interface{}) {
				atomic.AddInt64(&clearCounter, 1)
			})
		cache := getCache(i, builder)
		var wg sync.WaitGroup
		for i := 0; i < counter; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(i)
				if err != nil {
					t.Error(err)
				}
			}()
		}

		wg.Wait()

		if loaderCounter != int64(counter) {
			t.Errorf("%s: loaderCounter != %v", getName(i), loaderCounter)
		}

		cache.Clear()

		if evictCounter+clearCounter != loaderCounter {
			t.Logf("%s: evictCounter: %d", getName(i), evictCounter)
			t.Logf("%s: clearCounter: %d", getName(i), clearCounter)
			t.Logf("%s: loaderCounter: %d", getName(i), loaderCounter)
			t.Errorf("%s: load != evict+clear", getName(i))
		}
	}
}

func TestExpiredItems(t *testing.T) {

	for i := range 4 {
		t.Run(getName(i), func(t *testing.T) {
			testExpiredItems(t, getCache(i, New(8).
				Expiration(time.Millisecond*100)))
		})
	}
}
