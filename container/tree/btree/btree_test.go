// Copyright 2020 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package btree

// SEED=1603717878394178000 go test -run Random

import (
	"fmt"
	"math/rand/v2"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var seed int64

func init() {
	var err error
	seed, err = strconv.ParseInt(os.Getenv("SEED"), 10, 64)
	if err != nil {
		seed = time.Now().UnixNano()
	}
	fmt.Printf("seed: %d\n", seed)
}

func randKeys(N int) (keys []string) {
	format := fmt.Sprintf("%%0%dd", len(fmt.Sprintf("%d", N-1)))
	for _, i := range rand.Perm(N) {
		keys = append(keys, fmt.Sprintf(format, i))
	}
	return
}

var flatLeaf = true

func (tr *BTree[T]) print() {
	tr.root.print(0, tr.Height())
}

func (n *node[T]) print(level, height int) {
	if n == nil {
		println("NIL")
		return
	}
	if height == 0 && flatLeaf {
		fmt.Printf("%v", strings.Repeat("  ", level))
	}
	for i := int16(0); i < n.numItems; i++ {
		if height > 0 {
			n.children[i].print(level+1, height-1)
		}
		if height > 0 || (height == 0 && !flatLeaf) {
			fmt.Printf("%v%v\n", strings.Repeat("  ", level), n.items[i])
		} else {
			if i > 0 {
				fmt.Printf(",")
			}
			fmt.Printf("%v", n.items[i])
		}
	}
	if height == 0 && flatLeaf {
		fmt.Printf("\n")
	}
	if height > 0 {
		n.children[n.numItems].print(level+1, height-1)
	}
}

func (tr *BTree[T]) deepPrint() {
	fmt.Printf("%#v\n", tr)
	tr.root.deepPrint(0)
}

func (n *node[T]) deepPrint(level int) {
	if n == nil {
		fmt.Printf("%v %#v\n", strings.Repeat("  ", level), n)
		return
	}
	fmt.Printf("%v count: %v\n", strings.Repeat("  ", level), n.numItems)
	fmt.Printf("%v items: %v\n", strings.Repeat("  ", level), n.items[:n.numItems])
	if !n.leaf {
		fmt.Printf("%v child: %v\n", strings.Repeat("  ", level), n.children)
	}
	if !n.leaf {
		for i := int16(0); i < n.numItems; i++ {
			n.children[i].deepPrint(level + 1)
		}
		n.children[n.numItems].deepPrint(level + 1)
	}
}

func stringsEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type pair struct {
	key   string
	value interface{}
}

func pairLess(a, b pair) int {
	if a.key < b.key {
		return -1
	} else if a.key > b.key {
		return 1
	}
	return 0
}

func TestDescend(t *testing.T) {
	tr := New[pair](pairLess)
	var count int
	tr.Descend(&pair{"1", nil}, func(item pair) bool {
		count++
		return true
	})
	if count > 0 {
		t.Fatalf("expected 0, got %v", count)
	}
	var keys []string
	for i := 0; i < 1000; i += 10 {
		keys = append(keys, fmt.Sprintf("%03d", i))
		tr.Set(pair{keys[len(keys)-1], nil})
	}
	var exp []string
	tr.Descend(nil, func(item pair) bool {
		exp = append(exp, item.key)
		return true
	})
	for i := 999; i >= 0; i-- {
		var key string
		key = fmt.Sprintf("%03d", i)
		var all []string
		tr.Descend(&pair{key, nil}, func(item pair) bool {
			all = append(all, item.key)
			return true
		})
		for len(exp) > 0 && key < exp[0] {
			exp = exp[1:]
		}
		var count int
		tr.Descend(&pair{key, nil}, func(item pair) bool {
			if count == (i+1)%maxItems {
				return false
			}
			count++
			return true
		})
		if count > len(exp) {
			t.Fatalf("expected 1, got %v", count)
		}
		if !stringsEquals(exp, all) {
			fmt.Printf("exp: %v\n", exp)
			fmt.Printf("all: %v\n", all)
			t.Fatal("mismatch")
		}
	}
}

func TestAscend(t *testing.T) {
	tr := New(pairLess)
	var count int
	tr.Ascend(&pair{"1", nil}, func(item pair) bool {
		count++
		return true
	})
	if count > 0 {
		t.Fatalf("expected 0, got %v", count)
	}
	var keys []string
	for i := 0; i < 1000; i += 10 {
		keys = append(keys, fmt.Sprintf("%03d", i))
		tr.Set(pair{keys[len(keys)-1], nil})
	}
	exp := keys
	for i := -1; i < 1000; i++ {
		var key string
		if i == -1 {
			key = ""
		} else {
			key = fmt.Sprintf("%03d", i)
		}
		var all []string
		tr.Ascend(&pair{key, nil}, func(item pair) bool {
			all = append(all, item.key)
			return true
		})

		for len(exp) > 0 && key > exp[0] {
			exp = exp[1:]
		}
		var count int
		tr.Ascend(&pair{key, nil}, func(item pair) bool {
			if count == (i+1)%maxItems {
				return false
			}
			count++
			return true
		})
		if count > len(exp) {
			t.Fatalf("expected 1, got %v", count)
		}
		if !stringsEquals(exp, all) {
			t.Fatal("mismatch")
		}
	}
}

func TestBTree(t *testing.T) {

	N := 10000
	tr := New(pairLess)
	tr.sane()
	keys := randKeys(N)

	// insert all items
	for _, key := range keys {
		_, ok := tr.Set(pair{key, key})
		tr.sane()
		if ok {
			t.Fatal("expected not ok")
		}
	}

	// check length
	if tr.Len() != len(keys) {
		t.Fatalf("expected %v, got %v", len(keys), tr.Len())
	}

	// get each value
	for _, key := range keys {
		item, ok := tr.Get(pair{key, nil})
		if !ok || item.value != key {
			t.Fatalf("expected '%v', got '%v'", key, item.value)
		}
	}

	// scan all items
	var last string
	all := make(map[string]any)
	tr.Ascend(nil, func(item pair) bool {
		key := item.key
		value := item.value
		if key <= last {
			t.Fatal("out of order")
		}
		if value.(string) != key {
			t.Fatalf("mismatch")
		}
		last = key
		all[key] = value
		return true
	})
	if len(all) != len(keys) {
		t.Fatalf("expected '%v', got '%v'", len(keys), len(all))
	}

	// reverse all items
	var prev string
	all = make(map[string]interface{})
	tr.Descend(nil, func(item pair) bool {
		key := item.key
		value := item.value
		if prev != "" && key >= prev {
			t.Fatal("out of order")
		}
		if value.(string) != key {
			t.Fatalf("mismatch")
		}
		prev = key
		all[key] = value
		return true
	})
	if len(all) != len(keys) {
		t.Fatalf("expected '%v', got '%v'", len(keys), len(all))
	}

	// try to get an invalid item
	_, ok := tr.Get(pair{"invalid", nil})
	if ok {
		t.Fatal("expected not ok")
	}

	// scan and quit at various steps
	for i := 0; i < 100; i++ {
		var j int
		tr.Ascend(nil, func(item pair) bool {
			if j == i {
				return false
			}
			j++
			return true
		})
	}

	// reverse and quit at various steps
	for i := 0; i < 100; i++ {
		var j int
		tr.Descend(nil, func(item pair) bool {
			if j == i {
				return false
			}
			j++
			return true
		})
	}

	// delete half the items
	for _, key := range keys[:len(keys)/2] {
		item, ok := tr.Delete(pair{key, nil})
		if !ok {
			t.Fatal("expected ok")
		}
		value := item.value
		if value == nil || value.(string) != key {
			t.Fatalf("expected '%v', got '%v'", key, value)
		}
	}

	// check length
	if tr.Len() != len(keys)/2 {
		t.Fatalf("expected %v, got %v", len(keys)/2, tr.Len())
	}

	// try delete half again
	for _, key := range keys[:len(keys)/2] {
		_, ok := tr.Delete(pair{key, nil})
		tr.sane()
		if ok {
			t.Fatal("expected not ok")
		}
	}

	// try delete half again
	for _, key := range keys[:len(keys)/2] {
		_, ok := tr.Delete(pair{key, nil})
		tr.sane()
		if ok {
			t.Fatal("expected not ok")
		}
	}

	// check length
	if tr.Len() != len(keys)/2 {
		t.Fatalf("expected %v, got %v", len(keys)/2, tr.Len())
	}

	// scan items
	last = ""
	all = make(map[string]interface{})
	tr.Ascend(nil, func(item pair) bool {
		key := item.key
		value := item.value
		if key <= last {
			t.Fatal("out of order")
		}
		if value.(string) != key {
			t.Fatalf("mismatch")
		}
		last = key
		all[key] = value
		return true
	})
	if len(all) != len(keys)/2 {
		t.Fatalf("expected '%v', got '%v'", len(keys), len(all))
	}

	// replace second half
	for _, key := range keys[len(keys)/2:] {
		item, ok := tr.Set(pair{key, key})
		tr.sane()
		if !ok {
			t.Fatal("expected ok")
		}
		value := item.value
		if value == nil || value.(string) != key {
			t.Fatalf("expected '%v', got '%v'", key, value)
		}
	}

	// delete next half the items
	for _, key := range keys[len(keys)/2:] {
		item, ok := tr.Delete(pair{key, nil})
		tr.sane()
		if !ok {
			t.Fatal("expected ok")
		}
		value := item.value
		if value == nil || value.(string) != key {
			t.Fatalf("expected '%v', got '%v'", key, value)
		}
	}

	// check length
	if tr.Len() != 0 {
		t.Fatalf("expected %v, got %v", 0, tr.Len())
	}

	// do some stuff on an empty tree
	_, ok = tr.Get(pair{keys[0], nil})
	if ok {
		t.Fatal("expected not ok")
	}
	tr.Ascend(nil, func(item pair) bool {
		t.Fatal("should not be reached")
		return true
	})
	tr.Descend(nil, func(item pair) bool {
		t.Fatal("should not be reached")
		return true
	})

	_, ok = tr.Delete(pair{"invalid", nil})
	tr.sane()
	if ok {
		t.Fatal("expected not ok")
	}
}

func BenchmarkTidwallSequentialSet(b *testing.B) {
	tr := New[int](intLess)
	keys := rand.Perm(b.N)
	sort.Ints(keys)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Set(keys[i])
	}
}

