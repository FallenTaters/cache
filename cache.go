package cache

type Cache[K comparable, V any] interface {
	// Get returns true if the value is found
	Get(K) (V, bool)

	// Add adds the value to the cache
	Add(K, V)

	// GetOrAdd returns the value if found, otherwise runs AddFunc
	// if AddFunc returns an error, GetOrAdd returns it
	GetOrAdd(K, AddFunc[V]) (V, error)

	// MustGetOrAdd is like GetOrAdd, but will panic if AddFunc returns an error
	MustGetOrAdd(K, AddFunc[V]) V

	// Delete removes the cached entry if it exists
	Delete(K)
}

type AddFunc[V any] func() (V, error)
