package immich

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"

	"github.com/charmbracelet/log"
)

// albumAssets retrieves all assets associated with a specific album from Immich.
func (i *ImmichAsset) albumAssets(albumId, requestId string) (ImmichAlbum, error) {
	var album ImmichAlbum

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/albums/" + albumId,
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestId, album)
	body, err := immichApiCall(apiUrl.String())
	if err != nil {
		return immichApiFail(album, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &album)
	if err != nil {
		return immichApiFail(album, err, body, apiUrl.String())
	}

	return album, nil
}

// AlbumImageCount retrieves the number of images in a specific album from Immich.
func (i *ImmichAsset) AlbumImageCount(albumId, requestId string) (int, error) {
	album, err := i.albumAssets(albumId, requestId)
	return len(album.Assets), err
}

// RandomImageFromAlbum retrieve random image within a specified album from Immich
func (i *ImmichAsset) RandomImageFromAlbum(albumId, requestId, kioskDeviceID string, isPrefetch bool) error {
	album, err := i.albumAssets(albumId, requestId)
	if err != nil {
		return err
	}

	if len(album.Assets) == 0 {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	rand.Shuffle(len(album.Assets), func(i, j int) {
		album.Assets[i], album.Assets[j] = album.Assets[j], album.Assets[i]
	})

	for _, pick := range album.Assets {
		// We only want images and that are not trashed or archived (unless wanted by user)
		if pick.Type != "IMAGE" || pick.IsTrashed || (pick.IsArchived && !requestConfig.ShowArchived) || !i.ratioCheck(&pick) {
			continue
		}

		*i = pick
		break
	}

	if i.ID == "" {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	return nil
}
