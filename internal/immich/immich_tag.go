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
func (a *Asset) AllTags(requestID, deviceID string) (Tags, string, error) {
	var tags []Tag

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(tags, err, nil, "")
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "tags"),
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, tags)
	body, _, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
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
func (a *Asset) AssetsWithTagCount(tagID string, requestID, deviceID string) (int, error) {

	var totalAssetsCount int

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(totalAssetsCount, err, nil, "")
		return totalAssetsCount, err
	}

	requestBody := SearchRandomBody{
		Type:       string(ImageType),
		TagIDs:     []string{tagID},
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

// AssetsWithTag retrieves assets that have the specified tag from the Immich API.
// The tagID parameter is the unique identifier of the tag to find assets for.
// The requestID and deviceID are used for caching and logging purposes.
// It returns the list of assets, the API URL used, and any error encountered.
func (a *Asset) AssetsWithTag(tagID string, requestID, deviceID string) ([]Asset, string, error) {

	var immichAssets []Asset

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(immichAssets, err, nil, "")
	}

	requestBody := SearchRandomBody{
		Type:       string(ImageType),
		TagIDs:     []string{tagID},
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

// RandomAssetWithTag selects a random asset that has the specified tag.
// The tagID parameter is the unique identifier of the tag to find assets for.
// The requestID and deviceID are used for caching and logging purposes.
// The isPrefetch parameter indicates if this is a prefetch request.
// The method updates the receiver Asset with the randomly selected asset's data.
func (a *Asset) RandomAssetWithTag(tagID string, requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random asset with tag", "ID", tagID)
	} else {
		log.Debug(requestID+" Getting Random asset with tag", "ID", tagID)
	}

	for range MaxRetries {

		immichAssets, apiURL, err := a.AssetsWithTag(tagID, requestID, deviceID)
		if err != nil {
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL, deviceID, a.requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No assets left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)

			immichAssetsRetry, _, retryErr := a.AssetsWithTag(tagID, requestID, deviceID)
			if retryErr != nil || len(immichAssetsRetry) == 0 {
				return fmt.Errorf("no assets found with tag %s after refresh", tagID)
			}

			immichAssets = immichAssetsRetry
		}

		wantedAssetType := ImageOnlyAssetTypes
		if a.requestConfig.ShowVideos {
			wantedAssetType = AllAssetTypes
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourceTag
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

			asset.BucketID = tagID

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable assets left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("no assets found for '%s'. Max retries reached", tagID)
}

func (a *Asset) HasTag(tagID string) bool {
	_, tagGetErr := a.Tags.Get(tagID)
	return tagGetErr == nil
}

// AddTag adds a tag to the asset. If the tag doesn't exist, it will be created first.
// It first checks if the tag exists by calling AllTags. If not found, it creates the tag
// via upsertTag. Finally it associates the tag with the asset.
func (a *Asset) AddTag(tag Tag) error {
	tags, _, tagsError := a.AllTags("", "")
	if tagsError != nil {
		return tagsError
	}

	foundTag, tagGetErr := tags.Get(tag.Name)
	if tagGetErr != nil {
		createdTag, createdTagErr := a.upsertTag(tag)
		if createdTagErr != nil {
			return createdTagErr
		}

		foundTag = createdTag
	}

	return a.addTagToAsset(foundTag, a.ID)
}

func (a *Asset) RemoveTag(tag Tag) error {
	tags, _, tagsError := a.AllTags("", "")
	if tagsError != nil {
		return tagsError
	}

	foundTag, tagGetErr := tags.Get(tag.Name)
	if tagGetErr != nil {
		createdTag, createdTagErr := a.upsertTag(tag)
		if createdTagErr != nil {
			return createdTagErr
		}

		foundTag = createdTag
	}

	return a.removeTagFromAsset(foundTag, a.ID)
}

// upsertTag creates a new tag in the Immich system if it doesn't already exist.
// It makes a PUT request to the tags API endpoint with the tag name.
// Returns the created/existing tag and any error encountered.
func (a *Asset) upsertTag(tag Tag) (Tag, error) {

	var response UpsertTagResponse
	var createdTag Tag

	u, err := url.Parse(a.requestConfig.ImmichURL)
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

	apiBody, _, resErr := a.immichAPICall(a.ctx, http.MethodPut, apiURL.String(), jsonBody)
	if resErr != nil {
		_, _, resErr = immichAPIFail(response, resErr, apiBody, apiURL.String())
		log.Error("api call failed while creating tag", "error", resErr)
		return createdTag, resErr
	}

	err = json.Unmarshal(apiBody, &response)
	if err != nil {
		_, _, err = immichAPIFail(response, err, apiBody, apiURL.String())
		return createdTag, err
	}

	if len(response) == 0 || (len(response) > 0 && response[0].ID == "") {
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

// modifyTagAsset performs a tag modification operation on an asset in the Immich system.
// It makes an HTTP request to modify the association between a tag and an asset.
// Parameters:
//   - tag: The Tag object containing the tag information
//   - assetID: The ID of the asset to modify
//   - method: The HTTP method to use (PUT for add, DELETE for remove)
//   - action: Description of the action being performed for error messages
//
// Returns an error if the modification fails.
func (a *Asset) modifyTagAsset(tag Tag, assetID string, method string, action string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var response TagAssetsResponse

	u, err := url.Parse(a.requestConfig.ImmichURL)
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

	apiBody, _, resErr := a.immichAPICall(a.ctx, method, apiURL.String(), jsonBody)
	if resErr != nil {
		_, _, resErr = immichAPIFail(response, resErr, apiBody, apiURL.String())
		log.Error("api failed to "+action+" tag to asset", "error", resErr)
		return resErr
	}

	err = json.Unmarshal(apiBody, &response)
	if err != nil {
		_, _, err = immichAPIFail(response, err, apiBody, apiURL.String())
		return err
	}

	if len(response) == 0 {
		return fmt.Errorf("failed to "+action+" tag from asset: %s", tag.ID)
	}

	if !response[0].Success {
		return fmt.Errorf("failed to "+action+" tag from asset: %s", response[0].Error)
	}

	// remove asset data from cache as we've changed its tags
	cacheErr := a.RemoveAssetCache(a.DeviceID)
	if cacheErr != nil {
		log.Error("error removing asset from cache", "assetID", assetID, "error", cacheErr)
	}

	return nil
}

// addTagToAsset associates a tag with an asset in the Immich system.
// It makes a PUT request to associate the tag ID with the asset ID.
// Parameters:
//   - tag: The Tag object to associate with the asset
//   - assetID: The ID of the asset to tag
//
// Returns an error if the association fails.
func (a *Asset) addTagToAsset(tag Tag, assetID string) error {
	return a.modifyTagAsset(tag, assetID, http.MethodPut, "add")
}

// removeTagFromAsset removes a tag association from an asset in the Immich system.
// It makes a DELETE request to remove the association between the tag ID and asset ID.
// Parameters:
//   - tag: The Tag object to remove from the asset
//   - assetID: The ID of the asset to remove the tag from
//
// Returns an error if the removal fails.
func (a *Asset) removeTagFromAsset(tag Tag, assetID string) error {
	return a.modifyTagAsset(tag, assetID, http.MethodDelete, "remove")
}
