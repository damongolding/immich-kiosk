package immich

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"
	"path"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

// albums retrieves albums from Immich based on the shared parameter.
// It constructs the API URL, makes the API call, and returns the albums.
func (i *ImmichAsset) albums(requestID string, shared bool) (ImmichAlbums, error) {
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

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, albums)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &albums)
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	return albums, nil
}

// allSharedAlbums retrieves all shared albums from Immich.
func (i *ImmichAsset) allSharedAlbums(requestID string) (ImmichAlbums, error) {
	return i.albums(requestID, true)
}

// allAlbums retrieves all non-shared albums from Immich.
func (i *ImmichAsset) allAlbums(requestID string) (ImmichAlbums, error) {
	return i.albums(requestID, false)
}

// albumAssets retrieves all assets associated with a specific album from Immich.
func (i *ImmichAsset) albumAssets(albumID, requestID string) (ImmichAlbum, error) {
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

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, album)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		return immichApiFail(album, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &album)
	if err != nil {
		return immichApiFail(album, err, body, apiUrl.String())
	}

	return album, nil
}

func (i *ImmichAsset) countAssetsInAlbums(albums ImmichAlbums) int {
	total := 0
	for _, album := range albums {
		total += album.AssetCount
	}
	return total
}

// AlbumImageCount retrieves the number of images in a specific album from Immich.
func (i *ImmichAsset) AlbumImageCount(albumID string, requestID string) (int, error) {
	switch albumID {
	case AlbumKeywordAll:
		albums, err := i.allAlbums(requestID)
		if err != nil {
			return 0, fmt.Errorf("failed to get all albums: %w", err)
		}
		return i.countAssetsInAlbums(albums), nil

	case AlbumKeywordShared:
		albums, err := i.allSharedAlbums(requestID)
		if err != nil {
			return 0, fmt.Errorf("failed to get shared albums: %w", err)
		}
		return i.countAssetsInAlbums(albums), nil

	case AlbumKeywordFavourites, AlbumKeywordFavorites:
		favouriteImagesCount, err := i.favouriteImagesCount(requestID)
		if err != nil {
			return 0, fmt.Errorf("failed to get favorite images: %w", err)
		}
		return favouriteImagesCount, nil

	default:
		album, err := i.albumAssets(albumID, requestID)
		if err != nil {
			return 0, fmt.Errorf("failed to get album assets for album %s: %w", albumID, err)
		}
		return len(album.Assets), nil
	}
}

// RandomImageFromAlbum retrieve random image within a specified album from Immich
func (i *ImmichAsset) RandomImageFromAlbum(albumID, requestID, kioskDeviceID string, isPrefetch bool) error {
	album, err := i.albumAssets(albumID, requestID)
	if err != nil {
		return err
	}

	if len(album.Assets) == 0 {
		log.Error("no images found", "for album", albumID)
		return fmt.Errorf("no images found for album %s", albumID)
	}

	rand.Shuffle(len(album.Assets), func(i, j int) {
		album.Assets[i], album.Assets[j] = album.Assets[j], album.Assets[i]
	})

	for _, pick := range album.Assets {
		// We only want images and that are not trashed or archived (unless wanted by user)
		if pick.Type != ImageType || pick.IsTrashed || (pick.IsArchived && !requestConfig.ShowArchived) || !i.ratioCheck(&pick) {
			continue
		}

		*i = pick
		break
	}

	if i.ID == "" {
		log.Error("no images found", "for album", albumID)
		return fmt.Errorf("no images found for album %s", albumID)
	}

	return nil
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
func (i *ImmichAsset) RandomAlbumFromSharedAlbums(requestID string, excludedAlbums []string) (string, error) {
	albums, err := i.allSharedAlbums(requestID)
	if err != nil {
		return "", err
	}

	return i.selectRandomAlbum(albums, excludedAlbums)
}

// RandomAlbumFromAllAlbums returns a random album ID from all albums.
// It takes a requestID for API call tracking and a slice of excluded album IDs.
// The selection is weighted based on the number of assets in each album.
// Returns an error if there are no available albums after exclusions or if the API call fails.
func (i *ImmichAsset) RandomAlbumFromAllAlbums(requestID string, excludedAlbums []string) (string, error) {
	albums, err := i.allAlbums(requestID)
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
