package cache_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/FallenTaters/cache"
	"github.com/FallenTaters/cache/assert"
)

func newFIFO() cache.Cache[int, int] {
	return cache.FIFO[int, int](2)
}

func newAddFunc(x int, err error) func() (int, error) {
	return func() (int, error) {
		return x, err
	}
}

func calledAddFunc(x int, err error) (func() (int, error), *bool) {
	var called bool
	return func() (int, error) {
		called = true
		return 0, nil
	}, &called
}

// total goroutines must be below 8_128 for github
const concurrentCount = 5_000

func TestFIFO(t *testing.T) {
	t.Run(`get non-existing`, func(t *testing.T) {
		c := newFIFO()

		v, ok := c.Get(1)
		assert.False(t, ok)
		assert.Equal(t, 0, v)
	})

	t.Run(`get existing`, func(t *testing.T) {
		c := newFIFO()

		v, err := c.GetOrAdd(1, newAddFunc(1, nil))
		assert.Equal(t, 1, v)
		assert.NoError(t, err)

		v, ok := c.Get(1)
		assert.Equal(t, 1, v)
		assert.True(t, ok)
	})

	t.Run(`only get recently added`, func(t *testing.T) {
		c := newFIFO()

		_, _ = c.GetOrAdd(1, newAddFunc(1, nil))
		_, _ = c.GetOrAdd(2, newAddFunc(2, nil))
		_, _ = c.GetOrAdd(3, newAddFunc(3, nil))

		v, ok := c.Get(1)
		assert.Equal(t, 0, v)
		assert.False(t, ok)

		v, ok = c.Get(2)
		assert.Equal(t, 2, v)
		assert.True(t, ok)

		v, ok = c.Get(3)
		assert.Equal(t, 3, v)
		assert.True(t, ok)
	})

	t.Run(`return error`, func(t *testing.T) {
		myErr := errors.New(`myErr`)
		c := newFIFO()

		_, err := c.GetOrAdd(1, newAddFunc(0, myErr))
		assert.ErrorIs(t, myErr, err)
	})

	t.Run(`panic for must`, func(t *testing.T) {
		myErr := errors.New(`myErr`)
		defer func() {
			v := recover()
			assert.ErrorIs(t, myErr, v.(error))
		}()

		c := newFIFO()

		c.MustGetOrAdd(1, newAddFunc(0, myErr))
	})

	t.Run(`only call addFunc when necessary`, func(t *testing.T) {
		c := newFIFO()

		addFunc, called := calledAddFunc(1, nil)

		c.MustGetOrAdd(1, addFunc)
		assert.True(t, *called)

		*called = false
		c.MustGetOrAdd(1, addFunc)
		assert.False(t, *called)
	})

	t.Run(`delete`, func(t *testing.T) {
		c := newFIFO()

		_, _ = c.GetOrAdd(1, newAddFunc(1, nil))
		v, ok := c.Get(1)
		assert.True(t, ok)
		assert.Equal(t, 1, v)

		c.Delete(1)

		v, ok = c.Get(1)
		assert.False(t, ok)
		assert.Equal(t, 0, v)
	})

	t.Run(`concurrent operations`, func(t *testing.T) {
		c := newFIFO()

		var wg sync.WaitGroup

		wg.Add(concurrentCount)
		go func() {
			for i := 0; i < concurrentCount; i++ {
				expected := i
				go func() {
					actual, err := c.GetOrAdd(expected, newAddFunc(expected, nil))
					assert.Equal(t, expected, actual)
					assert.NoError(t, err)

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
					wg.Done()
				}()
			}
		}()

		wg.Wait()
	})
}
