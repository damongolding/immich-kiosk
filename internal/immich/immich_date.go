package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
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

	for range MaxRetries {

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

		immichApiCall := withImmichApiCache(i.immichApiCall, requestID, deviceID, immichAssets)
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

		for immichAssetIndex, asset := range immichAssets {

			if !asset.isValidAsset(ImageOnlyAssetTypes, i.RatioWanted) {
				continue
			}

			err := asset.AssetInfo(requestID, deviceID)
			if err != nil {
				log.Error("Failed to get additional asset data", "error", err)
			}

			if asset.containsTag(kiosk.TagSkip) {
				continue
			}

			if requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				immichAssetsToCache := slices.Delete(immichAssets, immichAssetIndex, immichAssetIndex+1)
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

			asset.Bucket = kiosk.SourceDateRange
			asset.BucketID = dateRange

			*i = asset

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
	case strings.EqualFold(dateRange, "today"):
		dateStart, dateEnd = processTodayDateRange()

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
	default:
		return dateStart, dateEnd, fmt.Errorf("invalid date filter format: %s. Expected format: YYYY-MM-DD_to_YYYY-MM-DD or last-X", dateRange)
	}

	return dateStart, dateEnd, err
}

// processTodayDateRange returns the start and end times for today's date range.
//
// The function:
// - Uses local timezone for time calculations
// - Sets start time to beginning of current day (00:00:00)
// - Sets end time to last nanosecond of current day (23:59:59.999999999)
//
// Returns:
// - Start time: Beginning of current day
// - End time: End of current day
func processTodayDateRange() (time.Time, time.Time) {
	now := time.Now().Local()
	dateStart := now.Truncate(24 * time.Hour)

	dateEnd := dateStart.AddDate(0, 0, 1).Add(-time.Nanosecond)
	return dateStart, dateEnd
}

// processDateRange parses a date range string in the format "YYYY-MM-DD_to_YYYY-MM-DD"
// and returns the start and end times for filtering images.
//
// The function:
// - Accepts a string in format "YYYY-MM-DD_to_YYYY-MM-DD"
// - Supports special value "today" for either date
// - Swaps dates if end date is before start date
// - Adjusts end date to last nanosecond of that day
// - Uses local timezone for parsing dates
// - Returns start time, end time and any error
//
// Example:
//
//	"2023-01-01_to_today" -> Jan 1 2023 00:00:00 to current date 23:59:59.999999999
//	"today_to_2023-12-31" -> current date 00:00:00 to Dec 31 2023 23:59:59.999999999
func processDateRange(dateRange string) (time.Time, time.Time, error) {

	var err error

	now := time.Now().Local()
	dateStart := now.Truncate(24 * time.Hour)
	dateEnd := dateStart

	dates := strings.SplitN(dateRange, "_to_", 2)
	if len(dates) != 2 {
		return dateStart, dateEnd, fmt.Errorf("Invalid date range format. Expected 'YYYY-MM-DD_to_YYYY-MM-DD', got '%s'", dateRange)
	}

	if !strings.EqualFold(dates[0], "today") {
		dateStart, err = time.ParseInLocation("2006-01-02", dates[0], time.Local)
		if err != nil {
			return dateStart, dateEnd, err
		}
	}

	if !strings.EqualFold(dates[1], "today") {
		dateEnd, err = time.ParseInLocation("2006-01-02", dates[1], time.Local)
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

	now := time.Now().Local()
	dateStart := now
	dateEnd := now

	days, err := extractDays(dateRange)
	if err != nil {
		return dateStart, dateEnd, err
	}

	dateStart = dateStart.AddDate(0, 0, -days)

	return dateStart, dateEnd, nil
}
