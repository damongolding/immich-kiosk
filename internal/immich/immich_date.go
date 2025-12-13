package immich

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

// RandomAssetInDateRange retrieves a random asset from the Immich API within the specified date range.
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
// - Filtering assets based on type/status
// - Ratio checking of assets
//
// Returns an error if no valid assets are found after max retries
func (a *Asset) RandomAssetInDateRange(dateRange, requestID, deviceID string, isPrefetch bool) error {

	dateStart, dateEnd, dateErr := determineDateRange(dateRange)
	if dateErr != nil {
		return dateErr
	}

	dateStartHuman := dateStart.Format("2006-01-02 15:04:05 MST")
	dateEndHuman := dateEnd.Format("2006-01-02 15:04:05 MST")

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random asset from", dateStartHuman, "to", dateEndHuman)
	} else {
		log.Debug(requestID+" Getting Random asset", "from", dateStartHuman, "to", dateEndHuman)
	}

	for range MaxRetries {

		var immichAssets []Asset

		u, err := url.Parse(a.requestConfig.ImmichURL)
		if err != nil {
			return fmt.Errorf("parsing url: %w", err)
		}

		requestBody := SearchRandomBody{
			Type:        string(ImageType),
			TakenAfter:  dateStart.Format(time.RFC3339),
			TakenBefore: dateEnd.Format(time.RFC3339),
			WithExif:    true,
			WithPeople:  true,
			Size:        a.requestConfig.Kiosk.FetchedAssetsSize,
		}

		// Include videos if show videos is enabled
		if a.requestConfig.ShowVideos {
			requestBody.Type = ""
		}

		if a.requestConfig.ShowArchived {
			requestBody.WithArchived = true
		}

		// convert body to queries so url is unique and can be cached
		queries, _ := query.Values(requestBody)

		apiURL := url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     "api/search/random",
			RawQuery: fmt.Sprintf("kiosk=%x", sha256.Sum256([]byte(queries.Encode()))),
		}

		jsonBody, marshalErr := json.Marshal(requestBody)
		if marshalErr != nil {
			return fmt.Errorf("marshaling request body: %w", marshalErr)
		}

		immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, immichAssets)
		apiBody, _, err := immichAPICall(a.ctx, http.MethodPost, apiURL.String(), jsonBody)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, apiBody, apiURL.String())
			return err
		}

		err = json.Unmarshal(apiBody, &immichAssets)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, apiBody, apiURL.String())
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL.String(), deviceID, a.requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No assets left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		wantedAssetType := ImageOnlyAssetTypes
		if a.requestConfig.ShowVideos {
			wantedAssetType = AllAssetTypes
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourceDateRange
			asset.requestConfig = a.requestConfig
			asset.ctx = a.ctx

			if !asset.isValidAsset(requestID, deviceID, wantedAssetType, a.RatioWanted) {
				continue
			}

			if a.requestConfig.Kiosk.Cache {
				// Remove the current asset from the slice
				immichAssetsToCache := slices.Delete(immichAssets, immichAssetIndex, immichAssetIndex+1)
				jsonBytes, cacheMarshalErr := json.Marshal(immichAssetsToCache)
				if cacheMarshalErr != nil {
					log.Error("Failed to marshal immichAssetsToCache", "error", cacheMarshalErr)
					return cacheMarshalErr
				}

				// replace cache with used asset(s) removed
				cache.Set(apiCacheKey, jsonBytes, a.requestConfig.Duration)
			}

			asset.BucketID = dateRange

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable assets left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("no assets found for '%s'. Max retries reached", dateRange)
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
// This function takes no parameters and returns the start and end times for the
// current day in the local timezone. The start time is set to midnight (00:00:00.000000000)
// and the end time is set to just before midnight (23:59:59.999999999).
//
// The time calculations use the system's local timezone settings.
//
// Returns:
//   - time.Time: Start time (beginning of current day at 00:00:00.000000000)
//   - time.Time: End time (end of current day at 23:59:59.999999999)
func processTodayDateRange() (time.Time, time.Time) {
	now := time.Now().Local()
	dateStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	dateEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.Local)
	return dateStart, dateEnd
}

// processDateRange parses a date range string in the format "YYYY-MM-DD_to_YYYY-MM-DD"
// and returns the start and end times for filtering assets.
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

	dateStart, dateEnd := processTodayDateRange()

	dates := strings.SplitN(dateRange, "_to_", 2)
	if len(dates) != 2 {
		return dateStart, dateEnd, fmt.Errorf("invalid date range format. Expected 'YYYY-MM-DD_to_YYYY-MM-DD', got '%s'", dateRange)
	}

	if !strings.EqualFold(dates[0], "today") {
		dateStart, err = time.ParseInLocation(time.DateOnly, dates[0], time.Local)
		if err != nil {
			return dateStart, dateEnd, err
		}
	}

	if !strings.EqualFold(dates[1], "today") {
		dateEnd, err = time.ParseInLocation(time.DateOnly, dates[1], time.Local)
		if err != nil {
			return dateStart, dateEnd, err
		}
	}

	if dateEnd.Before(dateStart) {
		dateStart, dateEnd = dateEnd, dateStart
	}

	dateEnd = time.Date(dateEnd.Year(), dateEnd.Month(), dateEnd.Day(), 23, 59, 59, 999999999, dateEnd.Location())

	return dateStart, dateEnd, nil
}

// extractDays extracts a number from a string using regex.
// Returns the first number found in the string or an error if no number is found.
func extractDays(s string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(s)
	if match == "" {
		return 0, errors.New("no number found")
	}
	return strconv.Atoi(match)
}

// processLastDays takes a date range string in the format "last_X" where X is a number of days
// and returns a time range from X days ago to now.
// Returns an error if the number of days cannot be extracted from the string.
func processLastDays(dateRange string) (time.Time, time.Time, error) {

	dateStart, dateEnd := processTodayDateRange()

	days, err := extractDays(dateRange)
	if err != nil {
		return dateStart, dateEnd, err
	}

	dateStart = dateStart.AddDate(0, 0, -days)

	return dateStart, dateEnd, nil
}
