package augmentedtree

import "fmt"

type PartialOrdered[T any] interface {
	LessThan(T) bool
	LessOrEqualThan(T) bool
}

type Point[T any] interface {
	comparable
	PartialOrdered[T]
}

type point[P Point[P]] struct{ inner P }

//go:inline
func (p point[P]) lessThan(rhs point[P]) bool { return p.inner.LessThan(rhs.inner) }

//go:inline
func (p point[P]) lessOrEqualThan(rhs point[P]) bool { return p.inner.LessOrEqualThan(rhs.inner) }

//go:inline
func (p point[P]) greaterThan(rhs point[P]) bool { return rhs.inner.LessThan(p.inner) }

//go:inline
func (p point[P]) greaterOrEqualThan(rhs point[P]) bool {
	return rhs.inner.LessOrEqualThan(p.inner)
}

type Interval[P Point[P]] struct{ low, high point[P] }

//go:inline
func (iv *Interval[P]) overlaps(rhs *Interval[P]) bool {
	return iv.high.greaterOrEqualThan(rhs.low) && iv.low.lessOrEqualThan(rhs.high)
}

//go:inline
func (iv *Interval[P]) contains(rhs *Interval[P]) bool {
	return iv.low.lessOrEqualThan(rhs.low) && rhs.high.lessOrEqualThan(iv.high)
}

//go:inline
func (iv *Interval[P]) Low() P { return iv.low.inner }

//go:inline
func (iv *Interval[P]) High() P { return iv.high.inner }

//go:inline
func (iv *Interval[P]) String() string {
	return fmt.Sprintf("%v~%v", iv.low.inner, iv.high.inner)
}

func NewInterval[P Point[P]](low, high P) Interval[P] {
	if !low.LessOrEqualThan(high) {
		low, high = high, low
	}
	return Interval[P]{point[P]{low}, point[P]{high}}
}
