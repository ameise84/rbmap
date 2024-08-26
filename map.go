package rbmap

import (
	"cmp"
	"fmt"
	"sync"
)

func New[K cmp.Ordered, V any]() RBMap[K, V] {
	m := &rbMap[K, V]{leaf: &node[K, V]{color: black}}
	m.root = m.leaf
	m.helpIter = &iterator[K, V]{m: m}
	m.endIter = &iterator[K, V]{m: m}
	m.helpIter.next = m.endIter
	m.endIter.prev = m.helpIter
	m.pool = sync.Pool{
		New: func() any {
			n := &node[K, V]{iter: &iterator[K, V]{}}
			n.iter.n = n
			return n
		}}
	return m
}

type rbMap[K cmp.Ordered, V any] struct {
	leaf     *node[K, V]
	zeroK    K
	zeroV    V
	root     *node[K, V]
	count    int
	helpIter *iterator[K, V] //第一个迭代器,用于方便寻找 begin 节点
	endIter  *iterator[K, V]
	pool     sync.Pool
}

func (m *rbMap[K, V]) Store(key K, val V) {
	_, _ = m.loadAndStore(key, val)
}

func (m *rbMap[K, V]) Load(key K) (val V, ok bool) {
	_, n := m.search(key, nil, m.root)
	if n != m.leaf {
		return n.iter.ctx, true
	}
	return m.zeroV, false
}

func (m *rbMap[K, V]) LoadAndStore(key K, val V) (old V, ok bool) {
	return m.loadAndStore(key, val)
}

func (m *rbMap[K, V]) Delete(key K) (V, bool) {
	return m.delete(key)
}

func (m *rbMap[K, V]) Clean() map[K]V {
	if m.count == 0 {
		return nil
	}
	out := make(map[K]V, m.count)
	itr := m.BeginIterator()
	for itr != m.endIter {
		out[itr.KeyNoError()] = itr.ValueNoError()
		itr, _ = itr.Next()
	}
	m.count = 0
	m.root = m.leaf
	m.helpIter.next = m.endIter
	m.endIter.prev = m.helpIter
	return out
}

func (m *rbMap[K, V]) Reset() {
	m.count = 0
	m.root = m.leaf
	m.helpIter.next = m.endIter
	m.endIter.prev = m.helpIter
}

func (m *rbMap[K, V]) BeginIterator() Iterator[K, V] {
	return m.helpIter.next
}

func (m *rbMap[K, V]) EndIterator() Iterator[K, V] {
	return m.endIter
}

func (m *rbMap[K, V]) Range(f func(Iterator[K, V]) bool) {
	iter := m.BeginIterator()
	for iter != m.endIter {
		goNext := f(iter)
		if !goNext {
			break
		}
		iter, _ = iter.Next()
	}
}

func (m *rbMap[K, V]) Search(key K, mode ...SearchMode) Iterator[K, V] {
	if m.root == m.leaf { //空树
		return m.endIter
	}

	md := SearchModeET
	if mode != nil {
		if mode[0]&(SearchModeLT|SearchModeGT) == SearchModeLT|SearchModeGT {
			panic("rbmap search mod ill")
		}
		md = mode[0]
	}

	f, n := m.search(key, nil, m.root)

	if f == nil {
		f = n //n是根节点
	}

	if n == m.leaf {
		//只找=
		if md == SearchModeET {
			return m.endIter
		}
		//要找的是<或<=
		if md&SearchModeLT != 0 {
			if f.iter.key < key { //查询到的父节点小于key
				return f.iter
			} else {
				if f.iter.prev.n == nil {
					return m.endIter
				}
				return f.iter.prev
			}
		}
		//要找的是>或>=
		if f.iter.key > key {
			return f.iter
		} else { //最小的节点都比要找的大
			return f.iter.next
		}
	} else {
		//要找包含=
		if md&SearchModeET != 0 {
			return n.iter
		}
		//要找的是<
		if md == SearchModeLT {
			if n.iter.key < key {
				return f.iter
			} else {
				if n.iter.prev.n == nil {
					return m.endIter
				}
				return n.iter.prev
			}
		}
		//要找的是>
		if n.iter.key > key {
			return n.iter
		} else { //最小的节点都比要找的大
			return n.iter.next
		}
	}
}

func (m *rbMap[K, V]) Size() int {
	return m.count
}

func (m *rbMap[K, V]) PrintTree() {
	fmt.Println("tree node count:", m.count)
	if m.root == m.leaf {
		fmt.Println("< nil >")
	} else {
		m.printNode(m.root, 0)
	}
	m.checkBalance()
}

func (m *rbMap[K, V]) PrintTreeIter() {
	fmt.Println("tree node count:", m.count)

	iter := m.BeginIterator()
	if iter == m.endIter {
		fmt.Println("< nil >")
	} else {
		m.printNodeIter(nil, iter)
	}
}

