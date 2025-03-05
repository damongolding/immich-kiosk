package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/google/go-querystring/query"
)

type Tags []Tag

// Get retrieves a Tag from the Tags slice by searching for a matching value or ID (case-insensitive).
// Returns the matching Tag and nil error if found, or empty Tag and error if not found.
// The tagValue parameter can be either the tag's text value or ID.
func (t Tags) Get(tagValue string) (Tag, error) {

	tagValue, err := url.PathUnescape(tagValue)
	if err != nil {
		return Tag{}, err
	}

	for _, tag := range t {
		if strings.EqualFold(tag.Value, tagValue) || strings.EqualFold(tag.ID, tagValue) {
			return tag, nil
		}
	}

	return Tag{}, fmt.Errorf("tag not found. tag=%s", tagValue)
}

// AllTags retrieves all tags from the Immich API.
// It returns the list of tags, the API URL used, and any error encountered.
// The requestID and deviceID are used for caching and logging purposes.
func (i *Asset) AllTags(requestID, deviceID string) (Tags, string, error) {
	var tags []Tag

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(tags, err, nil, "")
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "tags"),
	}

	immichAPICall := withImmichAPICache(i.immichAPICall, requestID, deviceID, i.requestConfig, tags)
	body, err := immichAPICall(i.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return immichAPIFail(tags, err, body, apiURL.String())
	}

	err = json.Unmarshal(body, &tags)
	if err != nil {
		return immichAPIFail(tags, err, body, apiURL.String())
	}

	return tags, apiURL.String(), nil
}

// AssetsWithTagCount returns the total number of assets that have the specified tag.
// The tagID parameter is the unique identifier of the tag to count assets for.
// The requestID and deviceID are used for caching and logging purposes.
func (i *Asset) AssetsWithTagCount(tagID string, requestID, deviceID string) (int, error) {

	var allAssetsCount int

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(allAssetsCount, err, nil, "")
		return allAssetsCount, err
	}

	requestBody := SearchRandomBody{
		Type:       string(ImageType),
		TagIDs:     []string{tagID},
		WithPeople: false,
		WithExif:   false,
		Size:       i.requestConfig.Kiosk.FetchedAssetsSize,
	}

	if i.requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	DateFilter(&requestBody, i.requestConfig.DateFilter)

	allAssetsCount, err = i.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)

	return allAssetsCount, err
}

// AssetsWithTag retrieves assets that have the specified tag from the Immich API.
// The tagID parameter is the unique identifier of the tag to find assets for.
// The requestID and deviceID are used for caching and logging purposes.
// It returns the list of assets, the API URL used, and any error encountered.
func (i *Asset) AssetsWithTag(tagID string, requestID, deviceID string) ([]Asset, string, error) {

	var immichAssets []Asset

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(immichAssets, err, nil, "")
	}

	requestBody := SearchRandomBody{
		Type:       string(ImageType),
		TagIDs:     []string{tagID},
		WithExif:   true,
		WithPeople: true,
		Size:       i.requestConfig.Kiosk.FetchedAssetsSize,
	}

	if i.requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	DateFilter(&requestBody, i.requestConfig.DateFilter)

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

	immichAPICall := withImmichAPICache(i.immichAPICall, requestID, deviceID, i.requestConfig, immichAssets)
	apiBody, err := immichAPICall(i.ctx, http.MethodPost, apiURL.String(), jsonBody)
	if err != nil {
		return immichAPIFail(immichAssets, err, nil, apiURL.String())
	}

	err = json.Unmarshal(apiBody, &immichAssets)
	if err != nil {
		return immichAPIFail(immichAssets, err, nil, apiURL.String())
	}

	return immichAssets, apiURL.String(), nil
}

// RandomAssetWithTag selects a random asset that has the specified tag.
// The tagID parameter is the unique identifier of the tag to find assets for.
// The requestID and deviceID are used for caching and logging purposes.
// The isPrefetch parameter indicates if this is a prefetch request.
// The method updates the receiver Asset with the randomly selected asset's data.
func (i *Asset) RandomAssetWithTag(tagID string, requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random image with tag", tagID)
	} else {
		log.Debug(requestID+" Getting Random image with tag", tagID)
	}

	for range MaxRetries {

		immichAssets, apiURL, err := i.AssetsWithTag(tagID, requestID, deviceID)
		if err != nil {
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL, deviceID, i.requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No images left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)

			immichAssetsRetry, _, retryErr := i.AssetsWithTag(tagID, requestID, deviceID)
			if retryErr != nil || len(immichAssetsRetry) == 0 {
				return fmt.Errorf("no assets found with tag %s after refresh", tagID)
			}

			continue
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourceTag
			asset.requestConfig = i.requestConfig
			asset.ctx = i.ctx

			if !asset.isValidAsset(requestID, deviceID, ImageOnlyAssetTypes, i.RatioWanted) {
				continue
			}

			if i.requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				immichAssetsToCache := slices.Delete(immichAssets, immichAssetIndex, immichAssetIndex+1)
				jsonBytes, cacheMarshalErr := json.Marshal(immichAssetsToCache)
				if cacheMarshalErr != nil {
					log.Error("Failed to marshal immichAssetsToCache", "error", cacheMarshalErr)
					return cacheMarshalErr
				}

				// replace cache with used image(s) removed
				cacheErr := cache.Replace(apiCacheKey, jsonBytes)
				if cacheErr != nil {
					log.Debug("Failed to update device cache for tag", "tagID", tagID, "deviceID", deviceID)
				}
			}

			asset.BucketID = tagID

			*i = asset

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("no images found for '%s'. Max retries reached", tagID)
}

func (i *Asset) AddTag(tagID string) {
	tags, _, tagsError := i.AllTags("", "")
	if tagsError != nil {

	}

	skipTag, tagFound := tags.Get(tagID)
	if tagFound != nil {
		// tag not found, create it
		// return skipTagError
	}

	i.addTagToAsset(skipTag, i.ID)
}

func (i *Asset) addTagToAsset(skipTag Tag, assetID string) error {

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return fmt.Errorf("parsing url: %w", err)
	}

	requestBody := TagAssetsBody{
		IDs: []string{skipTag.ID},
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("tags", skipTag.ID, "assets"),
	}

	jsonBody, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return fmt.Errorf("marshaling request body: %w", marshalErr)
	}

	apiBody, resErr := i.immichAPICall(i.ctx, http.MethodPut, apiURL.String(), jsonBody)
	if resErr != nil {
		log.Error("Failed to add tag to asset", "error", resErr)
	}

	err = json.Unmarshal(apiBody, &immichAssets)
	if err != nil {
		_, _, err = immichAPIFail(immichAssets, err, apiBody, apiURL.String())
		return err
	}
}
