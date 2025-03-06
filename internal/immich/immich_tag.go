package immich

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
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

// AddTag adds a tag to the asset. If the tag doesn't exist, it will be created first.
// It first checks if the tag exists by calling AllTags. If not found, it creates the tag
// via upsertTag. Finally it associates the tag with the asset.
func (i *Asset) AddTag(tag Tag) error {
	tags, _, tagsError := i.AllTags("", "")
	if tagsError != nil {
		return tagsError
	}

	foundTag, tagFound := tags.Get(tag.Name)
	if tagFound != nil {
		createdTag, createdTagErr := i.upsertTag(tag)
		if createdTagErr != nil {
			return createdTagErr
		}

		foundTag = createdTag
	}

	return i.addTagToAsset(foundTag, i.ID)
}

// upsertTag creates a new tag in the Immich system if it doesn't already exist.
// It makes a PUT request to the tags API endpoint with the tag name.
// Returns the created/existing tag and any error encountered.
func (i *Asset) upsertTag(tag Tag) (Tag, error) {

	var response UpsertTagResponse
	var createdTag Tag

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return createdTag, fmt.Errorf("parsing url: %w", err)
	}

	requestBody := UpsertTagBody{
		Tags: []string{tag.Name},
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/tags",
	}

	jsonBody, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return createdTag, fmt.Errorf("marshaling request body: %w", marshalErr)
	}

	apiBody, resErr := i.immichAPICall(i.ctx, http.MethodPut, apiURL.String(), jsonBody)
	if resErr != nil {
		log.Error("Failed to add tag to asset", "error", resErr)
		return createdTag, resErr
	}

	err = json.Unmarshal(apiBody, &response)
	if err != nil {
		_, _, err = immichAPIFail(response, err, apiBody, apiURL.String())
		return createdTag, err
	}

	if len(response) == 0 && response[0].ID == "" {
		log.Error("failed to create tag", "response", response, "error", err)
		return createdTag, errors.New("failed to create tag")
	}

	createdTag = Tag{
		ID:    response[0].ID,
		Name:  response[0].Name,
		Value: response[0].Value,
	}

	return createdTag, nil
}

// addTagToAsset associates a tag with an asset in the Immich system.
// It makes a PUT request to associate the tag ID with the asset ID.
// Returns an error if the association fails.
func (i *Asset) addTagToAsset(tag Tag, assetID string) error {

	var response TagAssetsResponse

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return fmt.Errorf("parsing url: %w", err)
	}

	requestBody := TagAssetsBody{
		IDs: []string{assetID},
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "tags", tag.ID, "assets"),
	}

	jsonBody, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return fmt.Errorf("marshaling request body: %w", marshalErr)
	}

	apiBody, resErr := i.immichAPICall(i.ctx, http.MethodPut, apiURL.String(), jsonBody)
	if resErr != nil {
		log.Error("Failed to add tag to asset", "error", resErr)
	}

	err = json.Unmarshal(apiBody, &response)
	if err != nil {
		_, _, err = immichAPIFail(response, err, apiBody, apiURL.String())
		return err
	}

	if len(response) == 0 {
		return fmt.Errorf("failed to add tag to asset: %s", tag.ID)
	}

	if !response[0].Success {
		return fmt.Errorf("failed to add tag to asset: %s", response[0].Error)
	}

	return nil
}
