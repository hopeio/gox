package slices

import (
	"iter"

	"github.com/hopeio/gox/types"
)

func OrderIterBy[S ~[]T, T any](s S, cmp types.Comparator[T]) iter.Seq[T] {
	c := Copy(s)
	return func(yield func(T) bool) {
		var maxIdx int
		var max T
		for {
			for i, v := range c {
				if cmp(v, max) > 0 {
					maxIdx = i
					max = v
				}
			}
			if !yield(max) {
				return
			}
			c = append(s[:maxIdx], c[maxIdx+1:]...)
		}
	}
}
