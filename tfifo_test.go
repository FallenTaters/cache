package cache_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/FallenTaters/cache"
	"github.com/FallenTaters/cache/assert"
)

const cacheDuration = 10 * time.Millisecond

func newTFIFO() cache.Cache[int, int] {
	return cache.TFIFO(2, cacheDuration, func(i int) (int, error) {
		return i, nil
	})
}

func sleep() {
	time.Sleep(cacheDuration)
}

func TestTFIFO(t *testing.T) {
	t.Run(`nil func passend`, func(t *testing.T) {
		defer func() {
			v := recover()
			assert.Equal(t, `nil addFunc passed`, v.(string))
		}()

		cache.TFIFO[int, int](1, cacheDuration, nil)
	})

	t.Run(`get non-existing`, func(t *testing.T) {
		c := newTFIFO()

		v, ok := c.Get(1)
		assert.False(t, ok)
		assert.Equal(t, 0, v)
	})

	t.Run(`get existing`, func(t *testing.T) {
		c := newTFIFO()

		v, err := c.GetOrAdd(1)
		assert.Equal(t, 1, v)
		assert.NoError(t, err)

		v, ok := c.Get(1)
		assert.Equal(t, 1, v)
		assert.True(t, ok)
	})

	t.Run(`only get recently added`, func(t *testing.T) {
		c := newTFIFO()

		c.GetOrAdd(1)
		c.GetOrAdd(2)
		c.GetOrAdd(3)

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
		c := cache.TFIFO(3, cacheDuration, func(i int) (int, error) {
			return 0, myErr
		})

		_, err := c.GetOrAdd(1)
		assert.ErrorIs(t, myErr, err)
	})

	t.Run(`panic for must`, func(t *testing.T) {
		myErr := errors.New(`myErr`)
		defer func() {
			v := recover()
			assert.ErrorIs(t, myErr, v.(error))
		}()

		c := cache.TFIFO(3, cacheDuration, func(i int) (int, error) {
			return 0, myErr
		})

		c.MustGetOrAdd(1)
	})

	t.Run(`only call addFunc when necessary`, func(t *testing.T) {
		called := false
		c := cache.TFIFO(3, cacheDuration, func(i int) (int, error) {
			called = true
			return i, nil
		})

		c.MustGetOrAdd(1)
		assert.True(t, called)

		called = false
		c.MustGetOrAdd(1)
		assert.False(t, called)
	})

	t.Run(`get empty`, func(t *testing.T) {
		c := newTFIFO()

		c.GetOrAdd(1)

		v := c.GetOrEmpty(1)
		assert.Equal(t, 1, v)

		v = c.GetOrEmpty(2)
		assert.Equal(t, 0, v)
	})

	t.Run(`delete`, func(t *testing.T) {
		c := newTFIFO()

		c.GetOrAdd(3)
		c.GetOrAdd(2)
		c.GetOrAdd(1)
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

		_, ok = c.Get(3)
		assert.False(t, ok)
		assert.Equal(t, 0, v)
	})

	t.Run(`concurrent operations`, func(t *testing.T) {
		c := newTFIFO()

		var wg sync.WaitGroup
		count := 0

		wg.Add(concurrentCount)
		go func() {
			for i := 0; i < concurrentCount; i++ {
				expected := i
				go func() {
					actual, err := c.GetOrAdd(expected)
					assert.Equal(t, expected, actual)
					assert.NoError(t, err)

					count++
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

					count++
					wg.Done()
				}()
			}
		}()

		wg.Wait()
	})

	t.Run(`expiring`, func(t *testing.T) {
		called := false
		c := cache.TFIFO(3, cacheDuration, func(i int) (int, error) {
			called = true
			return i, nil
		})

		// Get fails for expired
		c.GetOrAdd(1)
		_, ok := c.Get(1)
		assert.True(t, ok)
		sleep()
		_, ok = c.Get(1)
		assert.False(t, ok)

		// GetOrAdd calls onlyu after expire
		c.GetOrAdd(1)
		called = false
		c.GetOrAdd(1)
		assert.False(t, called)
		sleep()
		c.GetOrAdd(1)
		assert.True(t, called)
	})
}
