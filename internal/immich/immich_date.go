package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/google/go-querystring/query"
)

// RandomImageInDateRange retrieves a random image from the Immich API within the specified date range.
// Parameters:
//   - dateRange: A string in the format "YYYY-MM-DD_to_YYYY-MM-DD" or using "today" for current date
//   - requestID: Unique identifier for tracking the request
//   - deviceID: ID of the requesting device
//   - isPrefetch: Whether this is a prefetch request
//
// The function handles:
// - Date range parsing and validation
// - Making API requests with retries
// - Caching of results
// - Filtering images based on type/status
// - Ratio checking of images
//
// Returns an error if no valid images are found after max retries
func (i *ImmichAsset) RandomImageInDateRange(dateRange, requestID, deviceID string, isPrefetch bool) error {

	dateStart, dateEnd, err := determineDateRange(dateRange)
	if err != nil {
		return err
	}

	dateStartHuman := dateStart.Format("2006-01-02")
	dateEndHuman := dateEnd.Format("2006-01-02")

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random image from", dateStartHuman, "to", dateEndHuman)
	} else {
		log.Debug(requestID+" Getting Random image", "from", dateStartHuman, "to", dateEndHuman)
	}

	for retries := 0; retries < MaxRetries; retries++ {

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

		apiCacheKey := cache.ApiCacheKey(apiUrl.String(), deviceID, requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No images left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		for immichAssetIndex, img := range immichAssets {

			// We only want images and that are not trashed or archived (unless wanted by user)
			isInvalidType := img.Type != ImageType
			isTrashed := img.IsTrashed
			isArchived := img.IsArchived && !requestConfig.ShowArchived
			isInvalidRatio := !i.ratioCheck(&img)

			if isInvalidType || isTrashed || isArchived || isInvalidRatio {
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

			*i = img

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("No images found for '%s'. Max retries reached.", dateRange)
}

func determineDateRange(dateRange string) (time.Time, time.Time, error) {
	var dateStart time.Time
	var dateEnd time.Time
	var err error

	switch {
	case strings.Contains(dateRange, "_to_"):
		dateStart, dateEnd, err = processDateRange(dateRange)
		if err != nil {
			return dateStart, dateEnd, err
		}
	case strings.Contains(dateRange, "last-"):
		dateStart, dateEnd, err = processLastDays(dateRange)
		if err != nil {
			return dateStart, dateEnd, err
		}
	}

	return dateStart, dateEnd, err
}

// processDateRange parses a date range string in the format "YYYY-MM-DD_to_YYYY-MM-DD"
// and returns the start and end times. The special value "today" can be used for
// either date. If the end date is before the start date, they will be swapped.
// The end date is adjusted to be the last nanosecond of that day.
// Returns an error if the date range format is invalid or dates cannot be parsed.
func processDateRange(dateRange string) (time.Time, time.Time, error) {

	dateStart := time.Now()
	dateEnd := time.Now()

	dates := strings.SplitN(dateRange, "_to_", 2)
	if len(dates) != 2 {
		return dateStart, dateEnd, fmt.Errorf("Invalid date range format. Expected 'YYYY-MM-DD_to_YYYY-MM-DD', got '%s'", dateRange)
	}

	var err error

	if !strings.EqualFold(dates[0], "today") {
		dateStart, err = time.Parse("2006-01-02", dates[0])
		if err != nil {
			return dateStart, dateEnd, err
		}
	}

	if !strings.EqualFold(dates[1], "today") {
		dateEnd, err = time.Parse("2006-01-02", dates[1])
		if err != nil {
			return dateStart, dateEnd, err
		}
	}

	if dateEnd.Before(dateStart) {
		dateStart, dateEnd = dateEnd, dateStart
	}

	dateEnd = dateEnd.AddDate(0, 0, 1).Add(-time.Nanosecond)

	return dateStart, dateEnd, nil
}

// extractDays extracts a number from a string using regex.
// Returns the first number found in the string or an error if no number is found.
func extractDays(s string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(s)
	if match == "" {
		return 0, fmt.Errorf("no number found")
	}
	return strconv.Atoi(match)
}

// processLastDays takes a date range string in the format "last_X" where X is a number of days
// and returns a time range from X days ago to now.
// Returns an error if the number of days cannot be extracted from the string.
func processLastDays(dateRange string) (time.Time, time.Time, error) {
	dateStart := time.Now()
	dateEnd := time.Now()

	days, err := extractDays(dateRange)
	if err != nil {
		return dateStart, dateEnd, err
	}

	dateStart = dateStart.AddDate(0, 0, -days)

	return dateStart, dateEnd, nil
}
