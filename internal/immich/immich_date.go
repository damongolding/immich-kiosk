package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/google/go-querystring/query"
)

// RandomImageInDateRange gets a random image from the Immich server within the specified date range.
// dateRange should be in the format "YYYY-MM-DD_to_YYYY-MM-DD"
// requestID is used for tracing the request through logs
// deviceID identifies the kiosk device making the request
// isPrefetch indicates if this is a prefetch request
func (i *ImmichAsset) RandomImageInDateRange(dateRange, requestID, deviceID string, isPrefetch bool) error {
	return i.randomImageInDateRange(dateRange, requestID, deviceID, isPrefetch, 0)
}

// randomImageInDateRange is the internal implementation of RandomImageInDateRange that includes retry logic
// It makes requests to the Immich API to get random images within the date range and handles caching
// retries tracks the number of retry attempts made
func (i *ImmichAsset) randomImageInDateRange(dateRange, requestID, deviceID string, isPrefetch bool, retries int) error {

	if retries >= MaxRetries {
		return fmt.Errorf("No images found for '%s'. Max retries reached.", dateRange)
	}

	if !strings.Contains(dateRange, "_to_") {
		return fmt.Errorf("Invalid date range format. Expected 'YYYY-MM-DD_to_YYYY-MM-DD', got '%s'", dateRange)
	}

	dates := strings.SplitN(dateRange, "_to_", 2)
	if len(dates) != 2 {
		return fmt.Errorf("Invalid date range format. Expected 'YYYY-MM-DD_to_YYYY-MM-DD', got '%s'", dateRange)
	}

	dateStart := time.Now()
	dateEnd := time.Now()
	var err error

	if !strings.EqualFold(dates[0], "today") {
		dateStart, err = time.Parse("2006-01-02", dates[0])
		if err != nil {
			return err
		}
	}

	if !strings.EqualFold(dates[1], "today") {
		dateEnd, err = time.Parse("2006-01-02", dates[1])
		if err != nil {
			return err
		}
	}

	if dateEnd.Before(dateStart) {
		dateStart, dateEnd = dateEnd, dateStart
	}

	dateStartHuman := dateStart.Format("2006-01-02")
	dateEndHuman := dateEnd.Format("2006-01-02")

	dateEnd = dateEnd.AddDate(0, 0, 1).Add(-time.Nanosecond)

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random image from", dateStartHuman, "to", dateEndHuman)
	} else {
		log.Debug(requestID+" Getting Random image", "from", dateStartHuman, "to", dateEndHuman)
	}

	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		return fmt.Errorf("parsing url: %w", err)
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
		return fmt.Errorf("marshaling request body: %w", err)
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
		retries++
		return i.randomImageInDateRange(dateRange, requestID, deviceID, isPrefetch, retries)
	}

	log.Info("Date range res", "items", len(immichAssets))

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

			// replace cache with used image(s) removed
			err = cache.Replace(apiCacheKey, jsonBytes)
			if err != nil {
				log.Debug("Failed to update cache", "error", err, "url", apiUrl.String())
			}
		}

		img.KioskSourceName = fmt.Sprintf("%s to %s", dateStartHuman, dateEndHuman)

		*i = img

		return nil
	}

	log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
	cache.Delete(apiCacheKey)
	retries++
	return i.randomImageInDateRange(dateRange, requestID, deviceID, isPrefetch, retries)
}
