package cache

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"sync"
	"time"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/utils"
	gocache "github.com/patrickmn/go-cache"
)

type Position string

const (
	PREPEND Position = "prepend"
	APPEND  Position = "append"
)

// Package cache provides a simple in-memory cache implementation using github.com/patrickmn/go-cache
var (
	kioskCache *gocache.Cache

	defaultExpiration = 5 * time.Minute
	cleanupInterval   = 10 * time.Minute

	DemoMode = false
)

// initialize sets up the kiosk cache based on the current mode:
// - In Demo Mode: Uses 1 minute expiration and 2 minute cleanup interval
// - In Normal Mode: Uses default 5 minute expiration and 10 minute cleanup interval
//
// The expiration time determines when items are considered stale and should be removed.
// The cleanup interval determines how frequently the cache is scanned to remove expired items.
func Initialize() {
	// Setting up Immich api cache
	if DemoMode {
		kioskCache = gocache.New(time.Minute, 2*time.Minute)
	} else {
		kioskCache = gocache.New(defaultExpiration, cleanupInterval)
	}
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
func ViewCacheKey(apiURL, deviceID string) string {
	dateStamp := time.Now().Local().Format(time.DateOnly)
	key := fmt.Sprintf("%s:%s:view:%s", apiURL, deviceID, dateStamp)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

// APICacheKey generates a cache key from the API URL and device ID by combining them
// with ':api' suffix for cache API operations. The key is hashed using SHA-256
// for consistent length and character set.
func APICacheKey(apiURL, deviceID string, user string) string {
	dateStamp := time.Now().Local().Format(time.DateOnly)
	key := fmt.Sprintf("%s:%s:%s:api:%s", apiURL, deviceID, user, dateStamp)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

// Get retrieves an item from the cache by key, returning the item and a boolean indicating
// whether the key was found in the cache. If the key is not found or the item has expired,
// the boolean will be false.
func Get(s string) (any, bool) {
	return kioskCache.Get(s)
}

// Set stores a value in the cache under the given key, replacing any existing entry.
// The expiration is determined by taking the larger of deviceDuration and cacheDuration
// (each extended by one minute), with defaultExpiration used as a lower bound.
// If either duration is negative, gocache.DefaultExpiration is used and a warning is logged.
func Set(key string, value any, deviceDuration, cacheDuration int) {
	if deviceDuration < 0 || cacheDuration < 0 {
		log.Warn("Negative duration or cache duration provided, using default expiration", "deviceDuration", deviceDuration, "cacheDuration", cacheDuration)
		kioskCache.Set(key, value, gocache.DefaultExpiration)
		return
	}

	deviceDurationPlusMin := (time.Duration(deviceDuration) * time.Second) + time.Minute
	cacheDurationPlusMin := (time.Duration(cacheDuration) * time.Second) + time.Minute

	d := max(deviceDurationPlusMin, cacheDurationPlusMin, defaultExpiration)

	kioskCache.Set(key, value, d)
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

// AssetToCache adds a new item of type T to the cache array by appending it to the end.
// It uses the device ID and URL to generate a unique cache key for storing view-related data.
func AssetToCache[T any](viewDataToAdd T, requestConfig *config.Config, deviceID, url string) {
	assetToCache(viewDataToAdd, requestConfig, deviceID, url, APPEND)
}

// AssetToCacheWithPosition adds a new item of type T to the cache array at the specified position
// (either PREPEND or APPEND). It uses the device ID and URL to generate a unique cache key
// for storing view-related data.
func AssetToCacheWithPosition[T any](viewDataToAdd T, requestConfig *config.Config, deviceID, url string, position Position) {
	assetToCache(viewDataToAdd, requestConfig, deviceID, url, position)
}

// viewMu serialises read-modify-write updates to view queues. go-cache locks
// individual Get/Set calls, but not the gap between them, so without this
// concurrent prefetches, video-ready prepends and pops for the same device
// can overwrite each other: dropping queued assets or showing one twice.
var viewMu sync.Mutex

// assetToCache is an internal helper function that handles adding items to the cache array.
// It maintains a limited history size, retrieves existing cached data if available,
// and adds the new item either at the beginning (PREPEND) or end (APPEND) of the array.
// If the cached data is invalid or not found, it initializes a new empty array.
func assetToCache[T any](viewDataToAdd T, requestConfig *config.Config, deviceID, url string, position Position) {
	utils.TrimHistory(&requestConfig.History, kiosk.HistoryLimit)

	viewCacheKey := ViewCacheKey(url, deviceID)

	viewMu.Lock()
	defer viewMu.Unlock()

	var existing []T
	if data, found := Get(viewCacheKey); found {
		if typedData, ok := data.([]T); ok {
			existing = typedData
		} else {
			log.Error("Invalid cache data type")
		}
	}

	// Always build a new slice: the cached one may share its backing array
	// with a slice a reader is still holding.
	updated := make([]T, 0, len(existing)+1)
	switch position {
	case PREPEND:
		updated = append(updated, viewDataToAdd)
		updated = append(updated, existing...)
	default:
		updated = append(updated, existing...)
		updated = append(updated, viewDataToAdd)
	}

	Set(viewCacheKey, updated, requestConfig.Duration, requestConfig.CacheDuration)
}

// PopView removes and returns the first item in the view queue for url and
// deviceID. It returns false if the queue is missing, empty or holds a
// different type (in which case the bad entry is removed). The key is
// deleted once the queue is drained.
func PopView[T any](requestConfig *config.Config, deviceID, url string) (T, bool) {
	var zero T
	key := ViewCacheKey(url, deviceID)

	viewMu.Lock()
	defer viewMu.Unlock()

	data, found := Get(key)
	if !found {
		return zero, false
	}

	queue, ok := data.([]T)
	if !ok {
		log.Error("cache: invalid view data type", "type", fmt.Sprintf("%T", data), "key", key)
		Delete(key)
		return zero, false
	}

	if len(queue) == 0 {
		Delete(key)
		return zero, false
	}

	if len(queue) == 1 {
		Delete(key)
	} else {
		Set(key, slices.Clone(queue[1:]), requestConfig.Duration, requestConfig.CacheDuration)
	}

	return queue[0], true
}
