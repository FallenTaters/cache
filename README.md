# Cache

Simple, generic, threadsafe caches.

[![codecov](https://codecov.io/gh/FallenTaters/cache/branch/master/graph/badge.svg)](https://codecov.io/gh/FallenTaters/cache)

## FIFO

first in, first out, with a limit on the number of key-value pairs

### AddFunc

AddFunc is specified for adding new key-value pairs. The cache is not blocked if addFunc takes a longer time to complete.
If GetOrAdd is called multiple times with the same key while another addFunc is still busy, all calls will wait for the first addFunc to return and use that result.

## TFIFO

* A wrapper around FIFO for time-awareness
* All key-value pairs share a maximum age, **not** specified per pair
* Performance/Memory Note: Does **not** clean itself passively but checks age on access

## Example

```go
package example

import (
	"time"

	"github.com/FallenTaters/cache"
)

const (
	maxEntries = 1000
	maxAge     = time.Hour
)

var itemsCache = cache.TFIFO[int, item](maxEntries, maxAge)

type item struct {
	id    int
	name  string
	price int
}

func getItem(id int) func() (item, error) {
	return func() (item, error) {
		// HTTP call, DB Query, etc.
		return item{}, nil
	}
}

func doStuff() {
	// get item only if cached
	myItem, ok := itemsCache.Get(1)
	if !ok {
		// item not cached currently
	}

	// add something to cache manually
	itemsCache.Add(2, item{})

	// get item from cache, otherwise use AddFunc
	myItem, err := itemsCache.GetOrAdd(1, getItem(1))
	if err != nil {
		// this err comes from addFunc (getItem)
	}

	// panics if addFunc returns err
	myItem = itemsCache.MustGetOrAdd(1, getItem(1))

	// delete from cache
	itemsCache.Delete(1)
}

```

## TODO

* add an LRU (and TLRU) cache
