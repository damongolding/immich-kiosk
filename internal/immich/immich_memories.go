package immich

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"
	"path"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
)

// memories fetches memory lane assets from the Immich API
// requestID is used for request tracking
// deviceID identifies the requesting device
// assetCount determines if we want just the count of assets
// Returns the memory lane response, API URL used, and any error
func (i *ImmichAsset) memories(requestID, deviceID string, assetCount bool) (MemoryLaneResponse, string, error) {
	var memoryLane MemoryLaneResponse

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		return immichApiFail(memoryLane, err, nil, "")
	}

	now := time.Now()

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     path.Join("api", "assets", "memory-lane"),
		RawQuery: fmt.Sprintf("month=%d&day=%d", now.Month(), now.Day()),
	}

	// If we want the memories assets count we will use a seperate cache entry
	// because Kiosk removes used assets from the normal cache entry
	if assetCount {
		apiUrl.RawQuery += "&count=true"
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, deviceID, memoryLane)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		return immichApiFail(memoryLane, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &memoryLane)
	if err != nil {
		return immichApiFail(memoryLane, err, body, apiUrl.String())
	}

	return memoryLane, apiUrl.String(), nil
}

// memoriesCount is the internal implementation of MemoryLaneAssetsCount
// that counts the total assets in memory lane using a separate cache entry
func memoriesCount(memories MemoryLaneResponse) int {
	total := 0
	for _, memory := range memories {
		total += len(memory.Assets)
	}
	return total
}

// MemoryLaneAssetsCount returns the total count of memory lane assets
func (i *ImmichAsset) MemoryLaneAssetsCount(requestID, deviceID string) int {
	m, _, err := i.memories(requestID, deviceID, true)
	if err != nil {
		return 0
	}

	return memoriesCount(m)
}

func updateMemoryCache(memories MemoryLaneResponse, pickedMemoryIndex, assetIndex int, apiCacheKey string) error {
	// Deep copy the memories slice
	assetsToCache := make(MemoryLaneResponse, len(memories))
	for i, memory := range memories {
		assetsToCache[i].YearsAgo = memory.YearsAgo
		assetsToCache[i].Title = memory.Title
		assetsToCache[i].Assets = make([]ImmichAsset, len(memory.Assets))
		copy(assetsToCache[i].Assets, memory.Assets)
	}

	// Remove the current image from the slice
	assetsToCache[pickedMemoryIndex].Assets = append(assetsToCache[pickedMemoryIndex].Assets[:assetIndex], assetsToCache[pickedMemoryIndex].Assets[assetIndex+1:]...)

	if len(assetsToCache[pickedMemoryIndex].Assets) == 0 {
		assetsToCache = append(assetsToCache[:pickedMemoryIndex], assetsToCache[pickedMemoryIndex+1:]...)
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

// RandomMemoryLaneImage retrieves a random image from the memory lane assets
// requestID: Unique identifier for tracking the request
// deviceID: ID of the requesting device
// isPrefetch: Indicates if this is a prefetch request to warm the cache
// Returns error if unable to find a valid image after max retries
func (i *ImmichAsset) RandomMemoryLaneImage(requestID, deviceID string, isPrefetch bool) error {

	for retries := 0; retries < MaxRetries; retries++ {

		memories, apiUrl, err := i.memories(requestID, deviceID, false)
		if err != nil {
			return err
		}

		apiCacheKey := cache.ApiCacheKey(apiUrl, deviceID, requestConfig.SelectedUser)

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

			if !asset.isValidAsset(ImageOnlyAssetTypes) {
				continue
			}

			err := asset.AssetInfo(requestID, deviceID)
			if err != nil {
				log.Error("Failed to get additional asset data", "error", err)
			}

			if asset.containsTag(kiosk.TagSkip) {
				continue
			}

			if requestConfig.Kiosk.Cache {
				if err := updateMemoryCache(memories, pickedMemoryIndex, assetIndex, apiCacheKey); err != nil {
					return err
				}
			}

			asset.Bucket = kiosk.SourceMemories
			asset.MemoryTitle = memories[pickedMemoryIndex].Title

			*i = asset

			return nil
		}

		// no viable assets left in memory lane
		memories[pickedMemoryIndex].Assets = make([]ImmichAsset, 1)
		if err := updateMemoryCache(memories, pickedMemoryIndex, 0, apiCacheKey); err != nil {
			return err
		}

	}
	return fmt.Errorf("no assets found for memories after %d retries", MaxRetries)
}
