package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/google/go-querystring/query"
)

func (a *Asset) AssetsWithRatingCount(rating float32, requestID, deviceID string) (int, error) {

	var totalAssetsCount int

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(totalAssetsCount, err, nil, "")
		return totalAssetsCount, err
	}

	requestBody := SearchRandomBody{
		Type:       string(ImageType),
		Rating:     &rating,
		WithPeople: false,
		WithExif:   false,
		Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
	}

	// Include videos if show videos is enabled
	if a.requestConfig.ShowVideos {
		requestBody.Type = ""
	}

	if a.requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	DateFilter(&requestBody, a.requestConfig.DateFilter)

	allAssetsCount, assetsErr := a.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)
	if assetsErr != nil {
		return totalAssetsCount, assetsErr
	}

	totalAssetsCount += allAssetsCount

	return totalAssetsCount, nil
}

func (a *Asset) AssetsWithRating(rating float32, requestID, deviceID string) ([]Asset, string, error) {

	var immichAssets []Asset

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(immichAssets, err, nil, "")
	}

	requestBody := SearchRandomBody{
		Type:       string(ImageType),
		Rating:     &rating,
		WithExif:   true,
		WithPeople: true,
		Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
	}

	// Include videos if show videos is enabled
	if a.requestConfig.ShowVideos {
		requestBody.Type = ""
	}

	if a.requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	DateFilter(&requestBody, a.requestConfig.DateFilter)

	// convert body to queries so url is unique and can be cached
	queries, _ := query.Values(requestBody)

	apiURL := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "api/search/random",
		RawQuery: fmt.Sprintf("kiosk=%x", sha256.Sum256([]byte(queries.Encode()))),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return immichAPIFail(immichAssets, err, nil, apiURL.String())
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, immichAssets)
	apiBody, _, err := immichAPICall(a.ctx, http.MethodPost, apiURL.String(), jsonBody)
	if err != nil {
		return immichAPIFail(immichAssets, err, nil, apiURL.String())
	}

	err = json.Unmarshal(apiBody, &immichAssets)
	if err != nil {
		return immichAPIFail(immichAssets, err, nil, apiURL.String())
	}

	return immichAssets, apiURL.String(), nil
}

func (a *Asset) RandomAssetWithRating(ratingID string, requestID, deviceID string, isPrefetch bool) error {

	_, ratingStr, ok := strings.Cut(ratingID, "-")
	if !ok {
		return fmt.Errorf("invalid rating format (cut): %s", ratingID)
	}

	rating64, err := strconv.ParseFloat(ratingStr, 64)
	if err != nil {
		return fmt.Errorf("invalid rating format (parse): %s", ratingID)
	}

	rating := float32(rating64)

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random asset with rating", rating)
	} else {
		log.Debug(requestID+" Getting Random asset with", "rating", rating)
	}

	for range MaxRetries {

		immichAssets, apiURL, err := a.AssetsWithRating(rating, requestID, deviceID)
		if err != nil {
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL, deviceID, a.requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No assets left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)

			immichAssetsRetry, _, retryErr := a.AssetsWithRating(rating, requestID, deviceID)
			if retryErr != nil || len(immichAssetsRetry) == 0 {
				return fmt.Errorf("no assets found with rating %f after refresh", rating)
			}

			immichAssets = immichAssetsRetry
		}

		wantedAssetType := ImageOnlyAssetTypes
		if a.requestConfig.ShowVideos {
			wantedAssetType = AllAssetTypes
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourceRating
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

			asset.BucketID = fmt.Sprintf("rating-%.2f", rating)

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable assets left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("no assets found with rating '%.2f'. Max retries reached", rating)
}
