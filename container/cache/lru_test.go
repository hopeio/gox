package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestLRUGet(t *testing.T) {
	size := 1000
	gc := buildTestCache(t, size).LRU()
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingLRUGet(t *testing.T) {
	size := 1000
	gc := buildTestLoadingCache(t, size, loader).LRU()
	testGetCache(t, gc, size)
}

func TestLRULength(t *testing.T) {
	gc := buildTestLoadingCache(t, 1000, loader).LRU()
	gc.Get("test1")
	gc.Get("test2")
	length := gc.Len(true)
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
}

func TestLRUEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildTestLoadingCache(t, cacheSize, loader).LRU()

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestLRUHas(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, 2, 10*time.Millisecond).LRU()

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			gc.Get("test1")
			gc.Get("test2")

			if gc.Has("test0") {
				t.Fatal("should not have test0")
			}
			if !gc.Has("test1") {
				t.Fatal("should have test1")
			}
			if !gc.Has("test2") {
				t.Fatal("should have test2")
			}

			time.Sleep(20 * time.Millisecond)

			if gc.Has("test0") {
				t.Fatal("should not have test0")
			}
			if gc.Has("test1") {
				t.Fatal("should not have test1")
			}
			if gc.Has("test2") {
				t.Fatal("should not have test2")
			}
		})
	}
}
