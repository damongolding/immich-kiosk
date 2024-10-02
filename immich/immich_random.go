package immich

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/charmbracelet/log"
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

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "api/assets/random",
		RawQuery: "count=1000",
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, immichAssets)
	apiBody, err := immichApiCall("GET", apiUrl.String(), nil)
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

// GetRandomImage retrieve a random image from Immich
func (i *ImmichAsset) NewRandomImage(requestID, kioskDeviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", kioskDeviceID, "Getting Random image", true)
	} else {
		log.Debug(requestID + " Getting Random image")
	}

	var immichSearchResult ImmichSearchResult
	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal("parsing url", err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/search/random",
	}

	requestBody := strings.NewReader(`
		{
			"type": "IMAGE",
	  		"withExif": true,
			"page": 1,
			"size": 1000
		}`)

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, immichSearchResult)
	apiBody, err := immichApiCall("POST", apiUrl.String(), requestBody)
	if err != nil {
		_, err = immichApiFail(immichSearchResult, err, apiBody, apiUrl.String())
		return err
	}

	err = json.Unmarshal(apiBody, &immichSearchResult)
	if err != nil {
		_, err = immichApiFail(immichSearchResult, err, apiBody, apiUrl.String())
		return err
	}

	immichAssets = immichSearchResult.Assets.Items

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
