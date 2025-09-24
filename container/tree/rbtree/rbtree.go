package rbtree

import (
	"fmt"
	"github.com/hopeio/gox/cmp"
)

type color uint32

const (
	red color = iota
	black
)

type rbnode[K, V any] struct {
	c      color
	left   *rbnode[K, V]
	right  *rbnode[K, V]
	parent *rbnode[K, V]
	k      K
	v      V
}

func (n *rbnode[K, V]) color() color {
	if n == nil {
		return black
	}
	return n.c
}

func (n *rbnode[K, V]) grandparent() *rbnode[K, V] {
	return n.parent.parent
}

func (n *rbnode[K, V]) uncle() *rbnode[K, V] {
	if n.parent == n.grandparent().left {
		return n.grandparent().right
	}
	return n.grandparent().left
}

func (n *rbnode[K, V]) sibling() *rbnode[K, V] {
	if n == n.parent.left {
		return n.parent.right
	}
	return n.parent.left
}

func (n *rbnode[K, V]) maximumNode() *rbnode[K, V] {
	for n.right != nil {
		n = n.right
	}
	return n
}

// RBTree is a red-black tree
type RBTree[K, V any] struct {
	root *rbnode[K, V]
	len  int
	less cmp.LessFunc[K]
}

// NewRBTree creates a red-black tree
func NewRBTree[K, V any](less cmp.LessFunc[K]) *RBTree[K, V] {
	return &RBTree[K, V]{less: less}
}

// Len returns the size of the tree
func (t *RBTree[K, V]) Len() int {
	return t.len
}

// Put stores the value by given key
func (t *RBTree[K, V]) Put(key K, value V) {
	var insertedNode *rbnode[K, V]

	new := &rbnode[K, V]{k: key, v: value, c: red}
	if t.root != nil {
		node := t.root
	LOOP:
		for {
			switch {
			case t.less(key, node.k):
				if node.left == nil {
					node.left = new
					insertedNode = node.left
					break LOOP
				}
				node = node.left
			case t.less(node.k, key):
				if node.right == nil {
					node.right = new
					insertedNode = node.right
					break LOOP
				}
				node = node.right
			default: // =
				node.k = key
				node.v = value
				return
			}
		}
		insertedNode.parent = node
	} else {
		t.root = new
		insertedNode = t.root
	}
	t.insertCase1(insertedNode)
	t.len++
}

func (t *RBTree[K, V]) insertCase1(n *rbnode[K, V]) {
	if n.parent == nil {
		n.c = black
		return
	}
	t.insertCase2(n)
}
func (t *RBTree[K, V]) insertCase2(n *rbnode[K, V]) {
	if n.parent.color() == black {
		return
	}
	t.insertCase3(n)
}
func (t *RBTree[K, V]) insertCase3(n *rbnode[K, V]) {
	if n.uncle().color() == red {
		n.parent.c = black
		n.uncle().c = black
		n.grandparent().c = red
		t.insertCase1(n.grandparent())
		return
	}
	t.insertCase4(n)

}
func (t *RBTree[K, V]) insertCase4(n *rbnode[K, V]) {
	if n == n.parent.right && n.parent == n.grandparent().left {
		t.rotateLeft(n.parent)
		n = n.left
	} else if n == n.parent.left && n.parent == n.grandparent().right {
		t.rotateRight(n.parent)
		n = n.right
	}
	t.insertCase5(n)
}
func (t *RBTree[K, V]) insertCase5(n *rbnode[K, V]) {
	n.parent.c = black
	n.grandparent().c = red
	if n == n.parent.left && n.parent == n.grandparent().left {
		t.rotateRight(n.grandparent())
		return
	} else if n == n.parent.right && n.parent == n.grandparent().right {
		t.rotateLeft(n.grandparent())
	}
}

func (t *RBTree[K, V]) replace(old, new *rbnode[K, V]) {
	if old.parent == nil {
		t.root = new
	} else {
		if old == old.parent.left {
			old.parent.left = new
		} else {
			old.parent.right = new
		}
	}
	if new != nil {
		new.parent = old.parent
	}
}

