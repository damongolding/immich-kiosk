package cache

import (
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

var apiCache *gocache.Cache

func init() {
	// Setting up Immich api cache
	apiCache = gocache.New(5*time.Minute, 10*time.Minute)
}

func FlushApiCache() {
	apiCache.Flush()
}

func ApiCacheCount() int {
	return apiCache.ItemCount()
}

func ApiCacheKey(apiUrl, devideID string) string {
	return fmt.Sprintf("%s:%s", apiUrl, devideID)
}

func Get(s string) (interface{}, bool) {
	return apiCache.Get(s)
}
