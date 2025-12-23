package cache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoaderFunc(t *testing.T) {
	size := 2
	var testCaches = []*CacheBuilder{
		New(size).LRU(),
		New(size).LFU(),
		New(size).ARC(),
	}
	for _, builder := range testCaches {
		var testCounter int64
		counter := 1000
		cache := builder.
			LoaderFunc(func(key interface{}) (interface{}, time.Duration, error) {
				time.Sleep(10 * time.Millisecond)
				return atomic.AddInt64(&testCounter, 1), 0, nil
			}).
			EvictedFunc(func(key, value interface{}) {
				panic(key)
			}).Build()

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
	size := 2
	var testCaches = []*CacheBuilder{
		New(size).LRU(),
		New(size).LFU(),
		New(size).ARC(),
	}
	for _, builder := range testCaches {
		var testCounter int64
		counter := 1000
		cache := builder.
			LoaderFunc(func(key interface{}) (interface{}, time.Duration, error) {
				return atomic.AddInt64(&testCounter, 1), NoExpiration, nil
			}).
			EvictedFunc(func(key, value interface{}) {
				panic(key)
			}).Build()

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
	size := 2
	var testCaches = []*CacheBuilder{
		New(size).LRU(),
		New(size).LFU(),
		New(size).ARC(),
	}
	for _, builder := range testCaches {
		var testCounter int64
		counter := 1000
		expire := 200 * time.Millisecond
		cache := builder.
			LoaderFunc(func(key interface{}) (interface{}, time.Duration, error) {
				return atomic.AddInt64(&testCounter, 1), expire, nil
			}).
			Build()

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
			t.Errorf("testCounter != %v", testCounter)
		}
	}
}

func TestLoaderPurgeVisitorFunc(t *testing.T) {
	size := 7
	tests := []struct {
		name         string
		cacheBuilder *CacheBuilder
	}{
		{
			name:         "lru",
			cacheBuilder: New(size).LRU(),
		},
		{
			name:         "lfu",
			cacheBuilder: New(size).LFU(),
		},
		{
			name:         "arc",
			cacheBuilder: New(size).ARC(),
		},
	}

	for _, test := range tests {
		var purgeCounter, evictCounter, loaderCounter int64
		counter := 1000
		cache := test.cacheBuilder.
			LoaderFunc(func(key interface{}) (interface{}, time.Duration, error) {
				return atomic.AddInt64(&loaderCounter, 1), NoExpiration, nil
			}).
			EvictedFunc(func(key, value interface{}) {
				atomic.AddInt64(&evictCounter, 1)
			}).
			PurgeVisitorFunc(func(k, v interface{}) {
				atomic.AddInt64(&purgeCounter, 1)
			}).
			Build()

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
			t.Errorf("%s: loaderCounter != %v", test.name, loaderCounter)
		}

		cache.Purge()

		if evictCounter+purgeCounter != loaderCounter {
			t.Logf("%s: evictCounter: %d", test.name, evictCounter)
			t.Logf("%s: purgeCounter: %d", test.name, purgeCounter)
			t.Logf("%s: loaderCounter: %d", test.name, loaderCounter)
			t.Errorf("%s: load != evict+purge", test.name)
		}
	}
}

func TestExpiredItems(t *testing.T) {
	var tps = []string{
		TYPE_LRU,
		TYPE_LFU,
		TYPE_ARC,
	}
	for _, tp := range tps {
		t.Run(tp, func(t *testing.T) {
			testExpiredItems(t, tp)
		})
	}
}
