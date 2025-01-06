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

// albums retrieves albums from Immich based on the shared parameter.
// It constructs the API URL, makes the API call, and returns the albums.
func (i *ImmichAsset) albums(requestID, deviceID string, shared bool) (ImmichAlbums, string, error) {
	var albums ImmichAlbums

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/albums",
	}

	if shared {
		apiUrl.RawQuery = "shared=true"
	}

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
	return i.albums(requestID, deviceID, true)
}

// allAlbums retrieves all non-shared albums from Immich.
func (i *ImmichAsset) allAlbums(requestID, deviceID string) (ImmichAlbums, string, error) {
	return i.albums(requestID, deviceID, false)
}

// albumAssets retrieves all assets associated with a specific album from Immich.
func (i *ImmichAsset) albumAssets(albumID, requestID, deviceID string) (ImmichAlbum, string, error) {
	var album ImmichAlbum

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
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

func (i *ImmichAsset) countAssetsInAlbums(albums ImmichAlbums) int {
	total := 0
	for _, album := range albums {
		total += album.AssetCount
	}
	return total
}

// AlbumImageCount retrieves the number of images in a specific album from Immich.
func (i *ImmichAsset) AlbumImageCount(albumID string, requestID, deviceID string) (int, error) {
	switch albumID {
	case kiosk.AlbumKeywordAll:
		albums, _, err := i.allAlbums(requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get all albums: %w", err)
		}
		return i.countAssetsInAlbums(albums), nil

	case kiosk.AlbumKeywordShared:
		albums, _, err := i.allSharedAlbums(requestID, deviceID)
		if err != nil {
			return 0, fmt.Errorf("failed to get shared albums: %w", err)
		}
		return i.countAssetsInAlbums(albums), nil

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
		return len(album.Assets), nil
	}
}

// RandomImageFromAlbum retrieve random image within a specified album from Immich
func (i *ImmichAsset) RandomImageFromAlbum(albumID, requestID, deviceID string, isPrefetch bool) error {

	album, apiUrl, err := i.albumAssets(albumID, requestID, deviceID)
	if err != nil {
		return err
	}

	apiCacheKey := cache.ApiCacheKey(apiUrl, deviceID)

	if len(album.Assets) == 0 {
		log.Debug(requestID+" No images left in cache. Refreshing and trying again for album", albumID)
		cache.Delete(apiCacheKey)

		album, _, retryErr := i.albumAssets(albumID, requestID, deviceID)
		if retryErr != nil || len(album.Assets) == 0 {
			return fmt.Errorf("no assets found for album %s after refresh", albumID)
		}

		return i.RandomImageFromAlbum(albumID, requestID, deviceID, isPrefetch)
	}

	rand.Shuffle(len(album.Assets), func(i, j int) {
		album.Assets[i], album.Assets[j] = album.Assets[j], album.Assets[i]
	})

	for assetIndex, asset := range album.Assets {
		// We only want images and that are not trashed or archived (unless wanted by user)
		if asset.Type != ImageType || asset.IsTrashed || (asset.IsArchived && !requestConfig.ShowArchived) || !i.ratioCheck(&asset) {
			continue
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
				log.Debug("cache not found!")
			}

		}

		*i = asset

		i.KioskSourceName = album.AlbumName

		return nil
	}

	log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
	cache.Delete(apiCacheKey)
	return i.RandomImageFromAlbum(albumID, requestID, deviceID, isPrefetch)
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
			Asset:  utils.WeightedAsset{Type: "ALBUM", ID: album.ID},
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