// -------------------------------------------------//
func (m *rbMap[K, V]) loadAndStore(key K, val V) (old V, ok bool) {
	f, c := m.search(key, nil, m.root)
	if c == m.leaf {
		n := newNode(m, key, val)
		n.father = f
		if f != nil {
			if key < f.iter.key {
				f.left = n
				f.iter.prev.next = n.iter
				n.iter.prev = f.iter.prev
				n.iter.next = f.iter
				f.iter.prev = n.iter
			} else {
				f.right = n
				f.iter.next.prev = n.iter
				n.iter.next = f.iter.next
				n.iter.prev = f.iter
				f.iter.next = n.iter
			}
		} else {
			m.root = n
			m.helpIter.next = n.iter
			n.iter.prev = m.helpIter
			n.iter.next = m.endIter
			m.endIter.prev = n.iter
		}
		m.insertCase1(n)
		m.count++
		return m.zeroV, false
	} else {
		old, c.iter.ctx = c.iter.ctx, val
		return old, true
	}
}

func (m *rbMap[K, V]) search(key K, f, n *node[K, V]) (*node[K, V], *node[K, V]) {
	if n == m.leaf {
		return f, n
	}

	if key == n.iter.key {
		return f, n
	} else if key < n.iter.key {
		return m.search(key, n, n.left)
	} else {
		return m.search(key, n, n.right)
	}
}

func (m *rbMap[K, V]) delete(key K) (V, bool) {
	_, n := m.search(key, nil, m.root)
	if n == m.leaf {
		return m.zeroV, false
	}
	ctx := n.iter.ctx
	m.deleteNode(n)
	m.count--
	return ctx, true
}

func (m *rbMap[K, V]) deleteNode(n *node[K, V]) (isSwap bool) {
	//交换机制
	var tn *node[K, V]
	if n.left != m.leaf && n.right != m.leaf {
		tn = n.iter.next.n

		n.iter, tn.iter = tn.iter, n.iter
		n.iter.n, tn.iter.n = n, tn

		isSwap = true
	} else {
		tn = n
	}
	//根据原则5:n节点如果有孩子,n必然是黑节点,孩子必然是红色
	c := tn.right
	if tn.left != m.leaf {
		c = tn.left
	}
	//n为根节点时
	if tn.isRoot() {
		m.root = c
		c.father = nil
		c.color = black
	} else {
		if m.checkColor(tn, black) {
			//这里有孩子必然是红色,否则没有孩子
			if m.checkColor(c, red) {
				//孩子继承父亲的位置
				c.color = black
			} else {
				tn.left = nil
				tn.right = nil
				m.deleteCase1(tn)
			}
		}

		if tn.father.left == tn {
			tn.father.left = c
		} else {
			tn.father.right = c
		}

		if c != m.leaf {
			c.father = tn.father // c有可能是叶子,临时将c当做一个节点处理,方便后续做deleteCase
		}
	}
	tn.iter.prev.next = tn.iter.next
	tn.iter.next.prev = tn.iter.prev
	freeNode(tn)
	return
}

func (m *rbMap[K, V]) leftRotate(n *node[K, V]) {
	g := n.grandfather()
	f := n.father
	c := n.left

	n.father = g
	if g == nil {
		m.root = n
	} else {
		if g.left == f {
			g.left = n
		} else {
			g.right = n
		}
	}

	f.right = c
	if c != m.leaf {
		c.father = f
	}

	n.left = f
	f.father = n
}

func (m *rbMap[K, V]) rightRotate(n *node[K, V]) {
	g := n.grandfather()
	f := n.father
	c := n.right

	n.father = g
	if g == nil {
		m.root = n
	} else {
		if g.left == f {
			g.left = n
		} else {
			g.right = n
		}
	}

	f.left = c
	if c != m.leaf {
		c.father = f
	}

	n.right = f
	f.father = n
}

func (m *rbMap[K, V]) insertCase1(n *node[K, V]) {
	if n.isRoot() { //新节点是根节点
		n.color = black
	} else {
		m.insertCase2(n)
	}
}

func (m *rbMap[K, V]) insertCase2(n *node[K, V]) {
	if !m.checkColor(n.father, black) { //新节点的父亲是黑色
		m.insertCase3(n)
	}
}

func (m *rbMap[K, V]) insertCase3(n *node[K, V]) {
	//父亲是红色,那么祖父一定不为空且为黑色
	if m.checkColor(n.uncle(), red) { //新节点的父亲和叔叔是红色
		n.father.color = black
		n.grandfather().color = red
		n.uncle().color = black
		m.insertCase1(n.grandfather())
	} else {
		m.insertCase4(n)
	}
}

