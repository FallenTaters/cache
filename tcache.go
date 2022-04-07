package cache

import "time"

func TLRU[K comparable, V any](maxEntries int, maxAge time.Duration) Cache[K, V] {
	return tCache[K, V]{
		cache:  LRU[K, addedValue[V]](maxEntries).(*lru[K, addedValue[V]]),
		maxAge: maxAge,
	}
}

func TFIFO[K comparable, V any](maxEntries int, maxAge time.Duration) Cache[K, V] {
	return tCache[K, V]{
		cache:  FIFO[K, addedValue[V]](maxEntries).(*fifo[K, addedValue[V]]),
		maxAge: maxAge,
	}
}

type addedValue[V any] struct {
	value V
	added time.Time
}

func wrapAddFunc[V any](addFunc AddFunc[V]) AddFunc[addedValue[V]] {
	return func() (addedValue[V], error) {
		v, err := addFunc()
		return addedValue[V]{value: v, added: time.Now()}, err
	}
}

type cache[K comparable, V any] interface {
	Cache[K, V]

	RLock()
	RUnlock()
	Lock()
	Unlock()
}

type tCache[K comparable, V any] struct {
	cache[K, addedValue[V]]

	maxAge time.Duration
}

func (t tCache[K, V]) Get(key K) (V, bool) {
	var empty V

	v, ok := t.cache.Get(key)
	if !ok {
		return empty, false
	}

	if time.Since(v.added) > t.maxAge {
		t.Delete(key)
		return empty, false
	}

	return v.value, true
}

func (t tCache[K, V]) Add(key K, value V) {
	t.cache.Add(key, addedValue[V]{
		value: value,
		added: time.Now(),
	})
}

func (t tCache[K, V]) GetOrAdd(key K, addFunc AddFunc[V]) (V, error) {
	var empty V
	wrappedAddFunc := wrapAddFunc(addFunc)

	v, err := t.cache.GetOrAdd(key, wrappedAddFunc)
	if err != nil {
		return empty, err
	}

	if time.Since(v.added) <= t.maxAge {
		return v.value, nil
	}

	t.Delete(key)
	v, err = t.cache.GetOrAdd(key, wrappedAddFunc)

	return v.value, err
}

func (t tCache[K, V]) MustGetOrAdd(key K, addFunc AddFunc[V]) V {
	v, err := t.GetOrAdd(key, addFunc)
	if err != nil {
		panic(err)
	}

	return v
}
