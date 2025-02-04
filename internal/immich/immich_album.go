package immich

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"
	"path"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

// AlbumsThatContainAsset finds all albums that contain this asset and updates
// the AppearsIn field with the album names.
// Parameters:
//   - requestID: ID used for tracking API call chain
//   - deviceID: ID of device making the request
func (i *ImmichAsset) AlbumsThatContainAsset(requestID, deviceID string) {

	albumsContaingAsset := []string{}

	albums, _, err := i.albums(requestID, deviceID, false, i.ID)
	if err != nil {
		log.Error("Failed to get albums containing asset", "err", err)
		return
	}

	for _, album := range albums {
		albumsContaingAsset = append(albumsContaingAsset, album.AlbumName)
	}

	i.AppearsIn = albumsContaingAsset
}

// albums retrieves albums from Immich based on the shared parameter.
// It constructs the API URL, makes the API call, and returns the albums.
func (i *ImmichAsset) albums(requestID, deviceID string, shared bool, contains string) (ImmichAlbums, string, error) {
	var albums ImmichAlbums

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		return immichApiFail(albums, err, nil, "")
	}

	apiUrl := url.URL{
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

	apiUrl.RawQuery = queryParams.Encode()

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, deviceID, albums)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &albums)
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	return albums, apiUrl.String(), nil
}

// allSharedAlbums retrieves all shared albums from Immich.
func (i *ImmichAsset) allSharedAlbums(requestID, deviceID string) (ImmichAlbums, string, error) {
	return i.albums(requestID, deviceID, true, "")
}

// allAlbums retrieves all non-shared albums from Immich.
func (i *ImmichAsset) allAlbums(requestID, deviceID string) (ImmichAlbums, string, error) {
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
func (i *ImmichAsset) albumAssets(albumID, requestID, deviceID string) (ImmichAlbum, string, error) {
	var album ImmichAlbum

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		return immichApiFail(album, err, nil, "")
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "albums", albumID),
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, deviceID, album)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		return immichApiFail(album, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &album)
	if err != nil {
		return immichApiFail(album, err, body, apiUrl.String())
	}

	return album, apiUrl.String(), nil
}

// countAssetsInAlbums calculates the total number of assets across multiple albums.
// Parameters:
//   - albums: Slice of ImmichAlbums to count assets from
//
// Returns:
//   - int: Total number of assets across all provided albums
func countAssetsInAlbums(albums ImmichAlbums) int {
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
func (i *ImmichAsset) AlbumImageCount(albumID string, requestID, deviceID string) (int, error) {
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

// ImageFromAlbum retrieves and returns an image from an album in the Immich server.
// It handles retrying failed requests, caching of album assets, and filtering of images based on type and status.
// The returned image is set into the ImmichAsset receiver.
//
// Parameters:
//   - albumID: The ID of the album to get an image from
//   - albumAssetsOrder: The order to return assets (Rand for random, Asc for ascending)
//   - requestID: ID used to track the API request chain
//   - deviceID: ID of the device making the request
//   - isPrefetch: Whether this is a prefetch request for caching
//
// Returns:
//   - error: Any error encountered during the image retrieval process, including when no viable images are found
//     after maximum retry attempts
func (i *ImmichAsset) ImageFromAlbum(albumID string, albumAssetsOrder ImmichAssetOrder, requestID, deviceID string, isPrefetch bool) error {

	for retries := 0; retries < MaxRetries; retries++ {

		album, apiUrl, err := i.albumAssets(albumID, requestID, deviceID)
		if err != nil {
			return err
		}

		apiCacheKey := cache.ApiCacheKey(apiUrl, deviceID, requestConfig.SelectedUser)

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
		}

		allowedTypes := ImageOnlyAssetTypes

		if requestConfig.ExperimentalAlbumVideo {
			allowedTypes = AllAssetTypes
		}

		for assetIndex, asset := range album.Assets {

			isInvalidType := !slices.Contains(allowedTypes, asset.Type)
			isTrashed := asset.IsTrashed
			isArchived := asset.IsArchived && !requestConfig.ShowArchived
			isInvalidRatio := !i.ratioCheck(&asset)

			if isInvalidType || isTrashed || isArchived || isInvalidRatio {
				continue
			}

			if requestConfig.ShowPersonName {
				asset.CheckForFaces(requestID, deviceID)
			}

			if requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				assetsToCache := album
				assetsToCache.Assets = append(album.Assets[:assetIndex], album.Assets[assetIndex+1:]...)
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

			*i = asset

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("No images found for '%s'. Max retries reached.", albumID)
}

// selectRandomAlbum selects a random album from the given list of albums, excluding specific albums.
// It weights the selection based on the asset count of each album.
// Returns the selected album ID or an error if no albums are available after exclusions.
// Parameters:
//   - albums: List of albums to select from
//   - excludedAlbums: List of album IDs to exclude from selection
func (i *ImmichAsset) selectRandomAlbum(albums ImmichAlbums, excludedAlbums []string) (string, error) {
	albums.RemoveExcludedAlbums(excludedAlbums)
	if len(albums) == 0 {
		return "", fmt.Errorf("no albums available after applying exclusions")
	}

	albumsWithWeighting := []utils.AssetWithWeighting{}
	for _, album := range albums {
		albumsWithWeighting = append(albumsWithWeighting, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: kiosk.SourceAlbums, ID: album.ID},
			Weight: album.AssetCount,
		})
	}

	pickedAlbum := utils.PickRandomImageType(requestConfig.Kiosk.AssetWeighting, albumsWithWeighting)
	return pickedAlbum.ID, nil
}

// RandomAlbumFromSharedAlbums returns a random album ID from shared albums.
// It takes a requestID for API call tracking and a slice of excluded album IDs.
// The selection is weighted based on the number of assets in each album.
// Returns an error if there are no available albums after exclusions or if the API call fails.
func (i *ImmichAsset) RandomAlbumFromSharedAlbums(requestID, deviceID string, excludedAlbums []string) (string, error) {
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
func (i *ImmichAsset) RandomAlbumFromAllAlbums(requestID, deviceID string, excludedAlbums []string) (string, error) {
	albums, _, err := i.allAlbums(requestID, deviceID)
	if err != nil {
		return "", err
	}

	return i.selectRandomAlbum(albums, excludedAlbums)
}

// RemoveExcludedAlbums filters out albums whose IDs are in the exclude slice.
func (a *ImmichAlbums) RemoveExcludedAlbums(exclude []string) {
	if len(exclude) == 0 {
		return
	}

	// Create lookup map for O(1) performance
	excludeMap := make(map[string]struct{}, len(exclude))
	for _, id := range exclude {
		excludeMap[id] = struct{}{}
	}

	albums := *a
	withRemoved := slices.DeleteFunc(albums, func(album ImmichAlbum) bool {
		_, excluded := excludeMap[album.ID]
		return excluded
	})

	*a = withRemoved
}
