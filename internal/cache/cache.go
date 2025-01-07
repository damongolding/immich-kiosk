package cache

import (
	"fmt"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

var (
	kioskCache *gocache.Cache
	mu         sync.RWMutex
)

// init initializes the kiosk cache with a 5 minute default expiration and 10 minute cleanup interval
func init() {
	// Setting up Immich api cache
	kioskCache = gocache.New(5*time.Minute, 10*time.Minute)
}

// Flush removes all items from the cache
func Flush() {
	mu.Lock()
	defer mu.Unlock()

	kioskCache.Flush()
}

// ItemCount returns the number of items in the cache
func ItemCount() int {
	mu.RLock()
	defer mu.RUnlock()

	return kioskCache.ItemCount()
}

// ViewCacheKey generates a cache key from the API URL and device ID by combining them
// with ':view' suffix for cache view operations
func ViewCacheKey(apiUrl, deviceID string) string {
	return fmt.Sprintf("%s:%s:view", apiUrl, deviceID)
}

// ApiCacheKey generates a cache key from the API URL and device ID by combining them
// with ':api' suffix for cache API operations
func ApiCacheKey(apiUrl, deviceID string) string {
	return fmt.Sprintf("%s:%s:api", apiUrl, deviceID)
}

// Get retrieves an item from the cache by key, returning the item and whether it was found
func Get(s string) (any, bool) {
	mu.RLock()
	defer mu.RUnlock()

	return kioskCache.Get(s)
}

// Set adds an item to the cache with the default expiration time
func Set(key string, x any) {
	mu.Lock()
	defer mu.Unlock()

	kioskCache.Set(key, x, gocache.DefaultExpiration)
}

// SetWithExpiration adds an item to the cache with the specified expiration duration.
// The item will expire after the given duration has elapsed.
func SetWithExpiration(key string, x any, t time.Duration) {
	mu.Lock()
	defer mu.Unlock()

	kioskCache.Set(key, x, t)
}

// Delete removes an item from the cache by key
func Delete(key string) {
	mu.Lock()
	defer mu.Unlock()

	kioskCache.Delete(key)
}

// Replace updates an existing item in the cache with a new value
func Replace(key string, x any) error {
	mu.Lock()
	defer mu.Unlock()

	return kioskCache.Replace(key, x, gocache.DefaultExpiration)
}
