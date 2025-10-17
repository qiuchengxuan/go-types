package augmentedtree

import "fmt"

type Entry[P Point[P], V any] struct {
	interval Interval[P]
	value    V
}

func (e *Entry[P, V]) Interval() Interval[P] { return e.interval }

func (e *Entry[P, V]) Value() V { return e.value }

type color uint8

const (
	black, red color = 0, 1
)

type direction uint8

const (
	left, right direction = 0, 1
)

func (d direction) opposite() direction { return d ^ 1 }

type node[P Point[P], V any] struct {
	entry  Entry[P, V]
	minMax Interval[P]
	parent *node[P, V]
	childs [2]*node[P, V]
	color  color
}

func (n *node[P, V]) grandParent() *node[P, V] { return n.parent.parent }

func (n *node[P, V]) leftChild() *node[P, V] { return n.childs[left] }

func (n *node[P, V]) rightChild() *node[P, V] { return n.childs[right] }

func (n *node[P, V]) is(color color) bool {
	if n == nil {
		return black == color
	}
	return n.color == color
}

func (n *node[P, V]) child(direction direction) *node[P, V] { return n.childs[direction] }

func (n *node[P, V]) direction() direction {
	if n.parent.childs[left] == n {
		return left
	}
	return right
}

func (n *node[P, V]) uncle() *node[P, V] {
	grandParent := n.parent.parent
	if grandParent.leftChild() != n.parent {
		return grandParent.leftChild()
	}
	return grandParent.rightChild()
}

func (n *node[P, V]) slibing() *node[P, V] {
	if n != n.parent.leftChild() {
		return n.parent.leftChild()
	}
	return n.parent.rightChild()
}

func (n *node[P, V]) query(interval Interval[P], fn func(*Entry[P, V]) bool) *Entry[P, V] {
	if interval.hasIntersection(&n.entry.interval) {
		if fn(&n.entry) {
			return &n.entry
		}
	}
	if left := n.leftChild(); left != nil && left.minMax.hasIntersection(&interval) {
		if retval := left.query(interval, fn); retval != nil {
			return retval
		}
	}
	if right := n.rightChild(); right != nil && right.minMax.hasIntersection(&interval) {
		if retval := right.query(interval, fn); retval != nil {
			return retval
		}
	}
	return nil
}

func (n *node[P, V]) updateMinMax() {
	low := n.entry.interval.low
	if left := n.leftChild(); left != nil {
		low = left.minMax.low
	}
	high := n.entry.interval.high
	if right := n.rightChild(); right != nil {
		high = right.minMax.high
	}
	n.minMax = Interval[P]{low, high}
}

func compareIntervals[P Point[P]](a, b Interval[P]) int {
	if cmp := a.low.compare(b.low); cmp != 0 {
		return cmp
	}
	return a.high.compare(b.high)
}

/*
 *     N            N
 *    / \          / \
 *   A   R   ==>  R   C
 *      / \      / \
 *     B   C    A   B
 *
 * Without swapping node, swap N and R value
 */
func (n *node[P, V]) rotateLeft() {
	rightChild := n.rightChild()
	n.childs[right] = rightChild.rightChild()
	rightChild.childs = [2]*node[P, V]{n.leftChild(), rightChild.leftChild()}
	n.childs[left] = rightChild
	n.color, rightChild.color = rightChild.color, n.color
	n.entry, rightChild.entry = rightChild.entry, n.entry
	if child := n.rightChild(); child != nil { // C.parent = N
		child.parent = n
	}
	if child := rightChild.leftChild(); child != nil { // A.parent = R
		child.parent = rightChild
	}
	n.leftChild().updateMinMax()
	n.updateMinMax()
}

/*
 *      N            N
 *     / \          / \
 *    L   C   ==>  A   L
 *   / \              / \
 *  A   B            B   C
 *
 * Without swapping node，swap L and N node value
 */
func (n *node[P, V]) rotateRight() {
	leftChild := n.leftChild()
	n.childs[left] = leftChild.leftChild()
	leftChild.childs = [2]*node[P, V]{leftChild.rightChild(), n.rightChild()}
	n.childs[right] = leftChild
	n.color, leftChild.color = leftChild.color, n.color
	n.entry, leftChild.entry = leftChild.entry, n.entry
	if child := n.leftChild(); child != nil { // A.parent = N
		child.parent = n
	}
	if child := leftChild.rightChild(); child != nil { // C.parent = L
		child.parent = leftChild
	}
	n.rightChild().updateMinMax()
	n.updateMinMax()
}

func (n *node[P, V]) rotate(direction direction) {
	if direction == left {
		n.rotateLeft()
	} else {
		n.rotateRight()
	}
}

func (n *node[P, V]) String() string {
	color := 'b'
	if n.color == red {
		color = 'r'
	}
	return fmt.Sprintf("%c:%s", color, &n.entry.interval)
}
