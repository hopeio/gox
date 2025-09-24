package skiplist

import (
	"math/rand"

	"github.com/hopeio/gox/cmp"
)

// A SkipList maintains an ordered collection of key:valkue pairs.
// It support insertion, lookup, and deletion operations with O(log n) time complexity
// Paper: Pugh, William (June 1990). "Skip lists: a probabilistic alternative to balanced
// trees". Communications of the ACM 33 (6): 668–676
type SkipList[K any, V any] struct {
	header   *skiplistitem[K, V]
	len      int
	MaxLevel int
	compare  cmp.LessFunc[K]
	zero     V
}

// New returns a skiplist.
func New[K any, V any](compare cmp.LessFunc[K]) *SkipList[K, V] {
	return &SkipList[K, V]{
		header:   &skiplistitem[K, V]{forward: []*skiplistitem[K, V]{nil}},
		MaxLevel: 32,
		compare:  compare,
	}
}

// Len returns the length of given skiplist.
func (s *SkipList[K, V]) Len() int {
	return s.len
}

// Set sets given k and v pair into the skiplist.
func (s *SkipList[K, V]) Set(k K, v V) {
	// s.level starts from 0, we need to allocate one
	update := make([]*skiplistitem[K, V], s.level()+1, s.effectiveMaxLevel()+1) // make(type, len, cap)

	x := s.path(s.header, update, k)
	if x != nil && (s.compare(x.k, k) || s.compare(x.k, k)) { // if key Exist, update
		x.v = v
		return
	}

	newl := s.randomLevel()
	if curl := s.level(); newl > curl {
		for i := curl + 1; i <= newl; i++ {
			update = append(update, s.header)
			s.header.forward = append(s.header.forward, nil)
		}
	}

	item := &skiplistitem[K, V]{
		forward: make([]*skiplistitem[K, V], newl+1, s.effectiveMaxLevel()+1),
		k:       k,
		v:       v,
	}
	for i := 0; i <= newl; i++ {
		item.forward[i] = update[i].forward[i]
		update[i].forward[i] = item
	}

	s.len++
}

func (s *SkipList[K, V]) path(x *skiplistitem[K, V], update []*skiplistitem[K, V], k K) (candidate *skiplistitem[K, V]) {
	depth := len(x.forward) - 1
	for i := depth; i >= 0; i-- {
		for x.forward[i] != nil && s.compare(x.forward[i].k, k) {
			x = x.forward[i]
		}
		if update != nil {
			update[i] = x
		}
	}
	return x.next()
}

func (s *SkipList[K, V]) randomLevel() (n int) {
	for n = 0; n < s.effectiveMaxLevel() && rand.Float64() < 0.25; n++ {
	}
	return
}

// Get returns corresponding v with given k.
func (s *SkipList[K, V]) Get(k K) (v V, ok bool) {
	x := s.path(s.header, nil, k)
	if x == nil || (s.compare(x.k, k) || s.compare(x.k, k)) {
		return s.zero, false
	}
	return x.v, true
}

// Search returns true if k is founded in the skiplist.
func (s *SkipList[K, V]) Search(k K) (ok bool) {
	x := s.path(s.header, nil, k)
	if x != nil {
		ok = true
		return
	}
	return
}

// Range interates `from` to `to` with `op`.
func (s *SkipList[K, V]) Range(from, to K, op func(v V)) {
	for start := s.path(s.header, nil, from); start.next() != nil; start = start.next() {
		if !s.compare(start.k, to) {
			return
		}

		op(start.v)
	}
}

// Del returns the deleted value if ok
func (s *SkipList[K, V]) Del(k K) (v V, ok bool) {
	update := make([]*skiplistitem[K, V], s.level()+1, s.effectiveMaxLevel())

	x := s.path(s.header, update, k)
	if x == nil || (s.compare(x.k, k) || s.compare(x.k, k)) {
		ok = false
		return
	}

	v = x.v
	for i := 0; i <= s.level() && update[i].forward[i] == x; i++ {
		update[i].forward[i] = x.forward[i]
	}
	for s.level() > 0 && s.header.forward[s.level()] == nil {
		s.header.forward = s.header.forward[:s.level()]
	}
	s.len--
	ok = true
	return
}

func (s *SkipList[K, V]) level() int {
	return len(s.header.forward) - 1
}

func (s *SkipList[K, V]) effectiveMaxLevel() int {
	if s.level() < s.MaxLevel {
		return s.MaxLevel
	}
	return s.level()
}

type skiplistitem[K any, V any] struct {
	forward []*skiplistitem[K, V]
	k       K
	v       V
}

func (s *skiplistitem[K, V]) next() *skiplistitem[K, V] {
	if len(s.forward) == 0 {
		return nil
	}
	return s.forward[0]
}
