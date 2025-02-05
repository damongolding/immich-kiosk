package cache

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/utils"
	gocache "github.com/patrickmn/go-cache"
)

type CachePosition string

const (
	PREPEND CachePosition = "prepend"
	APPEND  CachePosition = "append"
)

// Package cache provides a simple in-memory cache implementation using github.com/patrickmn/go-cache
var (
	kioskCache *gocache.Cache

	defaultExpiration = 5 * time.Minute
	cleanupInterval   = 10 * time.Minute
)

// init initializes the kiosk cache with a 5 minute default expiration and 10 minute cleanup interval.
// The cleanup interval determines how often expired items are removed from the cache.
func init() {
	// Setting up Immich api cache
	kioskCache = gocache.New(defaultExpiration, cleanupInterval)
}

// Flush removes all items from the cache, both expired and unexpired.
// This operation cannot be undone.
func Flush() {
	kioskCache.Flush()
}

// ItemCount returns the number of items currently stored in the cache,
// including both expired and unexpired items.
func ItemCount() int {
	return kioskCache.ItemCount()
}

// ViewCacheKey generates a cache key from the API URL and device ID by combining them
// with ':view' suffix for cache view operations. The key is hashed using SHA-256
// for consistent length and character set.
func ViewCacheKey(apiUrl, deviceID string) string {
	key := fmt.Sprintf("%s:%s:view", apiUrl, deviceID)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

// ApiCacheKey generates a cache key from the API URL and device ID by combining them
// with ':api' suffix for cache API operations. The key is hashed using SHA-256
// for consistent length and character set.
func ApiCacheKey(apiUrl, deviceID string, user string) string {
	key := fmt.Sprintf("%s:%s:%s:api", apiUrl, deviceID, user)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

// Get retrieves an item from the cache by key, returning the item and a boolean indicating
// whether the key was found in the cache. If the key is not found or the item has expired,
// the boolean will be false.
func Get(s string) (any, bool) {
	return kioskCache.Get(s)
}

// Set adds an item to the cache with the default expiration time.
// If the key already exists, its value will be overwritten.
func Set(key string, x any) {
	kioskCache.Set(key, x, gocache.DefaultExpiration)
}

// SetWithExpiration adds an item to the cache with the specified expiration duration.
// The item will expire after the given duration has elapsed. If the key already exists,
// its value and expiration time will be overwritten.
func SetWithExpiration(key string, x any, t time.Duration) {
	kioskCache.Set(key, x, t)
}

// Delete removes an item from the cache by key.
// If the key does not exist, no action is taken.
func Delete(key string) {
	kioskCache.Delete(key)
}

// Replace updates an existing item in the cache with a new value.
// Returns an error if the key does not exist.
func Replace(key string, x any) error {
	return kioskCache.Replace(key, x, gocache.DefaultExpiration)
}

// ReplaceWithExpiration updates an existing item in the cache with a new value and specified expiration time.
// The item will expire after the given duration has elapsed. Returns an error if the key does not exist.
func ReplaceWithExpiration(key string, x any, t time.Duration) error {
	return kioskCache.Replace(key, x, t)
}

func AssetToCache[T any](viewDataToAdd T, requestConfig *config.Config, deviceID, url string) {
	assetToCache(viewDataToAdd, requestConfig, deviceID, url, APPEND)
}

func AssetToCacheWithPosition[T any](viewDataToAdd T, requestConfig *config.Config, deviceID, url string, position CachePosition) {
	assetToCache(viewDataToAdd, requestConfig, deviceID, url, position)
}

func assetToCache[T any](viewDataToAdd T, requestConfig *config.Config, deviceID, url string, position CachePosition) {
	utils.TrimHistory(&requestConfig.History, 10)

	cachedViewData := []T{}

	viewCacheKey := ViewCacheKey(url, deviceID)

	if data, found := Get(viewCacheKey); found {
		if typedData, ok := data.([]T); ok {
			cachedViewData = typedData
		} else {
			log.Error("Invalid cache data type")
			cachedViewData = []T{}
		}
	}

	switch position {
	case APPEND:
		cachedViewData = append(cachedViewData, viewDataToAdd)
	case PREPEND:
		cachedViewData = append([]T{viewDataToAdd}, cachedViewData...)
	}

	Set(viewCacheKey, cachedViewData)
}
