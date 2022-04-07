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

	adder addManager[K, V]

	maxEntries int
	entries    *list.List[keyValue[K, V]]
	values     map[K]listValue[K, V]
}

func FIFO[K comparable, V any](maxEntries int) Cache[K, V] {
	return &fifo[K, V]{
		maxEntries: maxEntries,
		adder: addManager[K, V]{
			busyKeys: make(map[K][]chan result[V]),
		},
		entries: list.New[keyValue[K, V]](),
		values:  make(map[K]listValue[K, V]),
	}
}

func (f *fifo[K, V]) Get(key K) (V, bool) {
	f.RLock()
	defer f.RUnlock()

	v, ok := f.values[key]
	return v.value, ok
}

func (f *fifo[K, V]) Add(key K, value V) {
	f.Lock()
	defer f.Unlock()

	f.add(key, value)
}

func (f *fifo[K, V]) GetOrAdd(key K, addFunc AddFunc[V]) (V, error) {
	f.RLock()
	v, ok := f.values[key]
	f.RUnlock()

	if ok {
		return v.value, nil
	}

	result := <-f.adder.waitOrAdd(key, addFunc)
	if result.Err == nil {
		f.Lock()
		f.add(key, result.Value)
		f.Unlock()
	}

	return result.Value, result.Err
}

func (f *fifo[K, V]) MustGetOrAdd(key K, addFunc AddFunc[V]) V {
	v, err := f.GetOrAdd(key, addFunc)
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
	if v, ok := f.values[key]; ok {
		v.value = value
		f.values[key] = v
		return
	}

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