func (m *rbMap[K, V]) insertCase4(n *node[K, V]) {
	if n == n.father.left { //n是左,父亲是右
		if n.father == n.grandfather().right {
			m.rightRotate(n)
			n.color = black
			n.father.color = red
			m.leftRotate(n)
		} else {
			n.father.color = black
			n.grandfather().color = red
			m.rightRotate(n.father)
		}

	} else if n == n.father.right {
		if n.father == n.grandfather().left {
			m.leftRotate(n)
			n.color = black
			n.father.color = red
			m.rightRotate(n)
		} else {
			n.father.color = black
			n.grandfather().color = red
			m.leftRotate(n.father)
		}
	}
}

func (m *rbMap[K, V]) deleteCase1(n *node[K, V]) {
	//根节点直接退出
	if n.isRoot() {
		return
	}
	//必然有兄弟
	b := n.brother()
	//兄弟是红色,兄弟必然有2个黑儿子
	if m.checkColor(n.brother(), red) {
		n.father.color = red
		b.color = black
		if n == n.father.left {
			m.leftRotate(b)
		} else {
			m.rightRotate(b)
		}
	}
	m.deleteCase2(n)
}

func (m *rbMap[K, V]) deleteCase2(n *node[K, V]) {
	b := n.brother()
	if m.checkColor(b.left, black) && m.checkColor(b.right, black) { //兄弟没有孩子 或有 2个黑孩子  (根据规则3,叶子算做黑色,所以可以从简判断)
		b.color = red
		if m.checkColor(n.father, red) { //父亲是红色
			n.father.color = black
		} else {
			m.deleteCase1(n.father)
		}
	} else {
		m.deleteCase3(n)
	}
}

func (m *rbMap[K, V]) deleteCase3(n *node[K, V]) {
	b := n.brother()

	for {
		if (n == n.father.left && m.checkColor(b.right, red)) || (n == n.father.right && m.checkColor(b.left, red)) {
			b.color = n.father.color
			n.father.color = black
			if n == n.father.left {
				b.right.color = black
				m.leftRotate(b)
			} else {
				b.left.color = black
				m.rightRotate(b)
			}
			return
		} else {
			b.color = red
			if n == n.father.left {
				b.left.color = black
				m.rightRotate(b.left)
			} else {
				b.right.color = black
				m.leftRotate(b.right)
			}
			b = n.brother()
		}
	}
}

func (m *rbMap[K, V]) printNodeIter(b Iterator[K, V], i Iterator[K, V]) {
	if i == m.endIter {
		return
	}
	if b != nil {
		if b.KeyNoError() >= i.KeyNoError() {
			panic("key err")
		}
	}
	fmt.Println("<", i.KeyNoError(), "><", i.ValueNoError(), ">")
	m.printNodeIter(i, i.NextNoError())
}

func (m *rbMap[K, V]) printNode(n *node[K, V], level int) {
	var color string
	if n.color == red {
		color = "red"
	} else {
		color = "black"
	}
	if n != m.leaf {
		for i := 0; i < level; i++ {
			fmt.Print("|     ")
		}
		var lr string
		if n.isRoot() {
			lr = "X"
		} else if n.father.left == n {
			lr = "L"
		} else if n.father.right == n {
			lr = "R"
		}
		fmt.Println("|", lr, "--<", n.iter.key, "><", color, level, n.iter.ctx, ">")
		m.printNode(n.left, level+1)
		m.printNode(n.right, level+1)
	}
}

func (m *rbMap[K, V]) checkBalance() {
	maxCnt := 0
	cnt := 0
	m.checkBalance2(m.root, &maxCnt, cnt)
}

func (m *rbMap[K, V]) checkBalance2(n *node[K, V], maxCnt *int, cnt int) {
	if n.iter != nil {
		fmt.Printf("check node:%d   maxCnt:%d  cnt:%d\n", n.iter.key, *maxCnt, cnt)
	}
	if m.checkColor(n, red) && (m.checkColor(n.left, red) || m.checkColor(n.right, red)) {
		panic("checkBalance failed")
	}
	if m.checkColor(n, black) {
		cnt += 1
	}

	if n == m.leaf {
		if *maxCnt == 0 {
			*maxCnt = cnt
		}
		if *maxCnt != cnt {
			panic("checkBalance failed")
		}
		return
	}

	if n.left != m.leaf {
		m.checkBalance2(n.left, maxCnt, cnt)
	} else {
		if *maxCnt == 0 {
			*maxCnt = cnt
		}
		if *maxCnt != cnt {
			panic("checkBalance failed")
		}
	}

	if n.right != m.leaf {
		m.checkBalance2(n.right, maxCnt, cnt)
	} else {
		if *maxCnt == 0 {
			*maxCnt = cnt
		}
		if *maxCnt != cnt {
			panic("checkBalance failed")
		}
	}
}

func (m *rbMap[K, V]) checkColor(n *node[K, V], color int) bool {
	if n == m.leaf {
		return black == color
	}
	return n.color == color
}
