package rbmap

import (
	"cmp"
)

type iterator[K cmp.Ordered, V any] struct {
	m    *rbMap[K, V]
	n    *node[K, V]     //迭代器指向的node
	prev *iterator[K, V] //上一个迭代器
	next *iterator[K, V] //下一个迭代器

	key K
	ctx V
}

func (itr *iterator[K, V]) Next() (Iterator[K, V], error) {
	if itr.m == nil {
		return nil, ErrorInvalid
	}
	if itr.next == nil {
		return itr, nil //说明这应该是end_iterator
	}
	return itr.next, nil
}

func (itr *iterator[K, V]) Prev() (Iterator[K, V], error) {
	if itr.m == nil {
		return nil, ErrorInvalid
	}
	if itr.prev.n == nil { //有一个help_iterator
		return itr, nil
	}
	return itr.prev, nil
}

func (itr *iterator[K, V]) Key() (K, error) {
	if itr.m == nil {
		return itr.key, ErrorInvalid
	}
	if itr == itr.m.endIter {
		return itr.key, ErrorEndIter
	}
	return itr.key, nil
}

func (itr *iterator[K, V]) Value() (V, error) {
	if itr.m == nil {
		return itr.ctx, ErrorInvalid
	}
	if itr == itr.m.endIter {
		return itr.ctx, ErrorEndIter
	}
	return itr.ctx, nil
}

func (itr *iterator[K, V]) SetValue(val V) error {
	if itr.m == nil {
		return ErrorInvalid
	}
	if itr == itr.m.endIter {
		return ErrorEndIter
	}
	itr.ctx = val
	return nil
}

func (itr *iterator[K, V]) Delete() (Iterator[K, V], error) {
	next, err := itr.Next()
	if err != nil {
		return nil, err
	}
	if itr.m.deleteNode(itr.n) {
		return itr, nil
	} else {
		return next, nil
	}
}

func (itr *iterator[K, V]) NextNoError() (out Iterator[K, V]) {
	if itr.m == nil {
		return
	}
	if itr.next == nil {
		return
	}
	return itr.next
}

func (itr *iterator[K, V]) PrevNoError() (out Iterator[K, V]) {
	if itr.m == nil {
		return nil
	}
	if itr.prev.n == nil { //有一个help_iterator
		return itr
	}
	return itr.prev
}

func (itr *iterator[K, V]) KeyNoError() (out K) {
	if itr.m == nil {
		return
	}
	if itr == itr.m.endIter {
		return
	}
	return itr.key
}

func (itr *iterator[K, V]) ValueNoError() (out V) {
	if itr.m == nil {
		return
	}
	if itr == itr.m.endIter {
		return
	}
	return itr.ctx
}
