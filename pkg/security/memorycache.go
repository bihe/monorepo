package security

import (
	"sync"
	"time"
)

// copied and derived from https://github.com/goenning/go-cache-demo
// see https://github.com/goenning/go-cache-demo/blob/master/LICENSE
// initially created by: https://github.com/goenning

// NewMemCache create a cache with the given TTL
func NewMemCache(duration time.Duration) *MemoryCache {
	return &MemoryCache{
		items:         make(map[string]cacheItem),
		cacheDuration: duration,
	}
}

// --------------------------------------------------------------------------
// memoryCache to store User objects
// --------------------------------------------------------------------------

// MemoryCache implements a simple cache
type MemoryCache struct {
	sync.Mutex
	items         map[string]cacheItem
	cacheDuration time.Duration
}

// Get returns an User object by the given key
func (s *MemoryCache) Get(key string) *User {
	s.Lock()
	defer s.Unlock()

	item := s.items[key]
	if item.expired() {
		delete(s.items, key)
		return nil
	}
	return item.user
}

// Set puts an User object into the cache
func (s *MemoryCache) Set(key string, user *User) {
	s.Lock()
	defer s.Unlock()

	s.items[key] = cacheItem{
		user:       user,
		expiration: time.Now().Add(s.cacheDuration).UnixNano(),
	}
}

// --------------------------------------------------------------------------
// cacheItem holding the data
// --------------------------------------------------------------------------

type cacheItem struct {
	user       *User
	expiration int64
}

func (item cacheItem) expired() bool {
	if item.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.expiration
}
