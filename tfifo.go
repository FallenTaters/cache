package cache

import (
	"time"
)

type addedValue[V any] struct {
	value V
	added time.Time
}

type tfifo[K comparable, V any] struct {
	*fifo[K, addedValue[V]]

	maxAge time.Duration
}

func TFIFO[K comparable, V any](maxEntries int, maxAge time.Duration, addFunc AddFunc[K, V]) Cache[K, V] {
	if addFunc == nil {
		panic(`nil addFunc passed`)
	}

	wrappedAddFunc := func(key K) (addedValue[V], error) {
		v, err := addFunc(key)
		return addedValue[V]{value: v, added: time.Now()}, err
	}

	return tfifo[K, V]{
		fifo:   FIFO(maxEntries, wrappedAddFunc).(*fifo[K, addedValue[V]]),
		maxAge: maxAge,
	}
}

func (t tfifo[K, V]) Get(key K) (V, bool) {
	var empty V

	t.RLock()
	v, ok := t.values[key]
	t.RUnlock()

	if !ok {
		return empty, false
	}

	if time.Since(v.value.added) > t.maxAge {
		t.Delete(key)
		return empty, false
	}

	return v.value.value, true
}

func (t tfifo[K, V]) GetOrEmpty(key K) V {
	v, _ := t.Get(key)
	return v
}

func (t tfifo[K, V]) GetOrAdd(key K) (V, error) {
	var empty V

	v, err := t.fifo.GetOrAdd(key)
	if err != nil {
		return empty, err
	}

	if time.Since(v.added) <= t.maxAge {
		return v.value, nil
	}

	t.Delete(key)
	v, err = t.fifo.GetOrAdd(key)

	return v.value, err
}

func (t tfifo[K, V]) MustGetOrAdd(key K) V {
	v, err := t.GetOrAdd(key)
	if err != nil {
		panic(err)
	}

	return v
}

func (t tfifo[K, V]) Delete(key K) {
	t.Lock()
	defer t.Unlock()

	v, ok := t.values[key]
	if !ok {
		return
	}

	for node := v.element; node != nil; node = node.Next() {
		delete(t.values, node.Value.key)
	}

	t.entries.RemoveAllAfter(v.element)
}
