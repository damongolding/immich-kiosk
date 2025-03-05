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

// favouriteImagesCount retrieves the total count of favorite images from the Immich server.
func (i *Asset) favouriteImagesCount(requestID, deviceID string) (int, error) {

	var allFavouritesCount int

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(allFavouritesCount, err, nil, "")
		return allFavouritesCount, err
	}

	requestBody := SearchRandomBody{
		Type:       string(ImageType),
		IsFavorite: true,
		WithPeople: false,
		WithExif:   false,
		Size:       i.requestConfig.Kiosk.FetchedAssetsSize,
	}

	if i.requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	DateFilter(&requestBody, i.requestConfig.DateFilter)

	allFavouritesCount, err = i.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)

	return allFavouritesCount, err
}

// RandomImageFromFavourites retrieves a random favorite image from the Immich server.
// It makes an API request to get random favorite images and caches them for future use.
// The function includes retries if no viable images are found and handles caching of
// unused images for subsequent requests. It filters images based on type, trash status,
// archive status and aspect ratio requirements. The response images are processed
// sequentially until a valid image is found that meets all criteria.
//
// A retry mechanism is implemented to handle cases where no viable images are found
// in the current cache. The cache is cleared and a new request is made up to MaxRetries
// times. Images are filtered based on:
// - Must be of type ImageType
// - Must not be trashed
// - Must meet archive status requirements (based on ShowArchived config)
// - Must pass ratio check requirements
//
// If caching is enabled, the selected image is removed from the cache and remaining
// images are stored for future requests to minimize API calls.
//
// Parameters:
//   - requestID: Unique identifier for tracking and logging the request
//   - deviceID: ID of the device making the request, used for cache segregation
//   - isPrefetch: Boolean indicating if this is a prefetch request for optimization
//
// Returns:
//   - error: Any error encountered during the operation, including API failures,
//     marshaling errors, cache operations, or when max retries are reached with no viable images found
func (i *Asset) RandomImageFromFavourites(requestID, deviceID string, _ []AssetType, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random favourite image", true)
	} else {
		log.Debug(requestID + " Getting Random favourite image")
	}

	for range MaxRetries {

		var immichAssets []Asset

		u, err := url.Parse(i.requestConfig.ImmichURL)
		if err != nil {
			return fmt.Errorf("parsing url: %w", err)
		}

		requestBody := SearchRandomBody{
			Type:       string(ImageType),
			IsFavorite: true,
			WithExif:   true,
			WithPeople: true,
			Size:       i.requestConfig.Kiosk.FetchedAssetsSize,
		}

		if i.requestConfig.ShowArchived {
			requestBody.WithArchived = true
		}

		DateFilter(&requestBody, i.requestConfig.DateFilter)

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

		immichAPICall := withImmichAPICache(i.immichAPICall, requestID, deviceID, i.requestConfig, immichAssets)
		apiBody, err := immichAPICall(http.MethodPost, apiURL.String(), jsonBody)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, apiBody, apiURL.String())
			return err
		}

		err = json.Unmarshal(apiBody, &immichAssets)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, apiBody, apiURL.String())
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL.String(), deviceID, i.requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No images left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourceAlbum
			asset.requestConfig = i.requestConfig

			if !asset.isValidAsset(requestID, deviceID, ImageOnlyAssetTypes, i.RatioWanted) {
				continue
			}

			if i.requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				immichAssetsToCache := slices.Delete(immichAssets, immichAssetIndex, immichAssetIndex+1)
				jsonBytes, err := json.Marshal(immichAssetsToCache)
				if err != nil {
					log.Error("Failed to marshal immichAssetsToCache", "error", err)
					return err
				}

				// replace cache minus used image
				err = cache.Replace(apiCacheKey, jsonBytes)
				if err != nil {
					log.Debug("cache not found!")
				}
			}

			asset.BucketID = kiosk.AlbumKeywordFavourites

			*i = asset

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return errors.New("no images found for favourites. Max retries reached")
}
