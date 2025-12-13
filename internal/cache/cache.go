package cache

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
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

// Set stores a value in the cache under the given key.
// If deviceDuration is less than the defaultExpiration, the default expiration is used.
// Otherwise, the item expires after deviceDuration plus one extra minute.
// If the key already exists, its value is replaced.
func Set(key string, x any, deviceDuration int) {
	if deviceDuration < 0 {
		log.Warn("Negative duration provided, using default expiration", "deviceDuration", deviceDuration)
		kioskCache.Set(key, x, gocache.DefaultExpiration)
		return
	}
	deviceDurationPlusMin := (time.Duration(deviceDuration) * time.Second) + time.Minute
	if deviceDurationPlusMin <= defaultExpiration {
		kioskCache.Set(key, x, gocache.DefaultExpiration)
		return
	}
	SetWithExpiration(key, x, deviceDurationPlusMin)
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

// assetToCache is an internal helper function that handles adding items to the cache array.
// It maintains a limited history size, retrieves existing cached data if available,
// and adds the new item either at the beginning (PREPEND) or end (APPEND) of the array.
// If the cached data is invalid or not found, it initializes a new empty array.
func assetToCache[T any](viewDataToAdd T, requestConfig *config.Config, deviceID, url string, position Position) {
	utils.TrimHistory(&requestConfig.History, kiosk.HistoryLimit)

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

	Set(viewCacheKey, cachedViewData, requestConfig.Duration)
}
