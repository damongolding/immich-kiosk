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

// favouriteAssetsCount retrieves the total count of favorite assets from the Immich server.
func (a *Asset) favouriteAssetsCount(requestID, deviceID string) (int, error) {

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(0, err, nil, "")
		return 0, err
	}

	requestBody := SearchRandomBody{
		Type:       string(ImageType),
		IsFavorite: true,
		WithPeople: false,
		WithExif:   false,
		Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
	}

	// Include videos if show videos is enabled
	if a.requestConfig.ShowVideos {
		requestBody.Type = ""
	}

	if a.requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	DateFilter(&requestBody, a.requestConfig.DateFilter)

	return a.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)
}

// RandomAssetFromFavourites retrieves a random favorite asset from the Immich server.
// It makes an API request to get random favorite assets and caches them for future use.
// The function includes retries if No viable assets are found and handles caching of
// unused assets for subsequent requests. It filters assets based on type, trash status,
// archive status and aspect ratio requirements. The response assets are processed
// sequentially until a valid asset is found that meets all criteria.
//
// A retry mechanism is implemented to handle cases where No viable assets are found
// in the current cache. The cache is cleared and a new request is made up to MaxRetries
// times. assets are filtered based on:
// - Must not be trashed
// - Must meet archive status requirements (based on ShowArchived config)
// - Must pass ratio check requirements
//
// If caching is enabled, the selected asset is removed from the cache and remaining
// assets are stored for future requests to minimize API calls.
//
// Parameters:
//   - requestID: Unique identifier for tracking and logging the request
//   - deviceID: ID of the device making the request, used for cache segregation
//   - isPrefetch: Boolean indicating if this is a prefetch request for optimization
//
// Returns:
//   - error: Any error encountered during the operation, including API failures,
//     marshaling errors, cache operations, or when max retries are reached with No viable assets found
func (a *Asset) RandomAssetFromFavourites(requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random favourite asset", true)
	} else {
		log.Debug(requestID + " Getting Random favourite asset")
	}

	for range MaxRetries {

		var assets []Asset

		u, err := url.Parse(a.requestConfig.ImmichURL)
		if err != nil {
			return fmt.Errorf("parsing url: %w", err)
		}

		requestBody := SearchRandomBody{
			Type:       string(ImageType),
			IsFavorite: true,
			WithExif:   true,
			WithPeople: true,
			Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
		}

		// Include videos if show videos is enabled
		if a.requestConfig.ShowVideos {
			requestBody.Type = ""
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
			return fmt.Errorf("marshaling request body: %w", err)
		}

		immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, assets)
		apiBody, _, err := immichAPICall(a.ctx, http.MethodPost, apiURL.String(), jsonBody)
		if err != nil {
			_, _, err = immichAPIFail(assets, err, apiBody, apiURL.String())
			return err
		}

		err = json.Unmarshal(apiBody, &assets)
		if err != nil {
			_, _, err = immichAPIFail(assets, err, apiBody, apiURL.String())
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL.String(), deviceID, a.requestConfig.SelectedUser)

		if len(assets) == 0 {
			log.Debug(requestID + " No assets left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		wantedAssetType := ImageOnlyAssetTypes
		if a.requestConfig.ShowVideos {
			wantedAssetType = AllAssetTypes
		}

		for assetIndex, asset := range assets {

			asset.Bucket = kiosk.SourceAlbum
			asset.requestConfig = a.requestConfig
			asset.ctx = a.ctx

			if !asset.isValidAsset(requestID, deviceID, wantedAssetType, a.RatioWanted) {
				continue
			}

			if a.requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				assetsToCache := slices.Delete(assets, assetIndex, assetIndex+1)
				jsonBytes, marshalErr := json.Marshal(assetsToCache)
				if marshalErr != nil {
					log.Error("Failed to marshal assetsToCache", "error", marshalErr)
					return marshalErr
				}

				// replace cache minus used image
				cache.Set(apiCacheKey, jsonBytes, a.requestConfig.Duration)
			}

			asset.BucketID = kiosk.AlbumKeywordFavourites

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable assets left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return errors.New("no assets found for favourites. Max retries reached")
}

func (a *Asset) FavouriteStatus(deviceID string, favourite bool) error {

	body := UpdateAssetBody{
		IsFavorite: favourite,
		IsArchived: a.IsArchived,
	}

	return a.updateAsset(deviceID, body)
}
