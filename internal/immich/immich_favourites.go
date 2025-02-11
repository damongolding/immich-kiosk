package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/google/go-querystring/query"
)

// favouriteImagesCount retrieves the total count of favorite images from the Immich server.
func (i *ImmichAsset) favouriteImagesCount(requestID, deviceID string) (int, error) {

	var allFavouritesCount int
	pageCount := 1

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		_, _, err = immichApiFail(allFavouritesCount, err, nil, "")
		return allFavouritesCount, err
	}

	requestBody := ImmichSearchRandomBody{
		Type:       string(ImageType),
		IsFavorite: true,
		WithPeople: false,
		WithExif:   false,
		Size:       requestConfig.Kiosk.FetchedAssetsSize,
	}

	if requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	for {

		var favourites ImmichSearchMetadataResponse

		requestBody.Page = pageCount

		// convert body to queries so url is unique and can be cached
		queries, _ := query.Values(requestBody)

		apiUrl := url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     "api/search/metadata",
			RawQuery: queries.Encode(),
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			_, _, err = immichApiFail(allFavouritesCount, err, nil, apiUrl.String())
			return allFavouritesCount, err
		}

		immichApiCall := withImmichApiCache(i.immichApiCall, requestID, deviceID, favourites)
		apiBody, err := immichApiCall("POST", apiUrl.String(), jsonBody)
		if err != nil {
			_, _, err = immichApiFail(favourites, err, apiBody, apiUrl.String())
			return allFavouritesCount, err
		}

		err = json.Unmarshal(apiBody, &favourites)
		if err != nil {
			_, _, err = immichApiFail(favourites, err, apiBody, apiUrl.String())
			return allFavouritesCount, err
		}

		allFavouritesCount += favourites.Assets.Total

		if favourites.Assets.NextPage == "" {
			break
		}

		pageCount++
	}

	return allFavouritesCount, nil
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
func (i *ImmichAsset) RandomImageFromFavourites(requestID, deviceID string, allowedAssetType []ImmichAssetType, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random favourite image", true)
	} else {
		log.Debug(requestID + " Getting Random favourite image")
	}

	for retries := 0; retries < MaxRetries; retries++ {

		var immichAssets []ImmichAsset

		u, err := url.Parse(requestConfig.ImmichUrl)
		if err != nil {
			return fmt.Errorf("parsing url: %w", err)
		}

		requestBody := ImmichSearchRandomBody{
			Type:       string(ImageType),
			IsFavorite: true,
			WithExif:   true,
			WithPeople: true,
			Size:       requestConfig.Kiosk.FetchedAssetsSize,
		}

		if requestConfig.ShowArchived {
			requestBody.WithArchived = true
		}

		// convert body to queries so url is unique and can be cached
		queries, _ := query.Values(requestBody)

		apiUrl := url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     "api/search/random",
			RawQuery: fmt.Sprintf("kiosk=%x", sha256.Sum256([]byte(queries.Encode()))),
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}

		immichApiCall := withImmichApiCache(i.immichApiCall, requestID, deviceID, immichAssets)
		apiBody, err := immichApiCall("POST", apiUrl.String(), jsonBody)
		if err != nil {
			_, _, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
			return err
		}

		err = json.Unmarshal(apiBody, &immichAssets)
		if err != nil {
			_, _, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
			return err
		}

		apiCacheKey := cache.ApiCacheKey(apiUrl.String(), deviceID, requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No images left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		for immichAssetIndex, asset := range immichAssets {

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
				// Remove the current image from the slice
				immichAssetsToCache := append(immichAssets[:immichAssetIndex], immichAssets[immichAssetIndex+1:]...)
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

			asset.Bucket = kiosk.SourceAlbums
			asset.BucketID = kiosk.AlbumKeywordFavourites

			*i = asset

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("No images found for favourites. Max retries reached.")
}
