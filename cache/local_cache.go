// Package cache provides a simple thread-safe in-memory cache for storing
// arbitrary values. It uses a global cache store with read-write mutex
// protection for concurrent access.
//
// # Usage
//
// Define cache keys as typed constants:
//
//	const (
//	    UserCacheKey    cache.LocalCacheKey = "user"
//	    SessionCacheKey cache.LocalCacheKey = "session"
//	)
//
// Store and retrieve values:
//
//	cache.LCSet(UserCacheKey, &User{ID: "123"})
//	if value, ok := cache.LCGet(UserCacheKey); ok {
//	    user := value.(*User)
//	}
package cache

import "sync"

// LocalCacheKey is a typed string for cache keys.
// Using typed keys helps prevent key collisions and provides better code clarity.
type LocalCacheKey string

var lcache map[LocalCacheKey]any
var llock sync.RWMutex

func init() {
	lcache = make(map[LocalCacheKey]any)
}

// LCSet stores a value in the cache with the given key.
// If a value already exists for the key, it will be overwritten.
// This operation is thread-safe.
func LCSet(key LocalCacheKey, value any) {
	llock.Lock()
	defer llock.Unlock()
	lcache[key] = value
}

// LCGet retrieves a value from the cache by key.
// Returns the value and true if found, or nil and false if not found.
// This operation is thread-safe.
func LCGet(key LocalCacheKey) (any, bool) {
	llock.RLock()
	defer llock.RUnlock()
	v, ok := lcache[key]
	return v, ok
}

// LCHas checks if a key exists in the cache.
// Returns true if the key exists, false otherwise.
// This operation is thread-safe.
func LCHas(key LocalCacheKey) bool {
	llock.RLock()
	defer llock.RUnlock()
	_, ok := lcache[key]
	return ok
}

// LCDel removes a value from the cache by key.
// If the key does not exist, this operation is a no-op.
// This operation is thread-safe.
func LCDel(key LocalCacheKey) {
	llock.Lock()
	defer llock.Unlock()
	delete(lcache, key)
}
