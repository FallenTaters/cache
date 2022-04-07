package cache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/FallenTaters/cache"
	"github.com/FallenTaters/cache/assert"
)

const cacheDuration = 10 * time.Millisecond

func newTFIFO() cache.Cache[int, int] {
	return cache.TFIFO[int, int](3, cacheDuration)
}

func newTLRU() cache.Cache[int, int] {
	return cache.TLRU[int, int](3, cacheDuration)
}

func sleep() {
	time.Sleep(cacheDuration)
}

func TestTFIFO(t *testing.T) {
	t.Run(`get non-existing`, func(t *testing.T) {
		c := newTLRU()

		v, ok := c.Get(1)
		assert.False(t, ok)
		assert.Equal(t, 0, v)
	})

	t.Run(`add and get`, func(t *testing.T) {
		c := newTLRU()

		c.Add(1, 1)

		v, ok := c.Get(1)
		assert.True(t, ok)
		assert.Equal(t, 1, v)
	})

	t.Run(`return error`, func(t *testing.T) {
		myErr := errors.New(`myErr`)
		c := newTLRU()

		_, err := c.GetOrAdd(1, newAddFunc(0, myErr))
		assert.ErrorIs(t, myErr, err)
	})

	t.Run(`panic for must`, func(t *testing.T) {
		myErr := errors.New(`myErr`)
		defer func() {
			v := recover()
			assert.ErrorIs(t, myErr, v.(error))
		}()

		c := newTLRU()

		c.MustGetOrAdd(1, newAddFunc(0, myErr))
	})

	t.Run(`expiring`, func(t *testing.T) {
		c := newTFIFO()

		addFunc, called := calledAddFunc(1, nil)

		// Get fails for expired
		_ = c.MustGetOrAdd(1, addFunc)
		_, ok := c.Get(1)
		assert.True(t, ok)
		sleep()
		_, ok = c.Get(1)
		assert.False(t, ok)

		// GetOrAdd calls onlyu after expire
		_ = c.MustGetOrAdd(1, addFunc)
		*called = false
		_ = c.MustGetOrAdd(1, addFunc)
		assert.False(t, *called)
		sleep()
		_ = c.MustGetOrAdd(1, addFunc)
		assert.True(t, *called)
	})
}
