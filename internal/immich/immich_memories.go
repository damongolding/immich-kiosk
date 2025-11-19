package immich

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/immich_open_api"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/dustin/go-humanize"
)

// MemoriesWithPastDays
// Fetches memories for a given device ID and user ID for a specified number of past days.
// Returns a MemoriesResponse, the URL used for the request, and an error if any occurred.
//
// As the returned MemoriesResponse is a combination of memories from the past days (multiple API calls)
// the cache is managed manually.
func (a *Asset) MemoriesWithPastDays(requestID, deviceID string, days int) (MemoriesResponse, string, error) {
	return a.memoriesWithPastDays(requestID, deviceID, false, days)
}

// MemoriesWithPastDays
// Fetches memories for a given device ID and user ID for a specified number of past days.
// Returns a MemoriesResponse, the URL used for the request, and an error if any occurred.
//
// As the returned MemoriesResponse is a combination of memories from the past days (multiple API calls)
// the cache is managed manually.
func (a *Asset) memoriesWithPastDays(requestID, deviceID string, assetCount bool, days int) (MemoriesResponse, string, error) {
	var memories MemoriesResponse

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(memories, err, nil, "")
	}

	startOfDay, _ := processTodayDateRange()

	apiURL := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     path.Join("api", "memories"),
		RawQuery: fmt.Sprintf("for=%s&pastDays=%d", url.PathEscape(startOfDay.Format("2006-01-02T15:04:05.000Z")), days),
	}

	// If we want the memories assets count we will use a separate cache entry
	// because Kiosk removes used assets from the normal cache entry
	if assetCount {
		apiURL.RawQuery += "&count=true"
	}

	cacheKey := cache.APICacheKey(apiURL.String(), deviceID, a.requestConfig.SelectedUser)

	if apiData, found := cache.Get(cacheKey); found {
		log.Debug(requestID+" Cache hit", "url", apiURL.String())
		data, ok := apiData.([]byte)
		if !ok {
			return memories, apiURL.String(), errors.New("could not parse past memories data")
		}

		err = json.Unmarshal(data, &memories)
		if err != nil {
			return memories, apiURL.String(), err
		}

		return memories, apiURL.String(), nil
	}

	for day := range days {
		// Fetch memories for each day
		m, memURL, memErr := a.memories(requestID, deviceID, false, day)
		if memErr != nil {
			return memories, memURL, memErr
		}

		memories = append(memories, m...)
	}

	b, marshalErr := json.Marshal(memories)
	if marshalErr != nil {
		return memories, apiURL.String(), marshalErr
	}

	cache.Set(cacheKey, b)

	return memories, apiURL.String(), nil
}

func (a *Asset) Memories(requestID, deviceID string) (MemoriesResponse, string, error) {
	return a.memories(requestID, deviceID, false, 0)
}

// memories fetches memory assets from the Immich API.
//
// Parameters:
//   - requestID: Used for request tracking
//   - deviceID: Identifies the requesting device and for caching purposes
//   - assetCount: Determines if we want just the count of assets
//
// Returns:
//   - MemoriesResponse: The memory response data
//   - string: The API URL used for the request
//   - error: Any error that occurred
func (a *Asset) memories(requestID, deviceID string, assetCount bool, days int) (MemoriesResponse, string, error) {
	var memories MemoriesResponse

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(memories, err, nil, "")
	}

	startOfDay, _ := processTodayDateRange()

	if days > 0 {
		startOfDay = startOfDay.AddDate(0, 0, -days)
	}

	apiURL := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     path.Join("api", "memories"),
		RawQuery: fmt.Sprintf("for=%s", url.PathEscape(startOfDay.Format("2006-01-02T15:04:05.000Z"))),
	}

	// If we want the memories assets count we will use a separate cache entry
	// because Kiosk removes used assets from the normal cache entry
	if assetCount {
		apiURL.RawQuery += "&count=true"
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, memories)
	body, _, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
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
func (a *Asset) MemoriesAssetsCount(requestID, deviceID string) int {
	var m MemoriesResponse
	var err error
	pastDays := a.requestConfig.PastMemoryDays

	if pastDays > 0 {
		m, _, err = a.memoriesWithPastDays(requestID, deviceID, true, pastDays)
	} else {
		m, _, err = a.memories(requestID, deviceID, true, 0)
	}

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
func (a *Asset) RandomMemoryAsset(requestID, deviceID string) error {

	for range MaxRetries {

		var memories []Memory
		var apiURL string
		var err error

		if a.requestConfig.PastMemoryDays > 0 {
			memories, apiURL, err = a.MemoriesWithPastDays(requestID, deviceID, a.requestConfig.PastMemoryDays)
		} else {
			memories, apiURL, err = a.Memories(requestID, deviceID)
		}
		if err != nil {
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL, deviceID, a.requestConfig.SelectedUser)

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
			asset.requestConfig = a.requestConfig
			asset.ctx = a.ctx

			// temp fix for memories not being supplied with EXIF
			infoErr := asset.AssetInfo(requestID, deviceID)
			if infoErr != nil {
				log.Error("failed to get asset info", "error", infoErr)
				continue
			}

			if !asset.isValidAsset(requestID, deviceID, ImageOnlyAssetTypes, a.RatioWanted) {
				continue
			}

			if a.requestConfig.Kiosk.Cache {
				if cacheErr := updateMemoryCache(memories, pickedMemoryIndex, assetIndex, apiCacheKey); cacheErr != nil {
					return cacheErr
				}
			}

			if memories[pickedMemoryIndex].Type == immich_open_api.OnThisDay {
				asset.MemoryTitle = humanize.Time(memories[pickedMemoryIndex].Assets[assetIndex].LocalDateTime)
			}

			*a = asset

			return nil
		}

		// no viable assets left in memories
		memories[pickedMemoryIndex].Assets = make([]Asset, 1)
		if cacheErr := updateMemoryCache(memories, pickedMemoryIndex, 0, apiCacheKey); cacheErr != nil {
			return cacheErr
		}

	}
	return fmt.Errorf("no assets found for memories after %d retries", MaxRetries)
}

// IsMemory checks if the asset is part of recent memories by querying the
// memories API with a 5-minute cache window.
//
// Returns:
//   - bool: true if the asset is found in memories
//   - Memory: the memory containing the asset (empty if not found)
//   - int: the index of the asset within the memory (0 if not found)
func (a *Asset) IsMemory() (bool, Memory, int) {

	memLookUp := strconv.FormatInt(time.Now().Unix()/int64(5*60), 10)

	var m []Memory
	var err error

	if a.requestConfig.PastMemoryDays > 0 {
		m, _, err = a.MemoriesWithPastDays(kiosk.GlobalCache, memLookUp, a.requestConfig.PastMemoryDays)
	} else {
		m, _, err = a.Memories(kiosk.GlobalCache, memLookUp)
	}

	if err != nil {
		log.Error("failed to get memories", "error", err)
		return false, Memory{}, 0
	}

	for _, memory := range m {
		for assetIndex, asset := range memory.Assets {
			if a.ID == asset.ID {
				return true, memory, assetIndex
			}
		}
	}

	return false, Memory{}, 0
}
