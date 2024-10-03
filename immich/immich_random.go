package immich

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/google/go-querystring/query"
	"github.com/patrickmn/go-cache"
)

// GetRandomImage retrieve a random image from Immich
func (i *ImmichAsset) RandomImage(requestID, kioskDeviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", kioskDeviceID, "Getting Random image", true)
	} else {
		log.Debug(requestID + " Getting Random image")
	}

	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal("parsing url", err)
	}

	requestBody := ImmichSearchBody{
		Type:     string(ImageType),
		WithExif: true,
		Size:     1000,
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
		RawQuery: queries.Encode(),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal("marshaling request body", err)
	}

	requestBodyReader := bytes.NewReader(jsonBody)

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, immichAssets)
	apiBody, err := immichApiCall("POST", apiUrl.String(), requestBodyReader)
	if err != nil {
		_, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
		return err
	}

	err = json.Unmarshal(apiBody, &immichAssets)
	if err != nil {
		_, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
		return err
	}

	if len(immichAssets) == 0 {
		log.Debug(requestID + " No images left in cache. Refreshing and trying again")
		apiCache.Delete(apiUrl.String())
		return i.RandomImage(requestID, kioskDeviceID, isPrefetch)
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

			log.Info("items in cache", "items", len(immichAssetsToCache))
			// replace cwith cache minus used image
			err = apiCache.Replace(apiUrl.String(), jsonBytes, cache.DefaultExpiration)
			if err != nil {
				log.Debug("cache not found!")
			}
		}

		*i = img
		return nil
	}

	log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
	apiCache.Delete(apiUrl.String())
	return i.RandomImage(requestID, kioskDeviceID, isPrefetch)
}
