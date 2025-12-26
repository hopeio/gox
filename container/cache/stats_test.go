package cache

import (
	"testing"
	"time"
)

func TestStats(t *testing.T) {
	var cases = []struct {
		hit  int
		miss int
		rate float64
	}{
		{3, 1, 0.75},
		{0, 1, 0.0},
		{3, 0, 1.0},
		{0, 0, 0.0},
	}

	for _, cs := range cases {
		st := &stats{}
		for i := 0; i < cs.hit; i++ {
			st.IncrHitCount()
		}
		for i := 0; i < cs.miss; i++ {
			st.IncrMissCount()
		}
		if rate := st.HitRate(); rate != cs.rate {
			t.Errorf("%v != %v", rate, cs.rate)
		}
	}
}

func getter(key interface{}) (interface{}, time.Duration, error) {
	return key, 0, nil
}

func TestCacheStats(t *testing.T) {
	var cases = []struct {
		builder func() *Cache
		rate    float64
	}{
		{
			builder: func() *Cache {
				cc := New(32).LRU()
				cc.Set(0, 0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() *Cache {
				cc := New(32).LFU()
				cc.Set(0, 0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() *Cache {
				cc := New(32).ARC()
				cc.Set(0, 0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() *Cache {
				cc := New(32).
					LoaderFunc(getter).
					LRU()
				cc.Set(0, 0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() *Cache {
				cc := New(32).
					LoaderFunc(getter).
					LFU()
				cc.Set(0, 0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() *Cache {
				cc := New(32).
					LoaderFunc(getter).
					ARC()
				cc.Set(0, 0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
	}

	for i, cs := range cases {
		cc := cs.builder()
		if rate := cc.HitRate(); rate != cs.rate {
			t.Errorf("case-%v: %v != %v", i, rate, cs.rate)
		}
	}
}
