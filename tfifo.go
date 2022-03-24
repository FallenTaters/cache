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

func TFIFO[K comparable, V any](maxEntries int, maxAge time.Duration) Cache[K, V] {
	return tfifo[K, V]{
		fifo:   FIFO[K, addedValue[V]](maxEntries).(*fifo[K, addedValue[V]]),
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

func (t tfifo[K, V]) Add(key K, value V) {
	t.fifo.Add(key, addedValue[V]{
		value: value,
		added: time.Now(),
	})
}

func (t tfifo[K, V]) GetOrAdd(key K, addFunc AddFunc[V]) (V, error) {
	var empty V
	wrappedAddFunc := wrapAddFunc(addFunc)

	v, err := t.fifo.GetOrAdd(key, wrappedAddFunc)
	if err != nil {
		return empty, err
	}

	if time.Since(v.added) <= t.maxAge {
		return v.value, nil
	}

	t.Delete(key)
	v, err = t.fifo.GetOrAdd(key, wrappedAddFunc)

	return v.value, err
}

func (t tfifo[K, V]) MustGetOrAdd(key K, addFunc AddFunc[V]) V {
	v, err := t.GetOrAdd(key, addFunc)
	if err != nil {
		panic(err)
	}

	return v
}

func wrapAddFunc[V any](addFunc AddFunc[V]) AddFunc[addedValue[V]] {
	return func() (addedValue[V], error) {
		v, err := addFunc()
		return addedValue[V]{value: v, added: time.Now()}, err
	}
}