func BenchmarkTidwallSequentialGet(b *testing.B) {
	tr := New[int](intLess)
	keys := rand.Perm(b.N)
	sort.Ints(keys)
	for i := 0; i < b.N; i++ {
		tr.Set(keys[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Get(keys[i])
	}
}

func BenchmarkTidwallRandomSet(b *testing.B) {
	tr := New(intLess)
	keys := rand.Perm(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Set(keys[i])
	}
}

func BenchmarkTidwallRandomGet(b *testing.B) {
	tr := New(intLess)
	keys := rand.Perm(b.N)
	for i := 0; i < b.N; i++ {
		tr.Set(keys[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Get(keys[i])
	}
}

func BenchmarkTidwallSequentialSetHint(b *testing.B) {
	tr := New(intLess)
	keys := rand.Perm(b.N)
	sort.Ints(keys)
	b.ResetTimer()
	var hint PathHint
	for i := 0; i < b.N; i++ {
		tr.SetHint(keys[i], &hint)
	}
}

func BenchmarkTidwallSequentialGetHint(b *testing.B) {
	// println("\n----------------------------------------------------------------")
	tr := New(intLess)
	keys := rand.Perm(b.N)
	sort.Ints(keys)
	for i := 0; i < b.N; i++ {
		tr.Set(keys[i])
	}
	b.ResetTimer()
	var hint PathHint
	for i := 0; i < b.N; i++ {
		tr.GetHint(keys[i], &hint)
		// fmt.Printf("%064b\n", hint)
	}
}

func BenchmarkTidwallLoad(b *testing.B) {
	tr := New(intLess)
	keys := rand.Perm(b.N)
	sort.Ints(keys)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Load(keys[i])
	}
}

func TestBTreeOne(t *testing.T) {
	tr := New(pairLess)
	tr.Set(pair{"1", "1"})
	tr.Delete(pair{"1", nil})
	tr.Set(pair{"1", "1"})
	tr.Delete(pair{"1", nil})
	tr.Set(pair{"1", "1"})
	tr.Delete(pair{"1", nil})
}

func TestBTree256(t *testing.T) {
	tr := New(pairLess)
	var n int
	for j := 0; j < 2; j++ {
		for _, i := range rand.Perm(256) {
			tr.Set(pair{fmt.Sprintf("%d", i), i})
			n++
			if tr.Len() != n {
				t.Fatalf("expected 256, got %d", n)
			}
		}
		for _, i := range rand.Perm(256) {
			item, ok := tr.Get(pair{fmt.Sprintf("%d", i), nil})
			if !ok {
				t.Fatal("expected ok")
			}
			if item.value.(int) != i {
				t.Fatalf("expected %d, got %d", i, item.value.(int))
			}
		}
		for _, i := range rand.Perm(256) {
			tr.Delete(pair{fmt.Sprintf("%d", i), nil})
			n--
			if tr.Len() != n {
				t.Fatalf("expected 256, got %d", n)
			}
		}
		for _, i := range rand.Perm(256) {
			_, ok := tr.Get(pair{fmt.Sprintf("%d", i), nil})
			if ok {
				t.Fatal("expected not ok")
			}
		}
	}
}
func shuffle(keys []int) {
	for i := range keys {
		j := rand.IntN(i + 1)
		keys[i], keys[j] = keys[j], keys[i]
	}
}

func intLess(a, b int) int {
	return a - b
}

func TestRandom(t *testing.T) {
	N := 200000
	keys := rand.Perm(N)
	func() {
		defer func() {
			msg := fmt.Sprint(recover())
			if msg != "nil cmp func" {
				t.Fatal("expected 'nil cmp func' panic")
			}
		}()
		New[int](nil)
		t.Fatalf("reached invalid code")
	}()
	tr := New(intLess)
	checkSane := func() {
		// tr.sane()
	}
	checkSane()
	if _, ok := tr.Min(); ok {
		t.Fatalf("expected nil")
	}
	if _, ok := tr.Max(); ok {
		t.Fatalf("expected nil")
	}
	if _, ok := tr.PopMin(); ok {
		t.Fatalf("expected nil")
	}
	if _, ok := tr.PopMax(); ok {
		t.Fatalf("expected nil")
	}
	if tr.Height() != 0 {
		t.Fatalf("expected 0, got %d", tr.Height())
	}
	checkSane()
	// keys = keys[:rand.Intn(len(keys))]
	shuffle(keys)
	for i := 0; i < len(keys); i++ {
		_, ok := tr.Set(keys[i])
		checkSane()
		if ok {
			t.Fatalf("expected not ok")
		}
		if i%12345 == 0 {
			tr.sane()
		}
	}
	tr.sane()
	sort.Ints(keys)
	var n int
	tr.Ascend(nil, func(item int) bool {
		n++
		return false
	})
	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	n = 0
	tr.Ascend(nil, func(item int) bool {
		if item != keys[n] {
			t.Fatalf("expected %d, got %d", keys[n], item)
		}
		n++
		return true
	})
	if n != len(keys) {
		t.Fatalf("expected %d, got %d", len(keys), n)
	}
	if tr.Len() != len(keys) {
		t.Fatalf("expected %d, got %d", tr.Len(), len(keys))
	}

	n = 0
	tr.Descend(nil, func(item int) bool {
		n++
		return false
	})
	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
	n = 0
	tr.Descend(nil, func(item int) bool {
		if item != keys[len(keys)-n-1] {
			t.Fatalf("expected %d, got %d", keys[len(keys)-n-1], item)
		}
		n++
		return true
	})
	if n != len(keys) {
		t.Fatalf("expected %d, got %d", len(keys), n)
	}
	if tr.Len() != len(keys) {
		t.Fatalf("expected %d, got %d", tr.Len(), len(keys))
	}

	checkSane()

	// tr.deepPrint()

	n = 0
	for i := 0; i < 1000; i++ {
		n := 0
		tr.Ascend(nil, func(item int) bool {
			if n == i {
				return false
			}
			n++
			return true
		})
		if n != i {
			t.Fatalf("expected %d, got %d", i, n)
		}
	}

	n = 0
	for i := 0; i < 1000; i++ {
		n = 0
		tr.Descend(nil, func(item int) bool {
			if n == i {
				return false
			}
			n++
			return true
		})
		if n != i {
			t.Fatalf("expected %d, got %d", i, n)
		}
	}

	sort.Ints(keys)
	for i := 0; i < len(keys); i++ {
		var res int
		var j int
		tr.Ascend(&keys[i], func(item int) bool {
			if j == 0 {
				res = item
			}
			j++
			return j == i%500
		})
		if res != keys[i] {
			t.Fatal("not equal")
		}
	}
	for i := len(keys) - 1; i >= 0; i-- {
		var res int
		var j int
		tr.Descend(&keys[i], func(item int) bool {
			if j == 0 {
				res = item
			}
			j++
			return j == i%500
		})
		if res != keys[i] {
			t.Fatal("not equal")
		}
	}

	if tr.Height() == 0 {
		t.Fatalf("expected non-zero")
	}
	if v, ok := tr.Min(); !ok || v != keys[0] {
		t.Fatalf("expected '%v', got '%v'", keys[0], v)
	}
	if v, ok := tr.Max(); !ok || v != keys[len(keys)-1] {
		t.Fatalf("expected '%v', got '%v'", keys[len(keys)-1], v)
	}
	min, ok := tr.PopMin()
	checkSane()
	if !ok || min != keys[0] {
		t.Fatalf("expected '%v', got '%v'", keys[0], min)
	}
	max, ok := tr.PopMax()
	checkSane()
	if !ok || max != keys[len(keys)-1] {
		t.Fatalf("expected '%v', got '%v'", keys[len(keys)-1], max)
	}
	tr.Set(min)
	tr.Set(max)
	shuffle(keys)
	var hint PathHint
	for i := 0; i < len(keys); i++ {
		prev, ok := tr.Get(keys[i])
		if !ok || prev != keys[i] {
			t.Fatalf("expected '%v', got '%v'", keys[i], prev)
		}
		prev, ok = tr.GetHint(keys[i], &hint)
		if !ok || prev != keys[i] {
			t.Fatalf("expected '%v', got '%v'", keys[i], prev)
		}
	}
	sort.Ints(keys)
	for i := 0; i < len(keys); i++ {
		item, ok := tr.PopMin()
		if !ok || item != keys[i] {
			t.Fatalf("expected '%v', got '%v'", keys[i], item)
		}
	}
	for i := 0; i < len(keys); i++ {
		_, ok := tr.Set(keys[i])
		if ok {
			t.Fatalf("expected not ok")
		}
	}
	for i := len(keys) - 1; i >= 0; i-- {
		item, ok := tr.PopMax()
		if !ok || item != keys[i] {
			t.Fatalf("expected '%v', got '%v'", keys[i], item)
		}
	}
	for i := 0; i < len(keys); i++ {
		_, ok := tr.Set(keys[i])
		if ok {
			t.Fatalf("expected not ok")
		}
	}
	checkSane()

	shuffle(keys)
	item, ok := tr.Delete(keys[len(keys)/2])
	checkSane()
	if item != keys[len(keys)/2] {
		t.Fatalf("expected '%v', got '%v'", keys[len(keys)/2], item)
	}
	item2, ok := tr.Delete(keys[len(keys)/2])
	checkSane()
	if ok {
		t.Fatalf("expected not ok, got '%v'", item2)
	}

	tr.Set(item)
	checkSane()
	for i := 0; i < len(keys); i++ {
		prev, ok := tr.Delete(keys[i])
		checkSane()
		if !ok || prev != keys[i] {
			t.Fatalf("expected '%v', got '%v'", keys[i], prev)
		}
		prev, ok = tr.Get(keys[i])
		if ok {
			t.Fatalf("expected not ok")
		}
		prev, ok = tr.GetHint(keys[i], &hint)
		if ok {
			t.Fatalf("expected not ok")
		}
		if i%12345 == 0 {
			tr.sane()
		}
	}
	if tr.Height() != 0 {
		t.Fatalf("expected 0, got %d", tr.Height())
	}
	shuffle(keys)
	for i := 0; i < len(keys); i++ {
		_, ok := tr.Load(keys[i])
		checkSane()
		if ok {
			t.Fatalf("expected not ok")
		}
	}
	for i := 0; i < len(keys); i++ {
		prev, ok := tr.Get(keys[i])
		if !ok || prev != keys[i] {
			t.Fatalf("expected '%v', got '%v'", keys[i], prev)
		}
	}
	shuffle(keys)
	for i := 0; i < len(keys); i++ {
		prev, ok := tr.Delete(keys[i])
		checkSane()
		if !ok || prev != keys[i] {
			t.Fatalf("expected '%v', got '%v'", keys[i], prev)
		}
		prev, ok = tr.Get(keys[i])
		if ok {
			t.Fatalf("expected not ok")
		}
	}
	sort.Ints(keys)
	for i := 0; i < len(keys); i++ {
		_, ok := tr.Load(keys[i])
		checkSane()
		if ok {
			t.Fatalf("expected not ok")
		}
	}
	shuffle(keys)
	item, ok = tr.Load(keys[0])
	checkSane()
	if !ok || item != keys[0] {
		t.Fatalf("expected '%v', got '%v'", keys[0], item)
	}
}

// type intT struct {
// 	val int
// }

// func (i *intT) Less(other btree.Item) bool {
// 	return i.val < other.(*intT).val
// }

// func TestPerf(t *testing.T) {
// 	less := func(a, b interface{}) bool {
// 		return a.(*intT).val < b.(*intT).val
// 	}
// 	N := 1_000_000
// 	keys := make([]*intT, N)
// 	for i := 0; i < N; i++ {
// 		keys[i] = &intT{i}
// 	}
// 	lotsa.Output = os.Stdout
// 	lotsa.MemUsage = true

// 	sortInts := func() {
// 		sort.Slice(keys, func(i, j int) bool {
// 			return less(keys[i], keys[j])
// 		})
// 	}

// 	shuffleInts := func() {
// 		for i := range keys {
// 			j := rand.Intn(i + 1)
// 			keys[i], keys[j] = keys[j], keys[i]
// 		}
// 	}

// 	print("google:  set-seq       ")
// 	tr2 := btree.New(256)
// 	sortInts()
// 	lotsa.Ops(N, 1, func(i, _ int) {
// 		tr2.ReplaceOrInsert(keys[i])
// 	})
// 	print("tidwall: set-seq       ")
// 	tr := New(less)
// 	sortInts()
// 	lotsa.Ops(N, 1, func(i, _ int) {
// 		tr.Set(keys[i])
// 	})
// 	print("tidwall: set-seq-hint  ")
// 	tr = New(less)
// 	sortInts()
// 	var hint PathHint
// 	lotsa.Ops(N, 1, func(i, _ int) {
// 		tr.SetHint(keys[i], &hint)
// 	})
// 	print("tidwall: load-seq      ")
// 	tr = New(less)
// 	sortInts()
// 	lotsa.Ops(N, 1, func(i, _ int) {
// 		tr.Load(keys[i])
// 	})
// 	print("go-arr:  load-seq      ")
// 	var arr []interface{}
// 	sortInts()
// 	lotsa.Ops(N, 1, func(i, _ int) {
// 		arr = append(arr, keys[i])
// 	})
// 	print("google:  set-rand      ")
// 	tr2 = btree.New(256)
// 	shuffleInts()
// 	lotsa.Ops(N, 1, func(i, _ int) {
// 		tr2.ReplaceOrInsert(keys[i])
// 	})
// 	print("tidwall: set-rand      ")
// 	tr = New(less)
// 	shuffleInts()
// 	lotsa.Ops(N, 1, func(i, _ int) {
// 		tr.Set(keys[i])
// 	})
// 	print("tidwall: set-rand-hint ")
// 	tr = New(less)
// 	shuffleInts()
// 	lotsa.Ops(N, 1, func(i, _ int) {
// 		tr.SetHint(keys[i], &hint)
// 	})
// 	print("tidwall: load-rand     ")
// 	tr = New(less)
// 	shuffleInts()
// 	lotsa.Ops(N, 1, func(i, _ int) {
// 		tr.Load(keys[i])
// 	})
// }

func (n *node[T]) saneheight(height, maxheight int) bool {
	if n.leaf {
		if height != maxheight {
			return false
		}
	} else {
		i := int16(0)
		for ; i < n.numItems; i++ {
			if !n.children[i].saneheight(height+1, maxheight) {
				return false
			}
		}
		if !n.children[i].saneheight(height+1, maxheight) {
			return false
		}
	}
	return true
}

// btree_saneheight returns true if the height of all leaves match the height
// of the btree.
func (tr *BTree[T]) saneheight() bool {
	height := tr.Height()
	if tr.root != nil {
		if height == 0 {
			return false
		}
		return tr.root.saneheight(1, height)
	}
	return height == 0
}

func (n *node[T]) deepcount() int {
	count := int(n.numItems)
	if !n.leaf {
		for i := int16(0); i <= n.numItems; i++ {
			count += n.children[i].deepcount()
		}
	}
	return count
}

// btree_deepcount returns the number of items in the btree.
func (tr *BTree[T]) deepcount() int {
	if tr.root != nil {
		return tr.root.deepcount()
	}
	return 0
}

func (tr *BTree[T]) nodesaneprops(n *node[T], height int) bool {
	if height == 1 {
		if n.numItems < 1 || n.numItems > maxItems {
			return false
		}
	} else {
		if n.numItems < minItems || n.numItems > maxItems {
			return false
		}
	}
	if !n.leaf {
		for i := int16(0); i < n.numItems; i++ {
			if !tr.nodesaneprops(n.children[i], height+1) {
				return false
			}
		}
		if !tr.nodesaneprops(n.children[n.numItems], height+1) {
			return false
		}
	}
	return true
}

func (tr *BTree[T]) saneprops() bool {
	if tr.root != nil {
		return tr.nodesaneprops(tr.root, 1)
	}
	return true
}

func (tr *BTree[T]) saneorder() bool {
	var last T
	var count int
	var bad bool
	tr.Walk(func(items []T) {
		for _, item := range items {
			if bad {
				return
			}
			if count != 0 {
				if tr.cmp(last, item) >= 0 {
					bad = true
					return
				}
			}

			last = item
			count++
		}
	})
	return !bad && count == tr.length
}

// btree_sane returns true if the entire btree and every node are valid.
// - height of all leaves are the equal to the btree height.
// - deep count matches the btree count.
// - all nodes have the correct number of items and counts.
// - all items are in order.
func (tr *BTree[T]) sane() {
	if tr == nil {
		panic("nil tree")
	}
	if !tr.saneheight() {
		panic("!sane-height")
	}
	if tr.Len() != tr.length || tr.deepcount() != tr.length {
		panic("!sane-count")
	}
	if !tr.saneprops() {
		panic("!sane-props")
	}
	if !tr.saneorder() {
		panic("!sane-order")
	}
}

// func TestActionAscend(t *testing.T) {
// 	tr := New(intLess)

// 	for i := 0; i < 10000; i += 10 {
// 		tr.Set(i)
// 	}
// 	tr.ActionAscend(nil, func(item interface{}) (interface{}, ActionCount) {
// 		if item.(int) > 3440 && item.(int) < 8710 {
// 			return nil, ActionDelete
// 		}
// 		return nil, ActionSkip
// 	})

// 	tr.ActionAscend(nil, func(item interface{}) (interface{}, ActionCount) {

// 		println(item.(int))
// 		return nil, ActionSkip
// 	})

// }
