package cache_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FallenTaters/cache"
	"github.com/FallenTaters/cache/assert"
)

const cacheDuration = 10 * time.Millisecond

func newTFIFO() cache.Cache[int, int] {
	return cache.TFIFO[int, int](3, cacheDuration)
}

func sleep() {
	time.Sleep(cacheDuration)
}

func TestTFIFO(t *testing.T) {
	t.Run(`get non-existing`, func(t *testing.T) {
		c := newTFIFO()

		v, ok := c.Get(1)
		assert.False(t, ok)
		assert.Equal(t, 0, v)
	})

	t.Run(`get existing`, func(t *testing.T) {
		c := newTFIFO()

		v, err := c.GetOrAdd(1, newAddFunc(1, nil))
		assert.Equal(t, 1, v)
		assert.NoError(t, err)

		v, ok := c.Get(1)
		assert.Equal(t, 1, v)
		assert.True(t, ok)
	})

	t.Run(`only get recently added`, func(t *testing.T) {
		c := newTFIFO()

		c.GetOrAdd(1, newAddFunc(1, nil))
		c.GetOrAdd(2, newAddFunc(2, nil))
		c.GetOrAdd(3, newAddFunc(3, nil))
		c.GetOrAdd(4, newAddFunc(4, nil))

		v, ok := c.Get(1)
		assert.Equal(t, 0, v)
		assert.False(t, ok)

		v, ok = c.Get(2)
		assert.Equal(t, 2, v)
		assert.True(t, ok)

		v, ok = c.Get(3)
		assert.Equal(t, 3, v)
		assert.True(t, ok)

		v, ok = c.Get(4)
		assert.Equal(t, 4, v)
		assert.True(t, ok)
	})

	t.Run(`return error`, func(t *testing.T) {
		myErr := errors.New(`myErr`)
		c := cache.TFIFO[int, int](3, cacheDuration)

		_, err := c.GetOrAdd(1, newAddFunc(0, myErr))
		assert.ErrorIs(t, myErr, err)
	})

	t.Run(`panic for must`, func(t *testing.T) {
		myErr := errors.New(`myErr`)
		defer func() {
			v := recover()
			assert.ErrorIs(t, myErr, v.(error))
		}()

		c := newTFIFO()

		c.MustGetOrAdd(1, newAddFunc(0, myErr))
	})

	t.Run(`only call addFunc when necessary`, func(t *testing.T) {
		c := cache.TFIFO[int, int](3, cacheDuration)

		addFunc, called := calledAddFunc(1, nil)

		c.MustGetOrAdd(1, addFunc)
		assert.True(t, *called)

		*called = false
		c.MustGetOrAdd(1, addFunc)
		assert.False(t, *called)
	})

	t.Run(`delete`, func(t *testing.T) {
		c := newTFIFO()

		c.GetOrAdd(3, newAddFunc(3, nil))
		c.GetOrAdd(2, newAddFunc(2, nil))
		c.GetOrAdd(1, newAddFunc(1, nil))
		v, ok := c.Get(1)
		assert.True(t, ok)
		assert.Equal(t, 1, v)

		c.Delete(2)

		v, ok = c.Get(1)
		assert.True(t, ok)
		assert.Equal(t, 1, v)

		v, ok = c.Get(2)
		assert.False(t, ok)
		assert.Equal(t, 0, v)

		v, ok = c.Get(3)
		assert.True(t, ok)
		assert.Equal(t, 3, v)
	})

	t.Run(`concurrent operations`, func(t *testing.T) {
		c := newTFIFO()

		var wg sync.WaitGroup
		var count int32

		wg.Add(concurrentCount)
		go func() {
			for i := 0; i < concurrentCount; i++ {
				expected := i
				go func() {
					actual, err := c.GetOrAdd(expected, newAddFunc(expected, nil))
					assert.Equal(t, expected, actual)
					assert.NoError(t, err)

					atomic.AddInt32(&count, 1)
					wg.Done()
				}()
			}
		}()

		wg.Add(concurrentCount)
		go func() {
			for i := concurrentCount - 1; i >= 0; i-- {
				expected := i
				go func() {
					c.Delete(expected)

					atomic.AddInt32(&count, 1)
					wg.Done()
				}()
			}
		}()

		wg.Wait()
	})

	t.Run(`expiring`, func(t *testing.T) {
		c := cache.TFIFO[int, int](3, cacheDuration)

		addFunc, called := calledAddFunc(1, nil)

		// Get fails for expired
		c.GetOrAdd(1, addFunc)
		_, ok := c.Get(1)
		assert.True(t, ok)
		sleep()
		_, ok = c.Get(1)
		assert.False(t, ok)

		// GetOrAdd calls onlyu after expire
		c.GetOrAdd(1, addFunc)
		*called = false
		c.GetOrAdd(1, addFunc)
		assert.False(t, *called)
		sleep()
		c.GetOrAdd(1, addFunc)
		assert.True(t, *called)
	})
}
