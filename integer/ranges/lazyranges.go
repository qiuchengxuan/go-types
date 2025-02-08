package ranges

import "golang.org/x/exp/constraints"

type LazyRanges[I constraints.Integer] struct{ left, right Ranges[I] }

func (r LazyRanges[I]) Eval() Ranges[I] {
	return r.left.Sub(r.right)
}

func (r LazyRanges[I]) Len() uint64 {
	return r.Eval().Len()
}

func (r LazyRanges[I]) Foreach(consumer func(x I)) {
	r.Eval().Foreach(consumer)
}

func LazySub[I constraints.Integer](left, right Ranges[I]) LazyRanges[I] {
	return LazyRanges[I]{left, right}
}
