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

func (p point[P]) lessThan(rhs point[P]) bool { return p.inner.LessThan(rhs.inner) }

func (p point[P]) lessOrEqualThan(rhs point[P]) bool { return p.inner.LessOrEqualThan(rhs.inner) }

func (p point[P]) greaterThan(rhs point[P]) bool { return rhs.inner.LessThan(p.inner) }

func (p point[P]) greaterOrEqualThan(rhs point[P]) bool {
	return rhs.inner.LessOrEqualThan(p.inner)
}

func (p point[P]) compare(rhs point[P]) int {
	if p.lessThan(rhs) {
		return -1
	} else if p.greaterThan(rhs) {
		return 1
	}
	return 0
}

type Interval[P Point[P]] struct{ low, high point[P] }

func (iv *Interval[P]) contains(rhs *Interval[P]) bool {
	return iv.low.lessOrEqualThan(rhs.low) && rhs.high.lessOrEqualThan(iv.high)
}

func (iv *Interval[P]) hasIntersection(rhs *Interval[P]) bool {
	lower, higher := iv, rhs
	if lower.low.greaterThan(higher.low) {
		lower, higher = higher, lower
	}
	return lower.high.greaterOrEqualThan(higher.low)
}

func (iv *Interval[P]) Low() P { return iv.low.inner }

func (iv *Interval[P]) High() P { return iv.high.inner }

func (iv *Interval[P]) String() string {
	return fmt.Sprintf("%v~%v", iv.low.inner, iv.high.inner)
}

func NewInterval[P Point[P]](low, high P) Interval[P] {
	if !low.LessOrEqualThan(high) {
		low, high = high, low
	}
	return Interval[P]{point[P]{low}, point[P]{high}}
}
