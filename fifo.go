package cache

import (
	"sync"

	"github.com/FallenTaters/cache/list"
)

type keyValue[K comparable, V any] struct {
	key   K
	value V
}

type listValue[K comparable, V any] struct {
	element *list.Element[keyValue[K, V]]
	value   V
}

type fifo[K comparable, V any] struct {
	sync.RWMutex

	maxEntries int
	addFunc    AddFunc[K, V]
	entries    *list.List[keyValue[K, V]]
	values     map[K]listValue[K, V]
}

func FIFO[K comparable, V any](maxEntries int, addFunc AddFunc[K, V]) Cache[K, V] {
	if addFunc == nil {
		panic(`nil addFunc passed`)
	}

	return &fifo[K, V]{
		maxEntries: maxEntries,
		addFunc:    addFunc,
		entries:    list.New[keyValue[K, V]](),
		values:     make(map[K]listValue[K, V]),
	}
}

func (f *fifo[K, V]) Get(key K) (V, bool) {
	f.RLock()
	defer f.RUnlock()

	v, ok := f.values[key]
	return v.value, ok
}

func (f *fifo[K, V]) GetOrEmpty(key K) V {
	f.RLock()
	defer f.RUnlock()

	return f.values[key].value
}

func (f *fifo[K, V]) GetOrAdd(key K) (V, error) {
	f.RLock()
	v, ok := f.values[key]
	f.RUnlock()

	if ok {
		return v.value, nil
	}

	val, err := f.addFunc(key)
	if err == nil {
		f.add(key, val)
	}

	return val, err
}

func (f *fifo[K, V]) MustGetOrAdd(key K) V {
	v, err := f.GetOrAdd(key)
	if err != nil {
		panic(err)
	}

	return v
}

func (f *fifo[K, V]) Delete(key K) {
	f.Lock()
	defer f.Unlock()

	v, ok := f.values[key]
	if !ok {
		return
	}

	f.entries.Remove(v.element)
	delete(f.values, key)
}

func (f *fifo[K, V]) add(key K, value V) {
	f.Lock()
	defer f.Unlock()

	if f.entries.Len() == f.maxEntries {
		back := f.entries.Back()
		delete(f.values, back.Value.key)
		f.entries.Remove(back)
	}

	elem := f.entries.PushFront(keyValue[K, V]{
		key:   key,
		value: value,
	})

	f.values[key] = listValue[K, V]{
		element: elem,
		value:   value,
	}
}
