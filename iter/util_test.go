/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package iter

import (
	"cmp"
	"iter"
	"slices"
	"testing"

	gcmp "github.com/hopeio/gox/cmp"
	gcontainer "github.com/hopeio/gox/container"
	"github.com/hopeio/gox/types"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// seqOf 构造一个 iter.Seq[T]，供包级函数测试使用。
func seqOf[T any](s []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// eq 比较两个切片是否相等（顺序敏感）。
func eq[T comparable](a, b []T) bool {
	return slices.Equal(a, b)
}

// itoa 仅测试用，避免引入 strconv。
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// ---------------------------------------------------------------------------
// file: util.go — 包级流操作函数
// ---------------------------------------------------------------------------

func TestUtil_Filter(t *testing.T) {
	got := ToSlice(Filter(seqOf([]int{1, 2, 3, 4, 5}), types.Predicate[int](func(v int) bool { return v%2 == 0 })))
	if !eq(got, []int{2, 4}) {
		t.Fatalf("Filter = %v, want [2 4]", got)
	}
}

func TestUtil_Map(t *testing.T) {
	got := ToSlice(Map(seqOf([]int{1, 2, 3}), types.UnaryFunction[int, int](func(v int) int { return v * 2 })))
	if !eq(got, []int{2, 4, 6}) {
		t.Fatalf("Map = %v, want [2 4 6]", got)
	}
}

func TestUtil_FlatMap(t *testing.T) {
	got := ToSlice(FlatMap(seqOf([]int{1, 2}), types.UnaryFunction[int, iter.Seq[int]](func(v int) iter.Seq[int] {
		return seqOf([]int{v, v * 10})
	})))
	if !eq(got, []int{1, 10, 2, 20}) {
		t.Fatalf("FlatMap = %v, want [1 10 2 20]", got)
	}
}

func TestUtil_Peek(t *testing.T) {
	var seen []int
	got := ToSlice(Peek(seqOf([]int{1, 2}), types.Consumer[int](func(v int) { seen = append(seen, v) })))
	if !eq(got, []int{1, 2}) {
		t.Fatalf("Peek result = %v, want [1 2]", got)
	}
	if !eq(seen, []int{1, 2}) {
		t.Fatalf("Peek seen = %v, want [1 2]", seen)
	}
}

func TestUtil_Distinct(t *testing.T) {
	got := ToSlice(Distinct(seqOf([]int{1, 1, 2, 3, 3}), types.UnaryFunction[int, int](func(v int) int { return v })))
	if !eq(got, []int{1, 2, 3}) {
		t.Fatalf("Distinct = %v, want [1 2 3]", got)
	}
}

func TestUtil_SortedIsSorted(t *testing.T) {
	got := ToSlice(Sorted(seqOf([]int{3, 1, 2}), types.Comparator[int](cmp.Compare[int])))
	if !eq(got, []int{1, 2, 3}) {
		t.Fatalf("Sorted = %v, want [1 2 3]", got)
	}
	if !IsSorted(seqOf([]int{1, 2, 3}), types.Comparator[int](cmp.Compare[int])) {
		t.Fatal("IsSorted(asc) = false, want true")
	}
	if IsSorted(seqOf([]int{3, 1, 2}), types.Comparator[int](cmp.Compare[int])) {
		t.Fatal("IsSorted(unordered) = true, want false")
	}
}

func TestUtil_LimitSkip(t *testing.T) {
	base := seqOf([]int{1, 2, 3, 4, 5})
	if got := ToSlice(Limit(base, 3)); !eq(got, []int{1, 2, 3}) {
		t.Fatalf("Limit = %v, want [1 2 3]", got)
	}
	if got := ToSlice(Skip(base, 2)); !eq(got, []int{3, 4, 5}) {
		t.Fatalf("Skip = %v, want [3 4 5]", got)
	}
	if got := ToSlice(Skip(Limit(base, 4), 1)); !eq(got, []int{2, 3, 4}) {
		t.Fatalf("Skip(Limit) = %v, want [2 3 4]", got)
	}
}

func TestUtil_Until(t *testing.T) {
	got := ToSlice(Until(seqOf([]int{1, 2, 3, 4}), types.Predicate[int](func(v int) bool { return v == 3 })))
	if !eq(got, []int{1, 2}) {
		t.Fatalf("Until = %v, want [1 2]", got)
	}
}

func TestUtil_UntilComparable(t *testing.T) {
	got := ToSlice(UntilComparable(seqOf([]int{1, 2, 3, 4}), 3))
	if !eq(got, []int{1, 2}) {
		t.Fatalf("UntilComparable = %v, want [1 2]", got)
	}
	got2 := ToSlice(UntilComparable(seqOf([]string{"a", "b", "c"}), "c"))
	if !eq(got2, []string{"a", "b"}) {
		t.Fatalf("UntilComparable(str) = %v, want [a b]", got2)
	}
}

func TestUtil_ForEach(t *testing.T) {
	var sum int
	ForEach(seqOf([]int{1, 2, 3}), types.Consumer[int](func(v int) { sum += v }))
	if sum != 6 {
		t.Fatalf("ForEach sum = %d, want 6", sum)
	}
}

func TestUtil_EverySomeAllAny(t *testing.T) {
	if !Every(seqOf([]int{2, 4, 6}), types.Predicate[int](func(v int) bool { return v%2 == 0 })) {
		t.Fatal("Every even = false, want true")
	}
	if Some(seqOf([]int{1, 3, 5}), types.Predicate[int](func(v int) bool { return v%2 == 0 })) {
		t.Fatal("Some even = true, want false")
	}
	if !AllMatch(seqOf([]int{1, 2, 3}), types.Predicate[int](func(v int) bool { return v > 0 })) {
		t.Fatal("AllMatch >0 = false, want true")
	}
	if !AnyMatch(seqOf([]int{1, 2, 3}), types.Predicate[int](func(v int) bool { return v == 2 })) {
		t.Fatal("AnyMatch ==2 = false, want true")
	}
}

func TestUtil_ReduceFold(t *testing.T) {
	r, ok := Reduce(seqOf([]int{1, 2, 3, 4}), types.BinaryOperator[int](func(a, b int) int { return a + b }))
	if !ok || r != 10 {
		t.Fatalf("Reduce = (%d, %v), want (10, true)", r, ok)
	}
	if _, ok := Reduce(seqOf([]int{}), types.BinaryOperator[int](func(a, b int) int { return a + b })); ok {
		t.Fatal("Reduce empty ok = true, want false")
	}
	f := Fold(seqOf([]int{1, 2, 3}), 100, types.BinaryFunction[int, int, int](func(a, b int) int { return a + b }))
	if f != 106 {
		t.Fatalf("Fold = %d, want 106", f)
	}
}

func TestUtil_FirstLastAtCount(t *testing.T) {
	if v, ok := First(seqOf([]int{7, 8, 9})); !ok || v != 7 {
		t.Fatalf("First = (%d, %v), want (7, true)", v, ok)
	}
	if _, ok := First(seqOf([]int{})); ok {
		t.Fatal("First empty ok = true, want false")
	}
	if v, ok := Last(seqOf([]int{7, 8, 9})); !ok || v != 9 {
		t.Fatalf("Last = (%d, %v), want (9, true)", v, ok)
	}
	if v, ok := At(seqOf([]int{7, 8, 9}), 1); !ok || v != 8 {
		t.Fatalf("At(1) = (%d, %v), want (8, true)", v, ok)
	}
	if _, ok := At(seqOf([]int{7, 8, 9}), 9); ok {
		t.Fatalf("At(9) ok = true, want false")
	}
	if c := Count(seqOf([]int{1, 2, 3})); c != 3 {
		t.Fatalf("Count = %d, want 3", c)
	}
}

func TestUtil_EnumerateChainMerge(t *testing.T) {
	got := ToSlice(Enumerate(seqOf([]string{"a", "b"})))
	want := []types.Pair[int, string]{types.PairOf(0, "a"), types.PairOf(1, "b")}
	if len(got) != len(want) {
		t.Fatalf("Enumerate len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].First != want[i].First || got[i].Second != want[i].Second {
			t.Fatalf("Enumerate[%d] = %v, want %v", i, got[i], want[i])
		}
	}
	if got := ToSlice(Chain(seqOf([]int{1}), seqOf([]int{2, 3}))); !eq(got, []int{1, 2, 3}) {
		t.Fatalf("Chain = %v, want [1 2 3]", got)
	}
	if got := ToSlice(Merge(seqOf([]int{1}), seqOf([]int{2}))); !eq(got, []int{1, 2}) {
		t.Fatalf("Merge = %v, want [1 2]", got)
	}
}

func TestUtil_Operator(t *testing.T) {
	if o := Operator(seqOf([]int{1, 2, 3, 4}), types.BinaryOperator[int](func(a, b int) int { return a + b })); o != 10 {
		t.Fatalf("Operator = %d, want 10", o)
	}
	if o := OperatorBy(seqOf([]int{1, 2, 3}), types.BinaryOperator[int](func(a, b int) int { return a + b })); o != 6 {
		t.Fatalf("OperatorBy = %d, want 6", o)
	}
}

func TestUtil_IsEmptyContains(t *testing.T) {
	if !IsEmpty(seqOf([]int{})) {
		t.Fatal("IsEmpty(empty) = false, want true")
	}
	if IsEmpty(seqOf([]int{1})) {
		t.Fatal("IsEmpty(non-empty) = true, want false")
	}
	if !IsNotEmpty(seqOf([]int{1})) {
		t.Fatal("IsNotEmpty(empty) = false, want true")
	}
	if !Contains(seqOf([]int{1, 2, 3}), 2) {
		t.Fatal("Contains 2 = false, want true")
	}
	if Contains(seqOf([]int{1, 2, 3}), 9) {
		t.Fatal("Contains 9 = true, want false")
	}
}

func TestUtil_UnzipToMap(t *testing.T) {
	pairs := seqOf([]types.Pair[string, int]{
		types.PairOf("a", 1), types.PairOf("b", 2),
	})
	a, b := Unzip(pairs)
	if !eq(a, []string{"a", "b"}) || !eq(b, []int{1, 2}) {
		t.Fatalf("Unzip = (%v, %v), want ([a b], [1 2])", a, b)
	}
	m := ToMap(seqOf([]types.Pair[string, int]{types.PairOf("x", 10)}))
	if m["x"] != 10 {
		t.Fatalf("ToMap = %v, want map[x:10]", m)
	}
}

func TestUtil_JoinBy(t *testing.T) {
	got := JoinBy(seqOf([]int{1, 2, 3}), func(v int) string { return itoa(v) }, ",")
	if got != "1,2,3" {
		t.Fatalf("JoinBy = %q, want %q", got, "1,2,3")
	}
}

func TestUtil_Seq2Conversions(t *testing.T) {
	s2 := iter.Seq2[int, string](func(yield func(int, string) bool) {
		yield(1, "a")
		yield(2, "b")
	})
	pairs := ToSlice(Seq2ToSeq(s2))
	if len(pairs) != 2 || pairs[0].First != 1 || pairs[1].Second != "b" {
		t.Fatalf("Seq2ToSeq = %v, want [(1,a) (2,b)]", pairs)
	}
	if got := ToSlice(Seq2Keys(s2)); !eq(got, []int{1, 2}) {
		t.Fatalf("Seq2Keys = %v, want [1 2]", got)
	}
	if got := ToSlice(Seq2Values(s2)); !eq(got, []string{"a", "b"}) {
		t.Fatalf("Seq2Values = %v, want [a b]", got)
	}
}

func TestUtil_Collect(t *testing.T) {
	// 用 gox/container.Collector 收集为切片
	col := sliceCollector[int]{}
	got := Collect(seqOf([]int{1, 2, 3}), col)
	if !eq(got, []int{1, 2, 3}) {
		t.Fatalf("Collect = %v, want [1 2 3]", got)
	}
}

// sliceCollector 是 container.Collector 的一个最小实现，收集为 []T。
// S 用 *[]T 以满足 Append 无返回值的接口约束。
type sliceCollector[T any] struct{}

func (sliceCollector[T]) Builder() *[]T { s := new([]T); return s }
func (sliceCollector[T]) Append(b *[]T, e T) { *b = append(*b, e) }
func (sliceCollector[T]) Finish(b *[]T) []T { return *b }

func TestUtil_SeqToSeq2(t *testing.T) {
	s2 := SeqToSeq2(seqOf([]int{1, 2, 3}))
	var got []types.Pair[int, int]
	s2(func(k, v int) bool {
		got = append(got, types.PairOf(k, v))
		return true
	})
	if len(got) != 3 || got[0].First != 0 || got[2].Second != 3 {
		t.Fatalf("SeqToSeq2 = %v, want [(0,1) (1,2) (2,3)]", got)
	}
}

// ---------------------------------------------------------------------------
// file: math.go — 聚合/数学函数
// ---------------------------------------------------------------------------

func TestMath_Sum(t *testing.T) {
	if Sum(seqOf([]int{1, 2, 3, 4})) != 10 {
		t.Fatal("Sum = wrong")
	}
	if SumComparable(seqOf([]float64{1.5, 2.5})) != 4.0 {
		t.Fatal("SumComparable = wrong")
	}
	if s, c := SumCount(seqOf([]int{1, 2, 3})); s != 6 || c != 3 {
		t.Fatalf("SumCount = (%d, %d), want (6, 3)", s, c)
	}
	if Product(seqOf([]int{2, 3, 4})) != 24 {
		t.Fatal("Product = wrong")
	}
	if Average(seqOf([]int{2, 4, 6})) != 4 {
		t.Fatal("Average = wrong")
	}
	if Mean(seqOf([]int{2, 4, 6})) != 4.0 {
		t.Fatal("Mean = wrong")
	}
}

func TestMath_MaxMin(t *testing.T) {
	maxV, ok := Max(seqOf([]int{1, 5, 3}))
	if !ok || maxV != 5 {
		t.Fatalf("Max = (%d, %v), want (5, true)", maxV, ok)
	}
	minV, ok := Min(seqOf([]int{1, 5, 3}))
	if !ok || minV != 1 {
		t.Fatalf("Min = (%d, %v), want (1, true)", minV, ok)
	}
	if _, ok := Max(seqOf([]int{})); ok {
		t.Fatal("Max empty ok = true, want false")
	}
}

func TestMath_MaxBy(t *testing.T) {
	maxBy, ok := MaxBy(seqOf([]string{"a", "bb", "ccc"}), gcmp.LessFunc[string](func(a, b string) bool { return len(a) < len(b) }))
	if !ok || maxBy != "ccc" {
		t.Fatalf("MaxBy = (%q, %v), want (ccc, true)", maxBy, ok)
	}
}

// ---------------------------------------------------------------------------
// file: std.go — 从具体容器构造 Stream 的构造函数
// ---------------------------------------------------------------------------

func TestStd_SliceAll(t *testing.T) {
	if got := SliceAllValues([]int{10, 20}).Collect(); !eq(got, []int{10, 20}) {
		t.Fatalf("SliceAllValues = %v, want [10 20]", got)
	}
	idx := SliceAll([]int{10, 20}).Collect()
	if idx[0].First != 0 || idx[1].Second != 20 {
		t.Fatalf("SliceAll = %v, want [(0,10) (1,20)]", idx)
	}
}

func TestStd_SliceBackward(t *testing.T) {
	if got := SliceBackwardValues([]int{1, 2, 3}).Collect(); !eq(got, []int{3, 2, 1}) {
		t.Fatalf("SliceBackwardValues = %v, want [3 2 1]", got)
	}
	back := SliceBackward([]int{10, 20, 30}).Collect()
	if back[0].First != 2 || back[2].Second != 10 {
		t.Fatalf("SliceBackward = %v, want [(2,30) (1,20) (0,10)]", back)
	}
}

func TestStd_HashMapAll(t *testing.T) {
	pairs := HashMapAll(map[string]int{"a": 1, "b": 2}).Collect()
	if len(pairs) != 2 {
		t.Fatalf("HashMapAll len = %d, want 2", len(pairs))
	}
}

func TestStd_StringAll(t *testing.T) {
	pairs := StringAll("ab").Collect()
	if len(pairs) != 2 || pairs[0].First != 0 || pairs[0].Second != 'a' || pairs[1].Second != 'b' {
		t.Fatalf("StringAll = %v, want [(0,a) (1,b)]", pairs)
	}
	s2 := ToSlice(Seq2ToSeq(StringAll2("ab")))
	if len(s2) != 2 || s2[0].First != 0 || s2[1].Second != 'b' {
		t.Fatalf("StringAll2 = %v, want [(0,a) (1,b)]", s2)
	}
	if got := ToSlice(StringRunes("你").Seq()); got[0] != '你' {
		t.Fatalf("StringRunes = %v, want [你]", got)
	}
}

func TestStd_ChannelAll(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)
	if got := ToSlice(ChannelAll(ch).Seq()); !eq(got, []int{1, 2, 3}) {
		t.Fatalf("ChannelAll = %v, want [1 2 3]", got)
	}

	ch2 := make(chan types.Pair[int, string], 2)
	ch2 <- types.PairOf(1, "a")
	ch2 <- types.PairOf(2, "b")
	close(ch2)
	got := ToSlice(Seq2ToSeq(ChannelAll2(ch2)))
	if len(got) != 2 || got[1].Second.Second != "b" {
		t.Fatalf("ChannelAll2 = %v, want [(0,(1,a)) (1,(2,b))]", got)
	}
}

func TestStd_RangeAll(t *testing.T) {
	if got := RangeAll(0, 6, 2).Collect(); !eq(got, []int{0, 2, 4, 6}) {
		t.Fatalf("RangeAll = %v, want [0 2 4 6]", got)
	}
	r2 := ToSlice(Seq2ToSeq(RangeAll2(0, 4, 2)))
	if len(r2) != 3 || r2[0].First != 0 || r2[2].Second != 4 {
		t.Fatalf("RangeAll2 = %v, want [(0,0) (2,2) (4,4)]", r2)
	}
}

// ---------------------------------------------------------------------------
// file: stream.go — Stream 类型与方法（含 go1.27 方法级泛型）
// ---------------------------------------------------------------------------

func TestStream_NewStreamAndSeq(t *testing.T) {
	s := NewStream(seqOf([]int{1, 2, 3}))
	// Seq() 转回 iter.Seq
	var got []int
	for v := range s.Seq() {
		got = append(got, v)
	}
	if !eq(got, []int{1, 2, 3}) {
		t.Fatalf("Stream.Seq() = %v, want [1 2 3]", got)
	}
}

func TestStream_MapGeneric(t *testing.T) {
	// go1.27 方法级泛型：Map 返回不同类型 R
	strs := NewStream(seqOf([]int{1, 2, 3})).
		Map(types.UnaryFunction[int, string](func(v int) string { return itoa(v) })).
		Collect()
	if !eq(strs, []string{"1", "2", "3"}) {
		t.Fatalf("Stream.Map(int->string) = %v, want [1 2 3]", strs)
	}
}

func TestStream_FlatMapGeneric(t *testing.T) {
	flat := NewStream(seqOf([]int{1, 2})).
		FlatMap(types.UnaryFunction[int, iter.Seq[int]](func(v int) iter.Seq[int] {
			return seqOf([]int{v, v * 10})
		})).
		Collect()
	if !eq(flat, []int{1, 10, 2, 20}) {
		t.Fatalf("Stream.FlatMap = %v, want [1 10 2 20]", flat)
	}
}

func TestStream_ChainOps(t *testing.T) {
	// Filter -> Map(int->string) -> Collect
	chained := NewStream(seqOf([]int{1, 2, 3, 4})).
		Filter(types.Predicate[int](func(v int) bool { return v%2 == 0 })).
		Map(types.UnaryFunction[int, string](func(v int) string { return itoa(v * 10) })).
		Collect()
	if !eq(chained, []string{"20", "40"}) {
		t.Fatalf("Stream chain = %v, want [20 40]", chained)
	}

	// Distinct / Sorted / Limit / Skip / Until / Peek / ForEach
	sorted := NewStream(seqOf([]int{3, 1, 2, 1})).
		Distinct(types.UnaryFunction[int, int](func(v int) int { return v })).
		Sorted(types.Comparator[int](cmp.Compare[int])).
		Collect()
	if !eq(sorted, []int{1, 2, 3}) {
		t.Fatalf("Stream Distinct+Sorted = %v, want [1 2 3]", sorted)
	}

	limited := NewStream(seqOf([]int{1, 2, 3, 4, 5})).
		Skip(1).Limit(2).Collect()
	if !eq(limited, []int{2, 3}) {
		t.Fatalf("Stream Skip+Limit = %v, want [2 3]", limited)
	}

	var peeked []int
	until := NewStream(seqOf([]int{1, 2, 3, 4})).
		Peek(types.Consumer[int](func(v int) { peeked = append(peeked, v) })).
		Until(types.Predicate[int](func(v int) bool { return v == 3 })).
		Collect()
	if !eq(until, []int{1, 2}) {
		t.Fatalf("Stream Peek+Until = %v, want [1 2]", until)
	}
	if !eq(peeked, []int{1, 2, 3}) {
		t.Fatalf("Stream Peek seen = %v, want [1 2 3]", peeked)
	}

	// terminal ops
	if !NewStream(seqOf([]int{2, 4, 6})).All(types.Predicate[int](func(v int) bool { return v%2 == 0 })) {
		t.Fatal("Stream.All = false, want true")
	}
	if NewStream(seqOf([]int{2, 4, 6})).Any(types.Predicate[int](func(v int) bool { return v%2 != 0 })) {
		t.Fatal("Stream.Any = true, want false")
	}
	if v, ok := NewStream(seqOf([]int{5, 3, 1})).First(); !ok || v != 5 {
		t.Fatalf("Stream.First = (%d, %v), want (5, true)", v, ok)
	}
	if NewStream(seqOf([]int{5, 3, 1})).Count() != 3 {
		t.Fatal("Stream.Count = wrong")
	}
	if NewStream(seqOf([]int{2, 4, 6})).Sum(types.BinaryOperator[int](func(a, b int) int { return a + b })) != 12 {
		t.Fatal("Stream.Sum = wrong")
	}
	var sum int
	NewStream(seqOf([]int{1, 2, 3})).ForEach(types.Consumer[int](func(v int) { sum += v }))
	if sum != 6 {
		t.Fatalf("Stream.ForEach sum = %d, want 6", sum)
	}
	if r := NewStream(seqOf([]int{1, 2, 3})).Fold(100, types.BinaryOperator[int](func(a, b int) int { return a + b })); r != 106 {
		t.Fatalf("Stream.Fold = %d, want 106", r)
	}
	if r, ok := NewStream(seqOf([]int{1, 2, 3})).Reduce(types.BinaryOperator[int](func(a, b int) int { return a + b })); !ok || r != 6 {
		t.Fatalf("Stream.Reduce = (%d, %v), want (6, true)", r, ok)
	}
}

func TestStream_Iter(t *testing.T) {
	it := NewStream(seqOf([]int{1, 2, 3})).Iter()
	if g, ok := it.(GoIter[int]); ok {
		defer g.Stop()
	}
	var got []int
	for {
		v, ok := it.Next()
		if !ok {
			break
		}
		got = append(got, v)
	}
	if !eq(got, []int{1, 2, 3}) {
		t.Fatalf("Stream.Iter() = %v, want [1 2 3]", got)
	}
}

// ---------------------------------------------------------------------------
// file: iter.go — Iterator / Iterable / Seq 互转
// ---------------------------------------------------------------------------

func TestIter_SeqIterRoundTrip(t *testing.T) {
	seq := seqOf([]int{1, 2, 3})
	it := SeqIter(seq)
	if g, ok := it.(GoIter[int]); ok {
		defer g.Stop()
	}
	var got []int
	for {
		v, ok := it.Next()
		if !ok {
			break
		}
		got = append(got, v)
	}
	if !eq(got, []int{1, 2, 3}) {
		t.Fatalf("SeqIter round trip = %v, want [1 2 3]", got)
	}
}

func TestIter_IterSeq(t *testing.T) {
	seq := seqOf([]int{1, 2, 3})
	it := SeqIter(seq)
	if g, ok := it.(GoIter[int]); ok {
		defer g.Stop()
	}
	back := IterSeq(it)
	if got := ToSlice(back); !eq(got, []int{1, 2, 3}) {
		t.Fatalf("IterSeq = %v, want [1 2 3]", got)
	}
}

// myIterable 是 Iterable 接口的一个最小实现，用于测试 Iterable 约束的函数。
type myIterable[T any] struct{ s []T }

func (m myIterable[T]) Iter() Iterator[T] { return SeqIter(seqOf(m.s)) }

func TestIter_Iterable(t *testing.T) {
	var _ Iterable[int] = myIterable[int]{}
	it := myIterable[int]{s: []int{4, 5, 6}}.Iter()
	if g, ok := it.(GoIter[int]); ok {
		defer g.Stop()
	}
	var got []int
	for {
		v, ok := it.Next()
		if !ok {
			break
		}
		got = append(got, v)
	}
	if !eq(got, []int{4, 5, 6}) {
		t.Fatalf("Iterable.Iter() = %v, want [4 5 6]", got)
	}
	// SeqSeq2 已在 TestUtil_SeqSeq2 覆盖
	_ = gcontainer.Collector[[]int, int, []int](nil)
}

// ---------------------------------------------------------------------------
// file: util.go (追加) — Zip/GroupBy/Partition/Chunk/Window/TakeWhile/DropWhile/
//                       NoneMatch/ForEachIndexed/Find/FindLast/Concat
// ---------------------------------------------------------------------------

func TestUtil_Zip(t *testing.T) {
	got := ToSlice(Zip(seqOf([]int{1, 2, 3}), seqOf([]string{"a", "b"})))
	if len(got) != 2 {
		t.Fatalf("Zip len = %d, want 2", len(got))
	}
	if got[0].First != 1 || got[0].Second != "a" || got[1].Second != "b" {
		t.Fatalf("Zip = %v, want [(1,a) (2,b)]", got)
	}
}

func TestUtil_GroupBy(t *testing.T) {
	m := GroupBy(seqOf([]int{1, 2, 3, 4, 5}), types.UnaryFunction[int, bool](func(v int) bool { return v%2 == 0 }))
	if len(m[true]) != 2 || len(m[false]) != 3 {
		t.Fatalf("GroupBy = %v, want even:2 odd:3", m)
	}
}

func TestUtil_Partition(t *testing.T) {
	yes, no := Partition(seqOf([]int{1, 2, 3, 4}), types.Predicate[int](func(v int) bool { return v%2 == 0 }))
	if !eq(yes, []int{2, 4}) || !eq(no, []int{1, 3}) {
		t.Fatalf("Partition = (%v, %v), want ([2 4], [1 3])", yes, no)
	}
}

func TestUtil_Chunk(t *testing.T) {
	got := ToSlice(Chunk(seqOf([]int{1, 2, 3, 4, 5}), 2))
	if len(got) != 3 || !eq(got[0], []int{1, 2}) || !eq(got[2], []int{5}) {
		t.Fatalf("Chunk = %v, want [[1 2] [3 4] [5]]", got)
	}
	if got := ToSlice(Chunk(seqOf([]int{1, 2}), 0)); len(got) != 0 {
		t.Fatalf("Chunk(0) len = %d, want 0", len(got))
	}
}

func TestUtil_Window(t *testing.T) {
	got := ToSlice(Window(seqOf([]int{1, 2, 3, 4}), 2))
	want := [][]int{{1, 2}, {2, 3}, {3, 4}}
	if len(got) != len(want) {
		t.Fatalf("Window len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if !eq(got[i], want[i]) {
			t.Fatalf("Window[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestUtil_TakeWhileDropWhile(t *testing.T) {
	if got := ToSlice(TakeWhile(seqOf([]int{1, 2, 3, 1}), types.Predicate[int](func(v int) bool { return v < 3 }))); !eq(got, []int{1, 2}) {
		t.Fatalf("TakeWhile = %v, want [1 2]", got)
	}
	if got := ToSlice(DropWhile(seqOf([]int{1, 2, 3, 1}), types.Predicate[int](func(v int) bool { return v < 3 }))); !eq(got, []int{3, 1}) {
		t.Fatalf("DropWhile = %v, want [3 1]", got)
	}
}

func TestUtil_NoneMatch(t *testing.T) {
	if !NoneMatch(seqOf([]int{1, 3, 5}), types.Predicate[int](func(v int) bool { return v%2 == 0 })) {
		t.Fatal("NoneMatch even = false, want true")
	}
	if NoneMatch(seqOf([]int{1, 2, 3}), types.Predicate[int](func(v int) bool { return v%2 == 0 })) {
		t.Fatal("NoneMatch even = true, want false")
	}
}

func TestUtil_ForEachIndexed(t *testing.T) {
	var idxs []int
	var vals []int
	ForEachIndexed(seqOf([]int{10, 20, 30}), func(i, v int) {
		idxs = append(idxs, i)
		vals = append(vals, v)
	})
	if !eq(idxs, []int{0, 1, 2}) || !eq(vals, []int{10, 20, 30}) {
		t.Fatalf("ForEachIndexed = (%v, %v), want ([0 1 2], [10 20 30])", idxs, vals)
	}
}

func TestUtil_FindFindLast(t *testing.T) {
	v, ok := Find(seqOf([]int{1, 2, 3, 4}), types.Predicate[int](func(v int) bool { return v > 2 }))
	if !ok || v != 3 {
		t.Fatalf("Find = (%d, %v), want (3, true)", v, ok)
	}
	if _, ok := Find(seqOf([]int{1, 2}), types.Predicate[int](func(v int) bool { return v > 9 })); ok {
		t.Fatal("Find missing ok = true, want false")
	}
	v2, ok := FindLast(seqOf([]int{1, 2, 3, 4}), types.Predicate[int](func(v int) bool { return v%2 == 0 }))
	if !ok || v2 != 4 {
		t.Fatalf("FindLast = (%d, %v), want (4, true)", v2, ok)
	}
}

func TestUtil_Concat(t *testing.T) {
	if got := ToSlice(Concat(seqOf([]int{1, 2}), seqOf([]int{3}))); !eq(got, []int{1, 2, 3}) {
		t.Fatalf("Concat = %v, want [1 2 3]", got)
	}
}

// ---------------------------------------------------------------------------
// file: math.go (追加) — MinBy / MaxByComparable / MinByComparable / Median / StdDev
// ---------------------------------------------------------------------------

func TestMath_MinMaxByComparable(t *testing.T) {
	maxV, ok := MaxByComparable(seqOf([]int{1, 5, 3}))
	if !ok || maxV != 5 {
		t.Fatalf("MaxByComparable = (%d, %v), want (5, true)", maxV, ok)
	}
	minV, ok := MinByComparable(seqOf([]int{1, 5, 3}))
	if !ok || minV != 1 {
		t.Fatalf("MinByComparable = (%d, %v), want (1, true)", minV, ok)
	}
}

func TestMath_MinBy(t *testing.T) {
	minBy, ok := MinBy(seqOf([]string{"a", "bb", "ccc"}), gcmp.LessFunc[string](func(a, b string) bool { return len(a) < len(b) }))
	if !ok || minBy != "a" {
		t.Fatalf("MinBy = (%q, %v), want (a, true)", minBy, ok)
	}
}

func TestMath_MedianStdDev(t *testing.T) {
	med, ok := Median(seqOf([]int{3, 1, 2}))
	if !ok || med != 2 {
		t.Fatalf("Median(3,1,2) = (%d, %v), want (2, true)", med, ok)
	}
	med2, _ := Median(seqOf([]int{4, 1, 3, 2}))
	if med2 != 2 {
		t.Fatalf("Median(4,1,3,2) = %d, want 2", med2)
	}
	sd, ok := StdDev(seqOf([]float64{2, 4, 4, 4, 5, 5, 7, 9}))
	if !ok {
		t.Fatal("StdDev ok = false, want true")
	}
	// population stddev of that set is 2
	if sd < 1.999 || sd > 2.001 {
		t.Fatalf("StdDev = %f, want ~2.0", sd)
	}
}

// ---------------------------------------------------------------------------
// file: stream.go (追加) — 新方法级泛型方法
// ---------------------------------------------------------------------------

func TestStream_ZipTakeWhileDropWhile(t *testing.T) {
	zipped := ToSlice(NewStream(seqOf([]int{1, 2, 3})).Zip(seqOf([]string{"a", "b"})))
	if len(zipped) != 2 || zipped[0].First != 1 || zipped[1].Second != "b" {
		t.Fatalf("Stream.Zip = %v, want [(1,a) (2,b)]", zipped)
	}

	tw := NewStream(seqOf([]int{1, 2, 3, 1})).TakeWhile(types.Predicate[int](func(v int) bool { return v < 3 })).Collect()
	if !eq(tw, []int{1, 2}) {
		t.Fatalf("Stream.TakeWhile = %v, want [1 2]", tw)
	}
	dw := NewStream(seqOf([]int{1, 2, 3, 1})).DropWhile(types.Predicate[int](func(v int) bool { return v < 3 })).Collect()
	if !eq(dw, []int{3, 1}) {
		t.Fatalf("Stream.DropWhile = %v, want [3 1]", dw)
	}
}

func TestStream_ChunkWindow(t *testing.T) {
	chunks := ToSlice(NewStream(seqOf([]int{1, 2, 3, 4, 5})).Chunk(2))
	if len(chunks) != 3 || !eq(chunks[0], []int{1, 2}) || !eq(chunks[2], []int{5}) {
		t.Fatalf("Stream.Chunk = %v, want [[1 2] [3 4] [5]]", chunks)
	}
	windows := ToSlice(NewStream(seqOf([]int{1, 2, 3, 4})).Window(2))
	if len(windows) != 3 || !eq(windows[0], []int{1, 2}) || !eq(windows[2], []int{3, 4}) {
		t.Fatalf("Stream.Window = %v, want [[1 2] [2 3] [3 4]]", windows)
	}
}

func TestStream_GroupByPartition(t *testing.T) {
	m := NewStream(seqOf([]int{1, 2, 3, 4, 5})).GroupBy(types.UnaryFunction[int, bool](func(v int) bool { return v%2 == 0 }))
	if len(m[true]) != 2 || len(m[false]) != 3 {
		t.Fatalf("Stream.GroupBy = %v, want even:2 odd:3", m)
	}
	yes, no := NewStream(seqOf([]int{1, 2, 3, 4})).Partition(types.Predicate[int](func(v int) bool { return v%2 == 0 }))
	if !eq(yes, []int{2, 4}) || !eq(no, []int{1, 3}) {
		t.Fatalf("Stream.Partition = (%v, %v), want ([2 4], [1 3])", yes, no)
	}
}

func TestStream_NoneMatchForEachIndexedFind(t *testing.T) {
	if !NewStream(seqOf([]int{1, 3, 5})).NoneMatch(types.Predicate[int](func(v int) bool { return v%2 == 0 })) {
		t.Fatal("Stream.NoneMatch = false, want true")
	}
	var idxs []int
	var vals []int
	NewStream(seqOf([]int{10, 20})).ForEachIndexed(func(i, v int) {
		idxs = append(idxs, i)
		vals = append(vals, v)
	})
	if !eq(idxs, []int{0, 1}) || !eq(vals, []int{10, 20}) {
		t.Fatalf("Stream.ForEachIndexed = (%v, %v), want ([0 1], [10 20])", idxs, vals)
	}
	if v, ok := NewStream(seqOf([]int{1, 2, 3, 4})).Find(types.Predicate[int](func(v int) bool { return v > 2 })); !ok || v != 3 {
		t.Fatalf("Stream.Find = (%d, %v), want (3, true)", v, ok)
	}
	if v, ok := NewStream(seqOf([]int{1, 2, 3, 4})).FindLast(types.Predicate[int](func(v int) bool { return v%2 == 0 })); !ok || v != 4 {
		t.Fatalf("Stream.FindLast = (%d, %v), want (4, true)", v, ok)
	}
}

func TestStream_MinMaxBy(t *testing.T) {
	maxBy, ok := NewStream(seqOf([]string{"a", "bb", "ccc"})).MaxBy(gcmp.LessFunc[string](func(a, b string) bool { return len(a) < len(b) }))
	if !ok || maxBy != "ccc" {
		t.Fatalf("Stream.MaxBy = (%q, %v), want (ccc, true)", maxBy, ok)
	}
	minBy, ok := NewStream(seqOf([]string{"a", "bb", "ccc"})).MinBy(gcmp.LessFunc[string](func(a, b string) bool { return len(a) < len(b) }))
	if !ok || minBy != "a" {
		t.Fatalf("Stream.MinBy = (%q, %v), want (a, true)", minBy, ok)
	}
}

// ---------------------------------------------------------------------------
// file: iter.go (追加) — 基础构造器 Empty/Once/Repeat/RepeatN/Cycle
// ---------------------------------------------------------------------------

func TestIter_Constructors(t *testing.T) {
	if got := ToSlice(Empty[int]()); len(got) != 0 {
		t.Fatalf("Empty len = %d, want 0", len(got))
	}
	if got := ToSlice(Once(7)); !eq(got, []int{7}) {
		t.Fatalf("Once = %v, want [7]", got)
	}
	if got := ToSlice(RepeatN(9, 3)); !eq(got, []int{9, 9, 9}) {
		t.Fatalf("RepeatN = %v, want [9 9 9]", got)
	}
	// Repeat / Cycle 是无限流，必须用 Limit 截断
	if got := ToSlice(Limit(Repeat(5), 3)); !eq(got, []int{5, 5, 5}) {
		t.Fatalf("Repeat+Limit = %v, want [5 5 5]", got)
	}
	if got := ToSlice(Limit(Cycle(seqOf([]int{1, 2})), 5)); !eq(got, []int{1, 2, 1, 2, 1}) {
		t.Fatalf("Cycle+Limit = %v, want [1 2 1 2 1]", got)
	}
}
