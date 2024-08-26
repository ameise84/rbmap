package rbmap

import (
	"cmp"
)

type SearchMode = uint8

const (
	SearchModeLT SearchMode = 1 << iota
	SearchModeET
	SearchModeGT
)

type RBMap[K cmp.Ordered, V any] interface {
	Store(key K, val V)
	Load(key K) (val V, ok bool)
	LoadAndStore(key K, val V) (old V, ok bool)
	Delete(key K) (V, bool)
	Clean() map[K]V
	Reset()
	BeginIterator() Iterator[K, V]
	EndIterator() Iterator[K, V]
	Range(func(Iterator[K, V]) bool)
	Search(key K, mode ...SearchMode) Iterator[K, V]
	Size() int
	//PrintTree()
	//PrintTreeIter()
}

type Iterator[K cmp.Ordered, V any] interface {
	Next() (Iterator[K, V], error)
	Prev() (Iterator[K, V], error)
	Key() (K, error)
	Value() (V, error)
	SetValue(V) error
	Delete() (Iterator[K, V], error) //返回下一个迭代器

	NextNoError() Iterator[K, V]
	PrevNoError() Iterator[K, V]
	KeyNoError() K
	ValueNoError() V
}
