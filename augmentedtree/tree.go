package augmentedtree

import (
	"fmt"
	"math"
	"strings"
)

// Augmented tree is a special kind of binary sort tree,
// each node will keep sub-tree max and min range.
// Here use red-black tree to maintain self rebalance
type Tree[P Point[P], V any] struct {
	root *node[P, V]
	size uint64
}

func (t *Tree[P, V]) Query(interval Interval[P]) *Entry[P, V] {
	if t.root == nil {
		return nil
	}
	if !t.root.minMax.overlaps(&interval) {
		return nil
	}
	entry := t.root.query(interval, func(_ *Entry[P, V]) bool { return true })
	if entry == nil {
		return nil
	}
	return entry
}

//go:inline
func (t *Tree[P, V]) Size() uint64 { return t.size }

//go:inline
func updateMinMaxFrom[P Point[P], V any](node *node[P, V]) {
	for ; node != nil; node = node.parent {
		node.updateMinMax()
	}
}

func rebalanceAfterInsert[P Point[P], V any](cur *node[P, V]) *node[P, V] {
	for {
		if cur.parent == nil { // case 1: root node
			cur.color = black
			break
		}
		if cur.parent.color == black { // case 2: parent node is black
			break
		}
		parent := cur.parent
		grandParent := cur.grandParent()
		if uncle := cur.uncle(); uncle.is(red) { // case 3: parent and uncle is red
			grandParent.color, parent.color, uncle.color = red, black, black
			parent.updateMinMax()
			grandParent.updateMinMax()
			cur = grandParent
			continue
		}
		direction := cur.direction()
		// case 4：grand-parent/parent and parent/node not on same side
		if parent.direction() != direction {
			parent.rotate(direction.opposite())
			// After rotation cur is the new sub-node, no need to reassignment
			parent = cur.parent
			direction = cur.direction()
		}
		// case 5: grand-parent/parent and parent/node on the same side
		if parent.direction() == direction {
			parent.color, grandParent.color = black, red
			grandParent.rotate(direction.opposite())
		}
		break
	}
	return cur
}

func (t *Tree[P, V]) Put(interval Interval[P], value V) bool {
	if t.size+1 == math.MaxUint64 {
		panic("Maximum size")
	}
	newNode := &node[P, V]{minMax: interval, entry: Entry[P, V]{interval, value}, color: red}
	if t.root == nil {
		t.root = newNode
		t.root.color = black
		t.size++
		return true
	}

	cur := t.root
	for newNode != nil {
		switch compareIntervals(interval, cur.entry.interval) {
		case 0:
			return false
		case -1:
			if cur.leftChild() == nil {
				newNode.parent = cur
				cur.childs[left], newNode = newNode, nil
			}
			cur = cur.leftChild()
		case 1:
			if cur.rightChild() == nil {
				newNode.parent = cur
				cur.childs[right], newNode = newNode, nil
			}
			cur = cur.rightChild()
		}
	}
	t.size++
	updateMinMaxFrom(rebalanceAfterInsert(cur))
	return true
}

func rebalanceAfterDelete[P Point[P], V any](cur *node[P, V]) *node[P, V] {
	for {
		if cur.parent == nil { // case 1: root node
			break
		}
		dir := cur.direction()
		parent := cur.parent
		sibling := cur.slibing()  // sibling won't be nil, because deleted node is black
		if sibling.color == red { // case 2: sibling is red
			parent.color, sibling.color = red, black
			parent.rotate(dir)
			parent, sibling = cur.parent, cur.slibing()
			dir = cur.direction()
		}
		// because case 2, sibling is black
		// case 3 and 4, sibling and two son of sibling is all black
		if sibling.leftChild().is(black) && sibling.rightChild().is(black) {
			if parent.color == black { // case 3: parent is black too
				sibling.color = red
				parent.updateMinMax()
				cur = parent
				continue
			} else if parent.color == red { // case 4: parent is red
				sibling.color, parent.color = parent.color, sibling.color
				break
			}
		}
		// case 5: sibling is black, sibling left son is red, right son is black
		if sibling.child(dir).is(red) && sibling.child(dir.opposite()).is(black) {
			sibling.color, sibling.child(dir).color = red, black
			sibling.rotate(dir.opposite())
			sibling = cur.slibing()
			dir = cur.direction()
		}
		// case 6: sibling is black, sibling left son is red,
		// current node is left son of parent
		sibling.color, parent.color = parent.color, black
		sibling.child(dir.opposite()).color = black
		cur.parent.rotate(dir)
		break
	}
	return cur
}

func (t *Tree[P, V]) locate(interval Interval[P]) *node[P, V] {
	for cur := t.root; cur != nil; {
		if !cur.minMax.contains(&interval) {
			return nil
		}
		switch compareIntervals(interval, cur.entry.interval) {
		case 0:
			if cur.entry.interval != interval {
				return nil
			}
			return cur
		case -1:
			cur = cur.leftChild()
		case 1:
			cur = cur.rightChild()
		}
	}
	return nil
}

func (t *Tree[P, V]) Delete(interval Interval[P]) bool {
	if t.root == nil {
		return false
	}

	deleting := t.locate(interval)
	if deleting == nil {
		return false
	}
	t.size--
	if t.root == deleting && t.size == 0 {
		t.root = nil
		return true
	}
	if deleting.leftChild() != nil && deleting.rightChild() != nil {
		// copy predecessor node value and delete predecessor node
		parent := deleting.leftChild()
		pre := parent
		for pre.rightChild() != nil {
			pre = pre.rightChild()
		}
		deleting.entry = pre.entry
		deleting = pre
	}
	replace := deleting.leftChild()
	if replace == nil {
		replace = deleting.rightChild()
	}
	if replace != nil {
		*deleting, *replace = *replace, *deleting
		deleting, replace = replace, deleting
		replace.parent = deleting.parent
	} else {
		replace = deleting
		replace.minMax = replace.parent.entry.interval
	}
	if deleting.color == red || replace.color == red {
		updateMinMaxFrom(replace)
		replace.color = black
	} else {
		updateMinMaxFrom(rebalanceAfterDelete(replace))
	}
	if replace == deleting {
		if deleting.parent.leftChild() == deleting {
			deleting.parent.childs[left] = nil
		} else {
			deleting.parent.childs[right] = nil
		}
	}
	return true
}

func (t *Tree[P, V]) String() string {
	if t.root == nil {
		return ""
	}
	nodes := []*node[P, V]{t.root}
	var childs []*node[P, V]
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%v\n", t.root))
	for len(nodes) > 0 {
		for _, node := range nodes {
			if left := node.leftChild(); left != nil {
				buf.WriteString(left.String())
				buf.WriteRune(' ')
				childs = append(childs, left)
			} else {
				buf.WriteString("nil ")
			}
			if right := node.rightChild(); right != nil {
				buf.WriteString(right.String())
				buf.WriteRune(' ')
				childs = append(childs, right)
			} else {
				buf.WriteString("nil ")
			}
		}
		buf.WriteRune('\n')
		nodes, childs = childs, nodes[:0]
	}
	return buf.String()
}
