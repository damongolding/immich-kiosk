package immich

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"path"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/immich_open_api"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/dustin/go-humanize"
)

// memories fetches memory assets from the Immich API.
//
// Parameters:
//   - requestID: Used for request tracking
//   - deviceID: Identifies the requesting device
//   - assetCount: Determines if we want just the count of assets
//
// Returns:
//   - MemoriesResponse: The memory response data
//   - string: The API URL used for the request
//   - error: Any error that occurred
func (i *Asset) memories(requestID, deviceID string, assetCount bool) (MemoriesResponse, string, error) {
	var memories MemoriesResponse

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(memories, err, nil, "")
	}

	startOfToday, _ := processTodayDateRange()

	apiURL := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     path.Join("api", "memories"),
		RawQuery: fmt.Sprintf("for=%s", url.PathEscape(startOfToday.Format("2006-01-02T15:04:05.000Z"))),
	}

	// If we want the memories assets count we will use a seperate cache entry
	// because Kiosk removes used assets from the normal cache entry
	if assetCount {
		apiURL.RawQuery += "&count=true"
	}

	immichAPICall := withImmichAPICache(i.immichAPICall, requestID, deviceID, i.requestConfig, memories)
	body, err := immichAPICall(http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return immichAPIFail(memories, err, body, apiURL.String())
	}

	err = json.Unmarshal(body, &memories)
	if err != nil {
		return immichAPIFail(memories, err, body, apiURL.String())
	}

	return memories, apiURL.String(), nil
}

// memoriesCount counts the total number of assets in memories.
// It iterates through all memories and sums up their assets.
func memoriesCount(memories MemoriesResponse) int {
	total := 0

	for _, memory := range memories {
		total += len(memory.Assets)
	}

	return total
}

// MemoriesAssetsCount returns the total count of memory assets.
//
// Parameters:
//   - requestID: Request tracking identifier
//   - deviceID: Device identifier
//
// Returns:
//   - int: Total number of assets, or 0 if error occurs
func (i *Asset) MemoriesAssetsCount(requestID, deviceID string) int {
	m, _, err := i.memories(requestID, deviceID, true)
	if err != nil {
		return 0
	}

	return memoriesCount(m)
}

// updateMemoryCache updates the cache by removing used assets from memories.
//
// Parameters:
//   - memories: Current memories response
//   - pickedMemoryIndex: Index of selected memory
//   - assetIndex: Index of asset within memory
//   - cache.ApiCacheKey: Cache key for API response
//
// Returns:
//   - error: Any error during cache update
func updateMemoryCache(memories MemoriesResponse, pickedMemoryIndex, assetIndex int, apiCacheKey string) error {

	// Deep copy the memories slice
	assetsToCache := make(MemoriesResponse, len(memories))
	for i, memory := range memories {
		assetsToCache[i] = memory
		assetsToCache[i].Assets = make([]Asset, len(memory.Assets))
		copy(assetsToCache[i].Assets, memory.Assets)
	}

	// Remove the current image from the slice
	assetsToCache[pickedMemoryIndex].Assets = slices.Delete(assetsToCache[pickedMemoryIndex].Assets, assetIndex, assetIndex+1)

	if len(assetsToCache[pickedMemoryIndex].Assets) == 0 {
		assetsToCache = slices.Delete(assetsToCache, pickedMemoryIndex, pickedMemoryIndex+1)
	}

	jsonBytes, err := json.Marshal(assetsToCache)
	if err != nil {
		log.Error("Failed to marshal assetsToCache", "error", err)
		return err
	}

	// replace with cache minus used asset
	err = cache.Replace(apiCacheKey, jsonBytes)
	if err != nil {
		log.Debug("Failed to update device cache for memories")
	}

	return nil
}

// RandomMemoryAsset retrieves a random image from memory assets.
//
// Parameters:
//   - requestID: Unique identifier for tracking the request
//   - deviceID: ID of the requesting device
//   - isPrefetch: Indicates if this is a prefetch request to warm the cache
//
// Returns:
//   - error: If unable to find valid image after max retries
func (i *Asset) RandomMemoryAsset(requestID, deviceID string) error {

	for range MaxRetries {

		memories, apiURL, err := i.memories(requestID, deviceID, false)
		if err != nil {
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL, deviceID, i.requestConfig.SelectedUser)

		if len(memories) == 0 {
			log.Debug(requestID + " No images left in cache. Refreshing and trying again for memories")
			cache.Delete(apiCacheKey)
			continue
		}

		pickedMemoryIndex := rand.IntN(len(memories))

		rand.Shuffle(len(memories[pickedMemoryIndex].Assets), func(i, j int) {
			memories[pickedMemoryIndex].Assets[i], memories[pickedMemoryIndex].Assets[j] = memories[pickedMemoryIndex].Assets[j], memories[pickedMemoryIndex].Assets[i]
		})

		for assetIndex, asset := range memories[pickedMemoryIndex].Assets {

			asset.Bucket = kiosk.SourceMemories
			asset.requestConfig = i.requestConfig

			if !asset.isValidAsset(requestID, deviceID, ImageOnlyAssetTypes, i.RatioWanted) {
				continue
			}

			if i.requestConfig.Kiosk.Cache {
				if err := updateMemoryCache(memories, pickedMemoryIndex, assetIndex, apiCacheKey); err != nil {
					return err
				}
			}

			if memories[pickedMemoryIndex].Type == immich_open_api.OnThisDay {
				asset.MemoryTitle = humanize.Time(memories[pickedMemoryIndex].Assets[assetIndex].LocalDateTime)
			}

			*i = asset

			return nil
		}

		// no viable assets left in memories
		memories[pickedMemoryIndex].Assets = make([]Asset, 1)
		if err := updateMemoryCache(memories, pickedMemoryIndex, 0, apiCacheKey); err != nil {
			return err
		}

	}
	return fmt.Errorf("no assets found for memories after %d retries", MaxRetries)
}
