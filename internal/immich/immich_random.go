package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

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
		return err
	}

	requestBody := ImmichSearchRandomBody{
		Type:       string(ImageType),
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

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, immichAssets)
	apiBody, err := immichApiCall("POST", apiUrl.String(), jsonBody)
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

func (i *ImmichAsset) RandomImageInDateRange(dateRange, requestID, kioskDeviceID string, isPrefetch bool) error {
	return i.randomImageInDateRange(dateRange, requestID, kioskDeviceID, isPrefetch, 0)
}

func (i *ImmichAsset) randomImageInDateRange(dateRange, requestID, kioskDeviceID string, isPrefetch bool, retires int) error {

	if retires >= MaxRetries {
		return fmt.Errorf("No images found for '%s'. Max retries reached.", dateRange)
	}

	dates := strings.Split(dateRange, "_")
	dateStart, err := time.Parse("2006-01-02", dates[1])
	if err != nil {
		return err
	}
	dateEnd, err := time.Parse("2006-01-02", dates[3])
	if err != nil {
		return err
	}

	if dateEnd.Before(dateStart) {
		dateStart, dateEnd = dateEnd, dateStart
	}

	dateStartHuman := dateStart.Format("2006-01-02")
	dateEndHuman := dateEnd.Format("2006-01-02")

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", kioskDeviceID, "Getting Random image between", dateStartHuman, "and", dateEndHuman)
	} else {
		log.Debug(requestID+" Getting Random image", "between", dateStartHuman, "and", dateEndHuman)
	}

	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal("parsing url", err)
		return err
	}

	requestBody := ImmichSearchRandomBody{
		Type:        string(ImageType),
		TakenAfter:  dateStart.Format(time.RFC3339),
		TakenBefore: dateEnd.Format(time.RFC3339),
		WithExif:    true,
		WithPeople:  true,
		Size:        requestConfig.Kiosk.FetchedAssetsSize,
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

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, immichAssets)
	apiBody, err := immichApiCall("POST", apiUrl.String(), jsonBody)
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
		retires++
		return i.randomImageInDateRange(dateRange, requestID, kioskDeviceID, isPrefetch, retires)
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

		img.KioskSourceName = fmt.Sprintf("Between %s and %s", dateStartHuman, dateEndHuman)

		*i = img

		return nil
	}

	log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
	apiCache.Delete(apiUrl.String())
	retires++
	return i.randomImageInDateRange(dateRange, requestID, kioskDeviceID, isPrefetch, retires)
}
