package cache


type Cache[K comparable, V any] interface {
	// Get returns true if the value is found
	Get(K) (V, bool)

	// GetOrEmpty returns the zero value if the value is not found
	GetOrEmpty(K) V

	// GetOrAdd returns the value if found, otherwise runs AddFunc
	GetOrAdd(K) (V, error)

	// MustGetOrAdd is like GetOrAdd, but will panic if AddFunc returns an error
	MustGetOrAdd(K) V

	// Delete removes the cached entry if it exists
	Delete(K)
}

type AddFunc[K comparable, V any] func(K) (V, error)
