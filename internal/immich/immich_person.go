package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/google/go-querystring/query"
)

// PersonImageCount returns the number of images associated with a specific person in Immich.
func (a *Asset) PersonImageCount(personID, requestID, deviceID string) (int, error) {

	var personStatistics PersonStatistics

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(personStatistics, err, nil, "")
		return 0, err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "people", personID, "statistics"),
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, personStatistics)
	body, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		_, _, err = immichAPIFail(personStatistics, err, body, apiURL.String())
		return 0, err
	}

	err = json.Unmarshal(body, &personStatistics)
	if err != nil {
		_, _, err = immichAPIFail(personStatistics, err, body, apiURL.String())
		return 0, err
	}

	return personStatistics.Assets, err
}

// RandomImageOfPerson retrieves a random image for a given person from the Immich API.
// It handles retries, caching, and filtering to find suitable images. The function will make
// multiple attempts to find a valid image that matches the criteria (not trashed, correct type, etc).
// If caching is enabled, it will maintain a cache of unused images for future requests.
//
// Parameters:
//   - personID: The ID of the person whose images to search for
//   - requestID: The ID of the API request for tracking purposes
//   - deviceID: The ID of the device making the request
//
// Returns:
//   - error: nil if successful, error otherwise. Returns specific error if no suitable
//     image is found after MaxRetries attempts or if there are API/parsing failures
//
// The function mutates the receiver (i *ImmichAsset) to store the selected image if successful.
func (a *Asset) RandomImageOfPerson(personID, requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random image of", personID)
	} else {
		log.Debug(requestID+" Getting Random image of", personID)
	}

	for range MaxRetries {

		var immichAssets []Asset

		u, err := url.Parse(a.requestConfig.ImmichURL)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, nil, "")
			return err
		}

		requestBody := SearchRandomBody{
			PersonIDs:  []string{personID},
			Type:       string(ImageType),
			WithExif:   true,
			WithPeople: true,
			Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
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

		jsonBody, bodyMarshalErr := json.Marshal(requestBody)
		if bodyMarshalErr != nil {
			_, _, bodyMarshalErr = immichAPIFail(immichAssets, bodyMarshalErr, nil, apiURL.String())
			return bodyMarshalErr
		}

		immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, immichAssets)
		apiBody, err := immichAPICall(a.ctx, http.MethodPost, apiURL.String(), jsonBody)
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
			log.Debug(requestID + " No images left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourcePerson
			asset.requestConfig = a.requestConfig
			asset.ctx = a.ctx

			if !asset.isValidAsset(requestID, deviceID, ImageOnlyAssetTypes, a.RatioWanted) {
				continue
			}

			if a.requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				immichAssetsToCache := slices.Delete(immichAssets, immichAssetIndex, immichAssetIndex+1)
				jsonBytes, cacheMarshalErr := json.Marshal(immichAssetsToCache)
				if cacheMarshalErr != nil {
					log.Error("Failed to marshal immichAssetsToCache", "error", cacheMarshalErr)
					return cacheMarshalErr
				}

				// Replace cache with remaining images after removing used image(s)
				cacheErr := cache.Replace(apiCacheKey, jsonBytes)
				if cacheErr != nil {
					log.Debug("cache not found!")
				}
			}

			asset.BucketID = personID

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}
	return fmt.Errorf("no images found for person '%s'. Max retries reached", personID)
}
