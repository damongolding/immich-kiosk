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
func (a *Asset) AlbumsThatContainAsset(requestID, deviceID string) {

	var albumsContainingAsset Albums

	albums, _, err := a.albums(requestID, deviceID, false, a.ID, false)
	if err != nil {
		log.Error("Failed to get albums containing asset", "err", err)
		return
	}

	albumsContainingAsset = append(albumsContainingAsset, albums...)

	a.AppearsIn = albumsContainingAsset
}

// albums retrieves albums from Immich based on the shared parameter.
// It constructs the API URL, makes the API call, and returns the albums.
func (a *Asset) albums(requestID, deviceID string, shared bool, contains string, bypassCache bool) (Albums, string, error) {
	var albums Albums

	u, err := url.Parse(a.requestConfig.ImmichURL)
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

	var body []byte

	if bypassCache {
		body, _, err = a.immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
		if err != nil {
			return immichAPIFail(albums, err, body, apiURL.String())
		}
	} else {
		immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, albums)
		body, _, err = immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
		if err != nil {
			return immichAPIFail(albums, err, body, apiURL.String())
		}
	}

	err = json.Unmarshal(body, &albums)
	if err != nil {
		return immichAPIFail(albums, err, body, apiURL.String())
	}

	return albums, apiURL.String(), nil
}

// allSharedAlbums retrieves all shared albums from Immich.
func (a *Asset) allSharedAlbums(requestID, deviceID string) (Albums, string, error) {
	return a.albums(requestID, deviceID, true, "", false)
}

// allAlbums retrieves all albums (owned and shared) from Immich.
func (a *Asset) allAlbums(requestID, deviceID string) (Albums, string, error) {
	owned, ownedURL, ownedErr := a.albums(requestID, deviceID, false, "", false)
	shared, sharedURL, sharedErr := a.albums(requestID, deviceID, true, "", false)
	all := make(Albums, len(owned)+len(shared))
	copy(all, owned)
	copy(all[len(owned):], shared)

	var err error
	if ownedErr != nil {
		err = errors.Join(err, ownedErr)
	}

	if sharedErr != nil {
		err = errors.Join(err, sharedErr)
	}

	return all, ownedURL + " && " + sharedURL, err
}

func (a *Asset) AllAlbums(requestID, deviceID string) (Albums, error) {
	all, _, err := a.allAlbums(requestID, deviceID)
	return all, err
}

