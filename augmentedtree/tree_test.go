package augmentedtree

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
)

type intPoint[T constraints.Integer] struct{ t T }

func (p intPoint[T]) LessThan(rhs intPoint[T]) bool {
	return p.t < rhs.t
}

func (p intPoint[T]) LessOrEqualThan(rhs intPoint[T]) bool {
	return p.t <= rhs.t
}

func (p intPoint[T]) String() string {
	return strconv.Itoa(int(int64(p.t)))
}

func newInterval(low, high int) Interval[intPoint[int]] {
	return NewInterval(intPoint[int]{low}, intPoint[int]{high})
}

func TestAugmentedTree(t *testing.T) {
	tree := Tree[intPoint[int], struct{}]{}
	tree.Put(newInterval(1, 2), struct{}{})
	tree.Put(newInterval(2, 3), struct{}{})
	assert.Equal(t, newInterval(1, 3), tree.root.minMax)
	tree.Put(newInterval(0, 1), struct{}{})
	assert.Equal(t, newInterval(0, 3), tree.root.minMax)
	assert.Empty(t, tree.QueryAll(newInterval(-1, -1)))
	assert.NotEmpty(t, tree.QueryAll(newInterval(0, 0)))
	assert.Equal(t, newInterval(1, 2), tree.QueryAll(newInterval(2, 2))[0].interval)
	assert.NotEmpty(t, tree.QueryAll(newInterval(3, 3)))
	assert.NotEmpty(t, tree.QueryAll(newInterval(0, 4)))
	assert.Empty(t, tree.QueryAll(newInterval(4, 4)))
	tree.Put(newInterval(1, 3), struct{}{})
	assert.Equal(t, 3, len(tree.QueryAll(newInterval(2, 2))))

	tree = Tree[intPoint[int], struct{}]{}
	for i := range 10 {
		tree.Put(newInterval(i, i), struct{}{})
	}
	for i := range 10 {
		assert.True(t, tree.Delete(newInterval(i, i)))
	}
}

func TestAugmentedTreeRotationLeft(t *testing.T) {
	tree := Tree[intPoint[int], struct{}]{}
	tree.Put(newInterval(1, 2), struct{}{})
	tree.Put(newInterval(3, 4), struct{}{})
	tree.Put(newInterval(5, 6), struct{}{})
	assert.Equal(t, newInterval(1, 6), tree.root.minMax)
}

func TestAugmentedTreeRotationRight(t *testing.T) {
	tree := Tree[intPoint[int], struct{}]{}
	tree.Put(newInterval(5, 6), struct{}{})
	tree.Put(newInterval(3, 4), struct{}{})
	tree.Put(newInterval(1, 2), struct{}{})
	assert.Equal(t, newInterval(1, 6), tree.root.minMax)
}

func TestAugmentedTreeRotationLeftRight(t *testing.T) {
	tree := Tree[intPoint[int], struct{}]{}
	tree.Put(newInterval(1, 2), struct{}{})
	tree.Put(newInterval(5, 6), struct{}{})
	tree.Put(newInterval(3, 4), struct{}{})
	assert.Equal(t, newInterval(1, 6), tree.root.minMax)
}

func TestAugmentedTreeRotationRightLeft(t *testing.T) {
	tree := Tree[intPoint[int], struct{}]{}
	tree.Put(newInterval(5, 6), struct{}{})
	tree.Put(newInterval(1, 2), struct{}{})
	tree.Put(newInterval(3, 4), struct{}{})
	assert.Equal(t, newInterval(1, 6), tree.root.minMax)
}

func TestAugmentedTreeDeleteNonExist(t *testing.T) {
	tree := Tree[intPoint[int], struct{}]{}
	assert.False(t, tree.Delete(newInterval(1, 2)))
	tree.Put(newInterval(1, 6), struct{}{})
	assert.False(t, tree.Delete(newInterval(1, 2)))
}

func TestAugmentedTreeDeleteRoot(t *testing.T) {
	tree := Tree[intPoint[int], struct{}]{}
	tree.Put(newInterval(1, 6), struct{}{})
	assert.True(t, tree.Delete(newInterval(1, 6)))
	tree.Put(newInterval(1, 2), struct{}{})
	tree.Put(newInterval(3, 4), struct{}{})
	assert.True(t, tree.Delete(newInterval(1, 2)))
	assert.Equal(t, newInterval(3, 4), tree.root.minMax)
}

func TestAugmentedTreeDeleteLeaf(t *testing.T) {
	tree := Tree[intPoint[int], struct{}]{}
	tree.Put(newInterval(1, 2), struct{}{})
	tree.Put(newInterval(3, 4), struct{}{})
	assert.True(t, tree.Delete(newInterval(3, 4)))

	tree = Tree[intPoint[int], struct{}]{}
	tree.Put(newInterval(3, 3), struct{}{})
	tree.Put(newInterval(2, 2), struct{}{})
	tree.Put(newInterval(5, 5), struct{}{})
	tree.Put(newInterval(1, 1), struct{}{})
	tree.Put(newInterval(6, 6), struct{}{})
	tree.Put(newInterval(4, 4), struct{}{})
	tree.Delete(newInterval(1, 1))
}

func TestAugmentedTreeDeleteNode(t *testing.T) {
	tree := Tree[intPoint[int], struct{}]{}
	tree.Put(newInterval(1, 2), struct{}{})
	tree.Put(newInterval(-3, -2), struct{}{})
	tree.Put(newInterval(3, 4), struct{}{})
	tree.Put(newInterval(-4, -3), struct{}{})
	tree.Put(newInterval(-2, -1), struct{}{})
	assert.True(t, tree.Delete(newInterval(1, 2)))
	assert.True(t, tree.Delete(newInterval(-2, -1)))
	assert.True(t, tree.Delete(newInterval(3, 4)))
	assert.True(t, tree.Delete(newInterval(-4, -3)))
	assert.True(t, tree.Delete(newInterval(-3, -2)))
}

func BenchmarkAugmentedTreePut(b *testing.B) {
	tree := Tree[intPoint[int], struct{}]{}
	for i := range b.N {
		tree.Put(newInterval(i, i), struct{}{})
	}
}

func BenchmarkAugmentedTreeDelete(b *testing.B) {
	tree := Tree[intPoint[int], struct{}]{}
	for i := range b.N {
		tree.Put(newInterval(i, i), struct{}{})
	}
	b.ResetTimer()
	for i := range b.N {
		tree.Delete(newInterval(i, i))
	}
}
