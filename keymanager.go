package cache

import (
	"sync"
)

type result[V any] struct {
	Value V
	Err   error
}

type addManager[K comparable, V any] struct {
	sync.RWMutex

	busyKeys map[K][]chan result[V]
}

func (km *addManager[K, V]) waitOrAdd(key K, addFunc AddFunc[V]) <-chan result[V] {
	km.Lock()
	defer km.Unlock()

	ch := make(chan result[V])

	channels, ok := km.busyKeys[key]
	if ok {
		// wait for existing addFunc
		km.busyKeys[key] = append(channels, ch)
		return ch
	}

	// start a new addFunc
	km.busyKeys[key] = []chan result[V]{ch}
	go func() {
		v, err := addFunc()
		result := result[V]{v, err}

		km.Lock()
		defer km.Unlock()
		for _, ch := range km.busyKeys[key] {
			ch <- result
			close(ch)
		}

		delete(km.busyKeys, key)
	}()

	return ch
}