// allOwnedAlbums retrieves all non-shared albums from Immich.
func (a *Asset) allOwnedAlbums(requestID, deviceID string) (Albums, string, error) {
	return a.albums(requestID, deviceID, false, "", false)
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
func (a *Asset) albumAssets(albumID, requestID, deviceID string) (Album, string, error) {
	var album Album

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(album, err, nil, "")
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "albums", albumID),
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, album)
	body, _, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
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
func (a *Asset) AlbumImageCount(albumID string, requestID, deviceID string) (int, error) {
	switch albumID {
	case kiosk.AlbumKeywordAll:
		albums, albumsURL, err := a.allAlbums(requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get all albums (%s) err=%w", albumsURL, err)
		}
		return countAssetsInAlbums(albums), nil

	case kiosk.AlbumKeywordOwned:
		albums, albumsURL, err := a.allOwnedAlbums(requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get owned albums (%s) err=%w", albumsURL, err)

		}
		return countAssetsInAlbums(albums), nil

	case kiosk.AlbumKeywordShared:
		albums, albumsURL, err := a.allSharedAlbums(requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get shared albums (%s) err=%w", albumsURL, err)
		}
		return countAssetsInAlbums(albums), nil

	case kiosk.AlbumKeywordFavourites, kiosk.AlbumKeywordFavorites:
		favouriteAssetCount, err := a.favouriteAssetsCount(requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get favorite assets: %w", err)
		}
		return favouriteAssetCount, nil

	default:
		album, _, err := a.albumAssets(albumID, requestID, deviceID)
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
//   - error: Any error encountered during the asset retrieval process, including when No viable assets are found
//     after maximum retry attempts
func (a *Asset) AssetFromAlbum(albumID string, albumAssetsOrder AssetOrder, requestID, deviceID string) error {

	for range MaxRetries {

		album, apiURL, err := a.albumAssets(albumID, requestID, deviceID)
		if err != nil {
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL, deviceID, a.requestConfig.SelectedUser)

		if len(album.Assets) == 0 {
			log.Debug(requestID+" No assets left in cache. Refreshing and trying again for album", albumID)
			cache.Delete(apiCacheKey)

			al, _, retryErr := a.albumAssets(albumID, requestID, deviceID)
			if retryErr != nil || len(al.Assets) == 0 {
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

		if a.requestConfig.ShowVideos {
			allowedTypes = AllAssetTypes
		}

		for assetIndex, asset := range album.Assets {

			asset.Bucket = kiosk.SourceAlbum
			asset.requestConfig = a.requestConfig
			asset.ctx = a.ctx

			if !asset.isValidAsset(requestID, deviceID, allowedTypes, a.RatioWanted) {
				continue
			}

			if a.requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				assetsToCache := album
				assetsToCache.Assets = slices.Delete(album.Assets, assetIndex, assetIndex+1)
				jsonBytes, marshalErr := json.Marshal(assetsToCache)
				if marshalErr != nil {
					log.Error("Failed to marshal assetsToCache", "error", marshalErr)
					return marshalErr
				}

				// replace with cache minus used asset
				cache.Set(apiCacheKey, jsonBytes, a.requestConfig.Duration)
			}

			asset.BucketID = album.ID

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable assets left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("no assets found for '%s'. Max retries reached", albumID)
}

// selectRandomAlbum selects a random album from the given list of albums, excluding specific albums.
// It weights the selection based on the asset count of each album.
// Returns the selected album ID or an error if no albums are available after exclusions.
// Parameters:
//   - albums: List of albums to select from
//   - excludedAlbums: List of album IDs to exclude from selection
func (a *Asset) selectRandomAlbum(albums Albums, excludedAlbums []string) (string, error) {
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

	pickedAlbum := utils.PickRandomImageType(a.requestConfig.Kiosk.AssetWeighting, albumsWithWeighting)
	return pickedAlbum.ID, nil
}

// RandomAlbumFromSharedAlbums returns a random album ID from shared albums.
// It takes a requestID for API call tracking and a slice of excluded album IDs.
// The selection is weighted based on the number of assets in each album.
// Returns an error if there are no available albums after exclusions or if the API call fails.
func (a *Asset) RandomAlbumFromSharedAlbums(requestID, deviceID string, excludedAlbums []string) (string, error) {
	albums, _, err := a.allSharedAlbums(requestID, deviceID)
	if err != nil {
		return "", err
	}

	return a.selectRandomAlbum(albums, excludedAlbums)
}

// RandomAlbumFromAllAlbums returns a random album ID from all albums.
// It takes a requestID for API call tracking and a slice of excluded album IDs.
// The selection is weighted based on the number of assets in each album.
// Returns an error if there are no available albums after exclusions or if the API call fails.
func (a *Asset) RandomAlbumFromAllAlbums(requestID, deviceID string, excludedAlbums []string) (string, error) {
	albums, _, err := a.allAlbums(requestID, deviceID)
	if err != nil {
		return "", err
	}

	return a.selectRandomAlbum(albums, excludedAlbums)
}

// RandomAlbumFromOwnedAlbums returns a random album ID from owned albums.
// It takes a requestID for API call tracking and a slice of excluded album IDs.
// The selection is weighted based on the number of assets in each album.
// Returns an error if there are no available albums after exclusions or if the API call fails.
func (a *Asset) RandomAlbumFromOwnedAlbums(requestID, deviceID string, excludedAlbums []string) (string, error) {
	albums, _, err := a.allOwnedAlbums(requestID, deviceID)
	if err != nil {
		return "", err
	}

	return a.selectRandomAlbum(albums, excludedAlbums)
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

// kioskLikedAlbum looks for and returns the kiosk "liked" album from all available albums.
// Parameters:
//   - requestID: ID used for tracking API call chain
//   - deviceID: ID of device making the request
//
// Returns:
//   - Album: The found liked album
//   - error: Error if album not found or API call fails
func (a *Asset) kioskLikedAlbum(requestID, deviceID string) (Album, error) {

	var album Album

	albums, _, err := a.albums(requestID, deviceID, false, "", true)
	if err != nil {
		return album, fmt.Errorf("failed to fetch albums: %w", err)
	}

	if len(albums) == 0 {
		return album, errors.New("no albums found")
	}

	for _, album := range albums {
		if album.AlbumName == kiosk.FavoriteAlbumName {
			return album, nil
		}
	}

	return album, errors.New("kiosk liked album not found")
}

// createKioskLikedAlbum creates a new album to store kiosk "liked" assets.
// Parameters:
//   - requestID: ID used for tracking API call chain
//   - deviceID: ID of device making the request
//
// Returns:
//   - string: ID of created album
//   - error: Error if creation fails
func (a *Asset) createKioskLikedAlbum() (string, error) {

	var res Album

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return "", err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "albums"),
	}

	requestBody := AlbumCreateBody{
		AlbumName:   kiosk.FavoriteAlbumName,
		Description: "Album for liked assets from Kiosk",
	}

	jsonBody, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return "", fmt.Errorf("marshaling request body: %w", marshalErr)
	}

	body, _, err := a.immichAPICall(a.ctx, http.MethodPost, apiURL.String(), jsonBody)
	if err != nil {
		_, _, resErr := immichAPIFail(res, err, body, apiURL.String())
		return "", resErr
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", err
	}

	return res.ID, nil
}

// AddToKioskLikedAlbum adds the current asset to the kiosk "liked" album,
// creating the album if it doesn't exist.
// Parameters:
//   - requestID: ID used for tracking API call chain
//   - deviceID: ID of device making the request
//
// Returns:
//   - error: Error if adding asset fails
func (a *Asset) AddToKioskLikedAlbum(requestID, deviceID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	album, err := a.kioskLikedAlbum(requestID, deviceID)
	if err != nil {
		album.ID, err = a.createKioskLikedAlbum()
		if err != nil {
			return fmt.Errorf("failed to create kiosk liked album: %w", err)
		}
		log.Debug(requestID+" Created", "albumName", kiosk.FavoriteAlbumName, "albumID", album.ID)
	}

	return a.addAssetToAlbum(album.ID)
}

func (a *Asset) RemoveFromKioskLikedAlbum(requestID, deviceID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	album, err := a.kioskLikedAlbum(requestID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get kiosk liked album: %w", err)
	}

	if album.ID == "" {
		return nil
	}

	return a.removeAssetFromAlbum(album.ID)
}

// modifyAssetInAlbum performs an API call to modify the current asset's presence in the specified album.
// The method parameter determines whether the asset is added (PUT) or removed (DELETE).
// Parameters:
//   - albumID: ID of album to modify
//   - method: HTTP method to use (PUT to add, DELETE to remove)
//
// Returns:
//   - error: Error if API call or unmarshaling response fails
func (a *Asset) modifyAssetInAlbum(albumID string, method string) error {
	var res AlbumCreateResponse

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "albums", albumID, "assets"),
	}

	requestBody := AddAssetsToAlbumBody{
		IDs: []string{a.ID},
	}

	jsonBody, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return fmt.Errorf("marshaling request body: %w", marshalErr)
	}

	body, _, err := a.immichAPICall(a.ctx, method, apiURL.String(), jsonBody)
	if err != nil {
		_, _, err = immichAPIFail(res, err, body, apiURL.String())
		return err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		_, _, err = immichAPIFail(res, err, body, apiURL.String())
		return err
	}

	return nil
}

// addAssetToAlbum adds the current asset to the specified album via the Immich API.
// Parameters:
//   - albumID: ID of album to add the asset to
//
// Returns:
//   - error: Error if adding the asset fails
func (a *Asset) addAssetToAlbum(albumID string) error {
	return a.modifyAssetInAlbum(albumID, http.MethodPut)
}

// removeAssetFromAlbum removes the current asset from the specified album via the Immich API.
// Parameters:
//   - albumID: ID of album to remove the asset from
//
// Returns:
//   - error: Error if removing the asset fails
func (a *Asset) removeAssetFromAlbum(albumID string) error {
	return a.modifyAssetInAlbum(albumID, http.MethodDelete)
}
