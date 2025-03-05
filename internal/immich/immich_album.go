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

	"github.com/charmbracelet/log"

	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

// AlbumsThatContainAsset finds all albums that contain this asset and updates
// the AppearsIn field with the list of albums. The asset must have its ID set.
// Parameters:
//   - requestID: ID used for tracking API call chain
//   - deviceID: ID of device making the request
//
// Results:
//   - AppearsIn field of the ImmichAsset is updated with list of albums
//   - Any error during API call is logged but function does not return an error
func (i *Asset) AlbumsThatContainAsset(requestID, deviceID string) {

	var albumsContainingAsset Albums

	albums, _, err := i.albums(requestID, deviceID, false, i.ID)
	if err != nil {
		log.Error("Failed to get albums containing asset", "err", err)
		return
	}

	albumsContainingAsset = append(albumsContainingAsset, albums...)

	i.AppearsIn = albumsContainingAsset
}

// albums retrieves albums from Immich based on the shared parameter.
// It constructs the API URL, makes the API call, and returns the albums.
func (i *Asset) albums(requestID, deviceID string, shared bool, contains string) (Albums, string, error) {
	var albums Albums

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(albums, err, nil, "")
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "albums"),
	}

	queryParams := url.Values{}

	if shared {
		queryParams.Set("shared", "true")
	}

	if contains != "" {
		queryParams.Set("assetId", contains)
	}

	apiURL.RawQuery = queryParams.Encode()

	immichAPICall := withImmichAPICache(i.immichAPICall, requestID, deviceID, i.requestConfig, albums)
	body, err := immichAPICall(http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return immichAPIFail(albums, err, body, apiURL.String())
	}

	err = json.Unmarshal(body, &albums)
	if err != nil {
		return immichAPIFail(albums, err, body, apiURL.String())
	}

	return albums, apiURL.String(), nil
}

// allSharedAlbums retrieves all shared albums from Immich.
func (i *Asset) allSharedAlbums(requestID, deviceID string) (Albums, string, error) {
	return i.albums(requestID, deviceID, true, "")
}

// allAlbums retrieves all non-shared albums from Immich.
func (i *Asset) allAlbums(requestID, deviceID string) (Albums, string, error) {
	return i.albums(requestID, deviceID, false, "")
}

// albumAssets retrieves details and assets for a specific album from Immich.
// Parameters:
//   - albumID: The ID of the album to fetch
//   - requestID: ID used for tracking API call
//   - deviceID: ID of the device making the request
//
// Returns:
//   - ImmichAlbum: The album details and associated assets
//   - string: The API URL that was called
//   - error: Any error encountered during the request
func (i *Asset) albumAssets(albumID, requestID, deviceID string) (Album, string, error) {
	var album Album

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(album, err, nil, "")
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "albums", albumID),
	}

	immichAPICall := withImmichAPICache(i.immichAPICall, requestID, deviceID, i.requestConfig, album)
	body, err := immichAPICall(http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return immichAPIFail(album, err, body, apiURL.String())
	}

	err = json.Unmarshal(body, &album)
	if err != nil {
		return immichAPIFail(album, err, body, apiURL.String())
	}

	return album, apiURL.String(), nil
}

// countAssetsInAlbums calculates the total number of assets across multiple albums.
// Parameters:
//   - albums: Slice of ImmichAlbums to count assets from
//
// Returns:
//   - int: Total number of assets across all provided albums
func countAssetsInAlbums(albums Albums) int {
	total := 0
	for _, album := range albums {
		total += album.AssetCount
	}
	return total
}

// AlbumImageCount retrieves the number of images in a specific album from Immich.
// Parameters:
//   - albumID: ID of the album to count images from, can be a special keyword
//   - requestID: ID used for tracking API call
//   - deviceID: ID of the device making the request
//
// Returns:
//   - int: Number of images in the album
//   - error: Any error encountered during the request
func (i *Asset) AlbumImageCount(albumID string, requestID, deviceID string) (int, error) {
	switch albumID {
	case kiosk.AlbumKeywordAll:
		albums, _, err := i.allAlbums(requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get all albums: %w", err)
		}
		return countAssetsInAlbums(albums), nil

	case kiosk.AlbumKeywordShared:
		albums, _, err := i.allSharedAlbums(requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get shared albums: %w", err)
		}
		return countAssetsInAlbums(albums), nil

	case kiosk.AlbumKeywordFavourites, kiosk.AlbumKeywordFavorites:
		favouriteImagesCount, err := i.favouriteImagesCount(requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get favorite images: %w", err)
		}
		return favouriteImagesCount, nil

	default:
		album, _, err := i.albumAssets(albumID, requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get album assets for album %s: %w", albumID, err)
		}
		return album.AssetCount, nil
	}
}

