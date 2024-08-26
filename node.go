package rbmap

import (
	"cmp"
)

const (
	red = iota
	black
)

func newNode[K cmp.Ordered, V any](m *rbMap[K, V], key K, val V) *node[K, V] {
	n := m.pool.Get().(*node[K, V])
	n.father = nil
	n.left = m.leaf
	n.right = m.leaf
	n.color = red

	n.iter.m = m
	n.iter.n = n
	n.iter.prev = nil
	n.iter.next = nil
	n.iter.key = key
	n.iter.ctx = val
	return n
}

func freeNode[K cmp.Ordered, V any](n *node[K, V]) {
	m := n.iter.m
	n.father = nil
	n.left = m.leaf
	n.right = m.leaf
	n.color = red

	n.iter.m = nil
	n.iter.n = nil
	n.iter.prev = nil
	n.iter.next = nil
	n.iter.key = m.zeroK
	n.iter.ctx = m.zeroV
	m.pool.Put(n)
}

type node[K cmp.Ordered, V any] struct {
	father *node[K, V] //父节点
	left   *node[K, V] //左子树节点
	right  *node[K, V] //右子树节点
	color  int         //当前节点着色
	iter   *iterator[K, V]
}

func (n *node[K, V]) isRoot() bool {
	return n.father == nil
}

func (n *node[K, V]) grandfather() *node[K, V] {
	if n.isRoot() {
		return nil
	}
	return n.father.father
}

func (n *node[K, V]) uncle() *node[K, V] {
	if n.grandfather() == nil {
		return nil
	}
	return n.father.brother()
}

func (n *node[K, V]) brother() *node[K, V] {
	if n.father.left == n {
		return n.father.right
	} else {
		return n.father.left
	}
}
