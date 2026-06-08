package immich

import (
	"encoding/json"
	"errors"
	"net/url"
	"slices"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
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

	FilterDate(&requestBody, a.requestConfig.FilterDate)

	return a.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)
}

// albumSearchAssetsCount counts assets in an album using the search API.
func (a *Asset) albumSearchAssetsCount(albumID, requestID, deviceID string) (int, error) {
	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(0, err, nil, "")
		return 0, err
	}

	requestBody := a.albumSearchBody(albumID)
	requestBody.WithPeople = false
	requestBody.WithExif = false

	return a.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)
}

// albumSearchBody builds a reusable search request scoped to a single album.
func (a *Asset) albumSearchBody(albumID string) SearchRandomBody {
	requestBody := SearchRandomBody{
		AlbumIDs:   []string{albumID},
		Type:       string(ImageType),
		WithExif:   true,
		WithPeople: true,
		Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
	}

	if a.requestConfig.ShowVideos {
		requestBody.Type = ""
	}

	if a.requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	return requestBody
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

		assets, apiURL, err := a.fetchAssets(requestID, deviceID, requestBody)
		if err != nil {
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
				cache.Set(apiCacheKey, jsonBytes, a.requestConfig.Duration, a.requestConfig.CacheDuration)
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

// AssetFromAlbumSearch retrieves an album asset using search-backed filters and ordering.
func (a *Asset) AssetFromAlbumSearch(albumID string, albumAssetsOrder AssetOrder, requestID, deviceID string) error {
	for range MaxRetries {
		requestBody := a.albumSearchBody(albumID)

		switch albumAssetsOrder {
		case Asc:
			requestBody.Order = string(Asc)
		case Desc:
			requestBody.Order = string(Desc)
		case Rand:
		}

		assets, apiURL, err := a.fetchAssets(requestID, deviceID, requestBody)
		if err != nil {
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL.String(), deviceID, a.requestConfig.SelectedUser)

		if len(assets) == 0 {
			log.Debug(requestID + " No filtered assets left in album cache. Refreshing and trying again")
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
				assetsToCache := slices.Delete(assets, assetIndex, assetIndex+1)
				jsonBytes, marshalErr := json.Marshal(assetsToCache)
				if marshalErr != nil {
					log.Error("Failed to marshal assetsToCache", "error", marshalErr)
					return marshalErr
				}

				cache.Set(apiCacheKey, jsonBytes, a.requestConfig.Duration, a.requestConfig.CacheDuration)
			}

			asset.BucketID = albumID
			if asset.requestConfig.SelectedUser != "" {
				asset.BucketID = albumID + "@" + asset.requestConfig.SelectedUser
			}

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable filtered assets left in album cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return errors.New("no filtered assets found for album. Max retries reached")
}

func (a *Asset) FavouriteStatus(deviceID string, favourite bool) error {
	body := UpdateAssetBody{
		IsFavorite: favourite,
		IsArchived: a.IsArchived,
	}

	return a.updateAsset(deviceID, body)
}