func (t *RBTree[K, V]) rotateLeft(n *rbnode[K, V]) {
	right := n.right
	t.replace(n, right)
	n.right = right.left
	if right.left != nil {
		right.left.parent = n
	}
	right.left = n
	n.parent = right
}
func (t *RBTree[K, V]) rotateRight(n *rbnode[K, V]) {
	left := n.left
	t.replace(n, left)
	n.left = left.right
	if left.right != nil {
		left.right.parent = n
	}
	left.right = n
	n.parent = left
}

// Get returns the stored value by given key
func (t *RBTree[K, V]) Get(key K) (V, bool) {
	n := t.find(key)
	if n == nil {
		return *new(V), false
	}
	return n.v, true
}

func (t *RBTree[K, V]) find(key K) *rbnode[K, V] {
	n := t.root
	for n != nil {
		switch {
		case t.less(key, n.k):
			n = n.left
		case t.less(n.k, key):
			n = n.right
		default:
			return n
		}
	}
	return nil
}

// Del deletes the stored value by given key
func (t *RBTree[K, V]) Del(key K) {
	var child *rbnode[K, V]

	n := t.find(key)
	if n == nil {
		return
	}

	if n.left != nil && n.right != nil {
		pred := n.left.maximumNode()
		n.k = pred.k
		n.v = pred.v
		n = pred
	}

	if n.left == nil || n.right == nil {
		if n.right == nil {
			child = n.left
		} else {
			child = n.right
		}
		if n.c == black {
			n.c = child.color()
			t.delCase1(n)
		}

		t.replace(n, child)
		if n.parent == nil && child != nil {
			child.c = black
		}
	}
	t.len--
}

func (t *RBTree[K, V]) delCase1(n *rbnode[K, V]) {
	if n.parent == nil {
		return
	}

	t.delCase2(n)
}
func (t *RBTree[K, V]) delCase2(n *rbnode[K, V]) {
	sibling := n.sibling()
	if sibling.color() == red {
		n.parent.c = red
		sibling.c = black
		if n == n.parent.left {
			t.rotateLeft(n.parent)
		} else {
			t.rotateRight(n.parent)
		}
	}
	t.delCase3(n)
}
func (t *RBTree[K, V]) delCase3(n *rbnode[K, V]) {
	sibling := n.sibling()
	if n.parent.color() == black &&
		sibling.color() == black &&
		sibling.left.color() == black &&
		sibling.right.color() == black {
		sibling.c = red
		t.delCase1(n.parent)
		return
	}
	t.delCase4(n)
}
func (t *RBTree[K, V]) delCase4(n *rbnode[K, V]) {
	sibling := n.sibling()
	if n.parent.color() == red &&
		sibling.color() == black &&
		sibling.left.color() == black &&
		sibling.right.color() == black {
		sibling.c = red
		n.parent.c = black
		return
	}
	t.delCase5(n)
}
func (t *RBTree[K, V]) delCase5(n *rbnode[K, V]) {
	sibling := n.sibling()
	if n == n.parent.left &&
		sibling.color() == black &&
		sibling.left.color() == red &&
		sibling.right.color() == black {
		sibling.c = red
		sibling.left.c = black
		t.rotateRight(sibling)
	} else if n == n.parent.right &&
		sibling.color() == black &&
		sibling.right.color() == red &&
		sibling.left.color() == black {
		sibling.c = red
		sibling.right.c = black
		t.rotateLeft(sibling)
	}
	t.delCase6(n)
}
func (t *RBTree[K, V]) delCase6(n *rbnode[K, V]) {
	sibling := n.sibling()
	sibling.c = n.parent.color()
	n.parent.c = black
	if n == n.parent.left && sibling.right.color() == red {
		sibling.right.c = black
		t.rotateLeft(n.parent)
		return
	}
	sibling.left.c = black
	t.rotateRight(n.parent)
}

func (t *RBTree[K, V]) String() string {
	str := "RBTree\n"
	if t.Len() != 0 {
		t.root.output("", true, &str)
	}
	return str
}

func (n *rbnode[K, V]) String() string {
	return fmt.Sprintf("%v", n.k)
}

func (n *rbnode[K, V]) output(prefix string, isTail bool, str *string) {
	if n.right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		n.right.output(newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += n.String() + "\n"
	if n.left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		n.left.output(newPrefix, true, str)
	}
}
