# Cache

Simple, generic, threadsafe caches.

## FIFO

first in, first out, with a limit on the number of key-value pairs

### AddFunc

AddFunc is specified for adding new key-value pairs. The cache is not blocked if addFunc takes a longer time to complete. However, there is currently no functionality that prevents simultaneous calls of AddFunc with the same key.

## TFIFO

* A wrapper around FIFO for time-awareness
* All key-value pairs share a maximum age, **not** per pair
* Performance/Memory Note: Does **not** clean itself passively but checks age on access

## Example

```go
package example

import "github.com/FallenTaters/cache"

const (
	maxEntries = 1000
    maxAge = time.Hour
)

var itemsCache = cache.TFIFO(maxEntries, maxAge, getItem)

type item struct {
	id    int
	name  string
	price int
}

func getItem(id int) (item, error) {
	// HTTP call, DB Query, etc.
	return item{}, nil
}

func doStuff() {
	item, ok := itemsCache.Get(1)
	if !ok {
		// item not cached currently
	}

	item, err = itemsCache.GetOrAdd(1)
	if err != nil {
		// this err comes from addFunc (getItem)
	}

	// panics if addFunc returns err
	item = itemsCache.MustGetOrAdd(1)

	// delete from cache
	itemsCache.Delete(1)
}

```

## TODO

* add an LRU and TLRU cache
* make resistant to cache stampede
