package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestARCGet(t *testing.T) {
	size := 1000
	gc := buildTestCache(t, size).ARC()
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingARCGet(t *testing.T) {
	size := 1000
	numbers := 1000
	testGetCache(t, buildTestLoadingCache(t, size, loader).ARC(), numbers)
}

func TestARCLength(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, 2, time.Millisecond).ARC()
	gc.Get("test1")
	gc.Get("test2")
	gc.Get("test3")
	length := gc.Len(true)
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", expectedLength, length)
	}
	time.Sleep(time.Millisecond)
	gc.Get("test4")
	length = gc.Len(true)
	expectedLength = 1
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", expectedLength, length)
	}
}

func TestARCEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := cacheSize + 1
	gc := buildTestLoadingCache(t, cacheSize, loader).ARC()

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestARCClearCache(t *testing.T) {
	cacheSize := 10
	clearCount := 0
	gc := New(cacheSize).
		LoaderFunc(loader).
		ClearVisitorFunc(func(k, v interface{}) {
			clearCount++
		}).
		ARC()

	for i := 0; i < cacheSize; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	gc.Clear()

	if clearCount != cacheSize {
		t.Errorf("failed to clear everything")
	}
}

func TestARCHas(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, 2, 10*time.Millisecond).ARC()

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