// AssetFromAlbum retrieves and returns an asset from an album in the Immich server.
// It handles retrying failed requests, caching of album assets, and filtering of assets based on type and status.
// The returned asset is set into the ImmichAsset receiver.
//
// Parameters:
//   - albumID: The ID of the album to get an asset from
//   - albumAssetsOrder: The order to return assets (Rand for random, Asc for ascending)
//   - requestID: ID used to track the API request chain
//   - deviceID: ID of the device making the request
//   - isPrefetch: Whether this is a prefetch request for caching
//
// Returns:
//   - error: Any error encountered during the asset retrieval process, including when no viable images are found
//     after maximum retry attempts
func (i *Asset) AssetFromAlbum(albumID string, albumAssetsOrder AssetOrder, requestID, deviceID string) error {

	for range MaxRetries {

		album, apiURL, err := i.albumAssets(albumID, requestID, deviceID)
		if err != nil {
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL, deviceID, i.requestConfig.SelectedUser)

		if len(album.Assets) == 0 {
			log.Debug(requestID+" No images left in cache. Refreshing and trying again for album", albumID)
			cache.Delete(apiCacheKey)

			album, _, retryErr := i.albumAssets(albumID, requestID, deviceID)
			if retryErr != nil || len(album.Assets) == 0 {
				return fmt.Errorf("no assets found for album %s after refresh", albumID)
			}

			continue
		}

		switch albumAssetsOrder {
		case Rand:
			rand.Shuffle(len(album.Assets), func(i, j int) {
				album.Assets[i], album.Assets[j] = album.Assets[j], album.Assets[i]
			})
		case Asc:
			if !album.AssetsOrdered {
				slices.Reverse(album.Assets)
				album.AssetsOrdered = true
			}
		case Desc:
		}

		allowedTypes := ImageOnlyAssetTypes

		if i.requestConfig.ExperimentalAlbumVideo {
			allowedTypes = AllAssetTypes
		}

		for assetIndex, asset := range album.Assets {

			asset.Bucket = kiosk.SourceAlbum
			asset.requestConfig = i.requestConfig

			if !asset.isValidAsset(requestID, deviceID, allowedTypes, i.RatioWanted) {
				continue
			}

			if i.requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				assetsToCache := album
				assetsToCache.Assets = slices.Delete(album.Assets, assetIndex, assetIndex+1)
				jsonBytes, err := json.Marshal(assetsToCache)
				if err != nil {
					log.Error("Failed to marshal assetsToCache", "error", err)
					return err
				}

				// replace with cache minus used asset
				err = cache.Replace(apiCacheKey, jsonBytes)
				if err != nil {
					log.Debug("Failed to update device cache for album", "albumID", albumID, "deviceID", deviceID)
				}

			}

			asset.BucketID = album.ID

			*i = asset

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("no images found for '%s'. Max retries reached", albumID)
}

// selectRandomAlbum selects a random album from the given list of albums, excluding specific albums.
// It weights the selection based on the asset count of each album.
// Returns the selected album ID or an error if no albums are available after exclusions.
// Parameters:
//   - albums: List of albums to select from
//   - excludedAlbums: List of album IDs to exclude from selection
func (i *Asset) selectRandomAlbum(albums Albums, excludedAlbums []string) (string, error) {
	albums.RemoveExcludedAlbums(excludedAlbums)
	if len(albums) == 0 {
		return "", errors.New("no albums available after applying exclusions")
	}

	albumsWithWeighting := []utils.AssetWithWeighting{}
	for _, album := range albums {
		albumsWithWeighting = append(albumsWithWeighting, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourceAlbum, ID: album.ID},
			Weight: album.AssetCount,
		})
	}

	pickedAlbum := utils.PickRandomImageType(i.requestConfig.Kiosk.AssetWeighting, albumsWithWeighting)
	return pickedAlbum.ID, nil
}

// RandomAlbumFromSharedAlbums returns a random album ID from shared albums.
// It takes a requestID for API call tracking and a slice of excluded album IDs.
// The selection is weighted based on the number of assets in each album.
// Returns an error if there are no available albums after exclusions or if the API call fails.
func (i *Asset) RandomAlbumFromSharedAlbums(requestID, deviceID string, excludedAlbums []string) (string, error) {
	albums, _, err := i.allSharedAlbums(requestID, deviceID)
	if err != nil {
		return "", err
	}

	return i.selectRandomAlbum(albums, excludedAlbums)
}

// RandomAlbumFromAllAlbums returns a random album ID from all albums.
// It takes a requestID for API call tracking and a slice of excluded album IDs.
// The selection is weighted based on the number of assets in each album.
// Returns an error if there are no available albums after exclusions or if the API call fails.
func (i *Asset) RandomAlbumFromAllAlbums(requestID, deviceID string, excludedAlbums []string) (string, error) {
	albums, _, err := i.allAlbums(requestID, deviceID)
	if err != nil {
		return "", err
	}

	return i.selectRandomAlbum(albums, excludedAlbums)
}

// RemoveExcludedAlbums filters out albums whose IDs are in the exclude slice.
func (a *Albums) RemoveExcludedAlbums(exclude []string) {
	if len(exclude) == 0 {
		return
	}

	// Create lookup map for O(1) performance
	excludeMap := make(map[string]struct{}, len(exclude))
	for _, id := range exclude {
		excludeMap[id] = struct{}{}
	}

	albums := *a
	withRemoved := slices.DeleteFunc(albums, func(album Album) bool {
		_, excluded := excludeMap[album.ID]
		return excluded
	})

	*a = withRemoved
}
