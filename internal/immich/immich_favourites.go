package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/google/go-querystring/query"
)

// favouriteImagesCount retrieves the total count of favorite images from the Immich server.
func (i *ImmichAsset) favouriteImagesCount(requestID, deviceID string) (int, error) {

	var allFavouritesCount int
	pageCount := 1

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal("parsing url", err)
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
			log.Fatal("marshaling request body", err)
		}

		immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, deviceID, favourites)
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
func (i *ImmichAsset) RandomImageFromFavourites(requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random favourite image", true)
	} else {
		log.Debug(requestID + " Getting Random favourite image")
	}

	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal("parsing url", err)
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
		log.Fatal("marshaling request body", err)
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, deviceID, immichAssets)
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

	apiCacheKey := cache.ApiCacheKey(apiUrl.String(), deviceID)

	if len(immichAssets) == 0 {
		log.Debug(requestID + " No images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
		return i.RandomImageFromFavourites(requestID, deviceID, isPrefetch)
	}

	for immichAssetIndex, img := range immichAssets {
		// We only want images and that are not trashed or archived (unless wanted by user)
		if img.Type != ImageType || img.IsTrashed || (img.IsArchived && !requestConfig.ShowArchived) || !i.ratioCheck(&img) {
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

			// replace cwith cache minus used image
			err = cache.Replace(apiCacheKey, jsonBytes)
			if err != nil {
				log.Debug("cache not found!")
			}
		}

		*i = img
		return nil
	}

	log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
	cache.Delete(apiCacheKey)
	return i.RandomImage(requestID, deviceID, isPrefetch)
}
