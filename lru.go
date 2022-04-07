package cache

import (
	"sync"

	"github.com/FallenTaters/cache/list"
)

type lru[K comparable, V any] struct {
	sync.RWMutex

	adder addManager[K, V]

	maxEntries int
	entries    *list.List[keyValue[K, V]]
	values     map[K]listValue[K, V]
}

func LRU[K comparable, V any](maxEntries int) Cache[K, V] {
	return &lru[K, V]{
		maxEntries: maxEntries,
		adder: addManager[K, V]{
			busyKeys: make(map[K][]chan result[V]),
		},
		entries: list.New[keyValue[K, V]](),
		values:  make(map[K]listValue[K, V]),
	}
}

func (l *lru[K, V]) Get(key K) (V, bool) {
	l.Lock()
	defer l.Unlock()

	v, ok := l.values[key]
	if !ok {
		return v.value, false
	}

	l.entries.MoveToFront(v.element)

	return v.value, true
}

func (l *lru[K, V]) Add(key K, value V) {
	l.Lock()
	defer l.Unlock()

	l.add(key, value)
}

func (l *lru[K, V]) GetOrAdd(key K, addFunc AddFunc[V]) (V, error) {
	v, ok := l.Get(key)

	if ok {
		return v, nil
	}

	result := <-l.adder.waitOrAdd(key, addFunc)
	if result.Err == nil {
		l.Lock()
		l.add(key, result.Value)
		l.Unlock()
	}

	return result.Value, result.Err
}

func (l *lru[K, V]) MustGetOrAdd(key K, addFunc AddFunc[V]) V {
	v, err := l.GetOrAdd(key, addFunc)
	if err != nil {
		panic(err)
	}

	return v
}

func (l *lru[K, V]) Delete(key K) {
	l.Lock()
	defer l.Unlock()

	v, ok := l.values[key]
	if !ok {
		return
	}

	l.entries.Remove(v.element)
	delete(l.values, key)
}

func (l *lru[K, V]) add(key K, value V) {
	if v, ok := l.values[key]; ok {
		l.entries.MoveToFront(v.element)
		v.value = value
		l.values[key] = v
		return
	}

	if l.entries.Len() == l.maxEntries {
		back := l.entries.Back()
		delete(l.values, back.Value.key)
		l.entries.Remove(back)
	}

	elem := l.entries.PushFront(keyValue[K, V]{
		key:   key,
		value: value,
	})

	l.values[key] = listValue[K, V]{
		element: elem,
		value:   value,
	}
}
