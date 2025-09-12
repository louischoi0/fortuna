package structure

import (
    "sort"
)

type SortedListKeyContraints interface {
	int64 | string | uint64
}

type SortedList[K SortedListKeyContraints, V any] struct {
    m    map[K]V
    keys []K
}

func NewSortedList[K SortedListKeyContraints, V any]() *SortedList[K, V] {
    return &SortedList[K, V]{
        m:    make(map[K]V),
        keys: make([]K, 0),
    }
}

func (sl *SortedList[K, V]) Set(key K, value V) {
    if _, exists := sl.m[key]; !exists {
        i := sort.Search(len(sl.keys), func(i int) bool {
            return sl.keys[i] >= key
        })
        sl.keys = append(sl.keys, key)      
        copy(sl.keys[i+1:], sl.keys[i:])   
        sl.keys[i] = key
    }
    sl.m[key] = value
}

func (sl *SortedList[K, V]) Get(key K) (V, bool) {
    v, ok := sl.m[key]
    return v, ok
}

func (sl *SortedList[K, V]) LowerBound(target K) (key K, value V, found bool) {
    i := sort.Search(len(sl.keys), func(i int) bool {
        return sl.keys[i] >= target
    })
    if i < len(sl.keys) {
        key = sl.keys[i]
        value = sl.m[key]
        found = true
        return
    }

    return 
}

func (sl *SortedList[K, V]) Delete(key K) {
    if _, exists := sl.m[key]; !exists {
        return
    }
    delete(sl.m, key)

    i := sort.Search(len(sl.keys), func(i int) bool {
        return sl.keys[i] >= key
    })
    if i < len(sl.keys) && sl.keys[i] == key {
        sl.keys = append(sl.keys[:i], sl.keys[i+1:]...)
    }
}

func (sl *SortedList[K, V]) Len() int {
    return len(sl.keys)
}

func (sl *SortedList[K, V]) Keys() []K {
    out := make([]K, len(sl.keys))
    copy(out, sl.keys)
    return out
}

func (sl *SortedList[K, V]) Values() []V {
    out := make([]V, len(sl.keys))
    for i, k := range sl.keys {
        out[i] = sl.m[k]
    }
    return out
}

func (sl *SortedList[K, V]) Iterate(f func(key K, value V)) {
    for _, k := range sl.keys {
        f(k, sl.m[k])
    }
}

func (sl *SortedList[K, V]) Iterator() *Iterator[K, V] {
    return &Iterator[K, V]{list: sl, index: -1}
}

type Iterator[K SortedListKeyContraints, V any] struct {
    list  *SortedList[K, V]
    index int
}

func (it *Iterator[K, V]) Next() bool {
    it.index++
    return it.index < len(it.list.keys)
}

func (it *Iterator[K, V]) Key() K {
    return it.list.keys[it.index]
}

func (it *Iterator[K, V]) Value() V {
    return it.list.m[it.Key()]
}

