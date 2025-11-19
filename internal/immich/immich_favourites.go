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
func (a *Asset) favouriteImagesCount(requestID, deviceID string) (int, error) {

	var allFavouritesCount int

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(allFavouritesCount, err, nil, "")
		return allFavouritesCount, err
	}

	requestBody := SearchRandomBody{
		Type:       string(ImageType),
		IsFavorite: true,
		WithPeople: false,
		WithExif:   false,
		Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
		WithVideo:  a.requestConfig.ShowVideos,
	}

	if a.requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	DateFilter(&requestBody, a.requestConfig.DateFilter)

	allImagesCount, imagesErr := a.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)
	if imagesErr != nil {
		return allFavouritesCount, imagesErr
	}

	allFavouritesCount += allImagesCount

	if a.requestConfig.ShowVideos {
		requestBody.Type = string(VideoType)
		allVideosCount, videosErr := a.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)
		if videosErr != nil {
			return allFavouritesCount, videosErr
		}
		allFavouritesCount += allVideosCount
	}

	return allFavouritesCount, err
}

// RandomImageFromFavourites retrieves a random favorite image from the Immich server.
// It makes an API request to get random favorite images and caches them for future use.
// The function includes retries if No viable assets are found and handles caching of
// unused images for subsequent requests. It filters images based on type, trash status,
// archive status and aspect ratio requirements. The response images are processed
// sequentially until a valid image is found that meets all criteria.
//
// A retry mechanism is implemented to handle cases where No viable assets are found
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
//     marshaling errors, cache operations, or when max retries are reached with No viable assets found
func (a *Asset) RandomImageFromFavourites(requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random favourite image", true)
	} else {
		log.Debug(requestID + " Getting Random favourite image")
	}

	for range MaxRetries {

		var immichAssets []Asset

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
			WithVideo:  a.requestConfig.ShowVideos,
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

		// Add videos if user wants them
		if a.requestConfig.ShowVideos {
			err = a.AddVideos(requestID, deviceID, &immichAssets, apiURL, requestBody)
			if err != nil {
				_, _, err = immichAPIFail(immichAssets, err, nil, apiURL.String())
				return err
			}
		}

		apiCacheKey := cache.APICacheKey(apiURL.String(), deviceID, a.requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No assets left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		wantedAssetType := ImageOnlyAssetTypes
		if a.requestConfig.ShowVideos {
			wantedAssetType = AllAssetTypes
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourceAlbum
			asset.requestConfig = a.requestConfig
			asset.ctx = a.ctx

			if !asset.isValidAsset(requestID, deviceID, wantedAssetType, a.RatioWanted) {
				continue
			}

			if a.requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				immichAssetsToCache := slices.Delete(immichAssets, immichAssetIndex, immichAssetIndex+1)
				jsonBytes, marshalErr := json.Marshal(immichAssetsToCache)
				if marshalErr != nil {
					log.Error("Failed to marshal immichAssetsToCache", "error", marshalErr)
					return marshalErr
				}

				// replace cache minus used image
				cacheErr := cache.Replace(apiCacheKey, jsonBytes)
				if cacheErr != nil {
					log.Debug("cache not found!")
				}
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
