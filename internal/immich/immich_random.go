package immich

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/google/go-querystring/query"
)

// RandomImage fetches a random image from the Immich API while handling caching and retries.
//
// This function performs the following:
// - Makes an API request to get random images based on configured parameters
// - Caches results to optimize subsequent requests
// - Filters images based on type, trash/archive status, and aspect ratio
// - Retries up to MaxRetries times if no suitable images are found
// - Updates the cache to remove used images
//
// Parameters:
//   - requestID: Unique identifier for tracking and logging
//   - deviceID: ID of the device making the request
//   - isPrefetch: Indicates if this is a prefetch request
//
// Returns an error if no suitable image is found after retries or if there
// are any issues with API calls, caching, or image processing.
func (a *Asset) RandomImage(requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random image", true)
	} else {
		log.Debug(requestID + " Getting Random image")
	}

	for range MaxRetries {

		var immichAssets []Asset

		u, err := url.Parse(a.requestConfig.ImmichURL)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, nil, "")
			return err
		}

		requestBody := SearchRandomBody{
			Type:       string(ImageType),
			WithExif:   true,
			WithPeople: true,
			Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
		}

		if a.requestConfig.ShowArchived {
			requestBody.WithArchived = true
		}

		DateFilter(&requestBody, a.requestConfig.DateFilter)

		// convert body to queries so url is unique and can be cached
		queries, _ := query.Values(requestBody)

		apiURL := url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     "api/search/random",
			RawQuery: fmt.Sprintf("kiosk=%x", sha256.Sum256([]byte(queries.Encode()))),
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, nil, "")
			return err
		}

		immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, immichAssets)
		apiBody, _, err := immichAPICall(a.ctx, http.MethodPost, apiURL.String(), jsonBody)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, apiBody, apiURL.String())
			return err
		}

		err = json.Unmarshal(apiBody, &immichAssets)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, apiBody, apiURL.String())
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL.String(), deviceID, a.requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No images left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourceRandom
			asset.requestConfig = a.requestConfig
			asset.ctx = a.ctx

			if !asset.isValidAsset(requestID, deviceID, ImageOnlyAssetTypes, a.RatioWanted) {
				continue
			}

			if a.requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				immichAssetsToCache := slices.Delete(immichAssets, immichAssetIndex, immichAssetIndex+1)
				jsonBytes, cacheMarshalErr := json.Marshal(immichAssetsToCache)
				if cacheMarshalErr != nil {
					log.Error("Failed to marshal immichAssetsToCache", "error", cacheMarshalErr)
					return cacheMarshalErr
				}

				// replace with cache minus used image
				cacheErr := cache.Replace(apiCacheKey, jsonBytes)
				if cacheErr != nil {
					log.Debug("cache not found!")
				}
			}

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}
	return errors.New("no images found for random. Max retries reached")
}
