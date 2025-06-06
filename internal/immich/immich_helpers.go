package immich

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/demo"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/google/go-querystring/query"
)

// immichAPIFail handles failures in Immich API calls by unmarshaling the error response,
// logging the error, and returning a formatted error along with the original value.
func immichAPIFail[T APIResponse](value T, err error, body []byte, apiURL string) (T, string, error) {
	var immichError Error
	errorUnmarshalErr := json.Unmarshal(body, &immichError)
	if errorUnmarshalErr != nil {
		log.Error("Couldn't read error", "body", string(body), "url", apiURL)
		return value, apiURL, err
	}
	log.Errorf("%s : %v", immichError.Error, immichError.Message)
	return value, apiURL, fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
}

// withImmichAPICache Decorator to implement cache for the immichAPICall func
func withImmichAPICache[T APIResponse](immichAPICall apiCall, requestID, deviceID string, requestConfig config.Config, jsonShape T) apiCall {
	return func(ctx context.Context, method, apiURL string, body []byte, headers ...map[string]string) ([]byte, error) {

		if !requestConfig.Kiosk.Cache {
			return immichAPICall(ctx, method, apiURL, body, headers...)
		}

		apiCacheKey := cache.APICacheKey(apiURL, deviceID, requestConfig.SelectedUser)

		if apiData, found := cache.Get(apiCacheKey); found {
			log.Debug(requestID+" Cache hit", "url", apiURL)
			data, ok := apiData.([]byte)
			if !ok {
				return nil, errors.New("cache data type assertion failed")
			}
			return data, nil
		}

		if requestConfig.Kiosk.DebugVerbose {
			log.Debug(requestID+" Cache miss", "url", apiURL)
		}

		apiBody, err := immichAPICall(ctx, method, apiURL, body)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// Unpack api json into struct which discards data we don't use (for smaller cache size)
		err = json.Unmarshal(apiBody, &jsonShape)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// get bytes and store in cache
		jsonBytes, err := json.Marshal(jsonShape)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		cache.Set(apiCacheKey, jsonBytes)
		if requestConfig.Kiosk.DebugVerbose {
			log.Debug(requestID+" Cache saved", "url", apiURL)
		}

		return jsonBytes, nil
	}
}

// immichAPICall bootstrap for immich api call
func (a *Asset) immichAPICall(ctx context.Context, method, apiURL string, body []byte, headers ...map[string]string) ([]byte, error) {

	var responseBody []byte
	var lastErr error

	_, err := url.Parse(apiURL)
	if err != nil {
		log.Error("Invalid URL", "url", apiURL, "err", err)
		return responseBody, err
	}

	for attempts := range 3 {

		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, reqErr := http.NewRequestWithContext(ctx, method, apiURL, bodyReader)
		if reqErr != nil {
			log.Error(reqErr)
			return responseBody, reqErr
		}

		req.Header.Set("Accept", "application/json")

		switch a.requestConfig.Kiosk.DemoMode {
		case true:
			token, demoLoginErr := demo.Login(a.ctx, false)
			if demoLoginErr != nil {
				log.Error(demoLoginErr)
				return responseBody, demoLoginErr
			}
			req.Header.Set("Authorization", "Bearer "+token)

		default:
			apiKey := a.requestConfig.ImmichAPIKey
			if a.requestConfig.SelectedUser != "" {
				if key, ok := a.requestConfig.ImmichUsersAPIKeys[a.requestConfig.SelectedUser]; ok {
					apiKey = key
				} else {
					return responseBody, fmt.Errorf("no API key found for user %s in the config", a.requestConfig.SelectedUser)
				}
			}

			req.Header.Set("x-api-key", apiKey)
		}

		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
			req.Header.Set("Content-Type", "application/json")
		}

		// Add any additional headers
		for _, header := range headers {
			for key, value := range header {
				req.Header.Set(key, value)
			}
		}

		res, resErr := HTTPClient.Do(req)
		if resErr != nil {
			lastErr = resErr

			// Type assert to get more details about the error
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				log.Error("Request failed",
					"attempt", attempts,
					"URL", apiURL,
					"method", method,
					"operation", urlErr.Op,
					"error_type", fmt.Sprintf("%T", urlErr.Err),
					"error", urlErr.Err)
			} else {
				log.Error("Request failed",
					"attempt", attempts,
					"URL", apiURL,
					"method", method,
					"error_type", fmt.Sprintf("%T", resErr),
					"error", resErr)
			}
			time.Sleep(time.Duration(1<<attempts) * time.Second)
			continue
		}

		defer res.Body.Close()

		// in demo mode and unauthorized, attempt to login again
		if res.StatusCode == http.StatusUnauthorized && a.requestConfig.Kiosk.DemoMode {
			_, _ = io.Copy(io.Discard, res.Body)
			if !demo.ValidateToken(a.ctx, demo.DemoToken) {
				_, err = demo.Login(a.ctx, true)
				if err != nil {
					return responseBody, err
				}
				continue
			}
		}

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			responseBody, err = io.ReadAll(res.Body)
			if err != nil {
				log.Error("reading unexpected response body", "method", method, "url", apiURL, "err", err)
				return responseBody, err
			}

			if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
				return responseBody, fmt.Errorf("received %d (unauthorised) code from Immich. Please check your Immich API is correct", res.StatusCode)
			}

			return responseBody, fmt.Errorf("HTTP %d: unexpected status code", res.StatusCode)
		}

		responseBody, err = io.ReadAll(res.Body)
		if err != nil {
			log.Error("reading response body", "method", method, "url", apiURL, "err", err)
			return responseBody, err
		}

		return responseBody, nil
	}

	return responseBody, fmt.Errorf("request failed: max retries exceeded. last err=%w", lastErr)
}

// ratioCheck checks if an image's orientation matches a desired ratio.
// First, it computes the image's ratio (portrait/landscape) by calling addRatio().
// Then it checks if the image matches any desired ratio requirements:
// - If no specific ratio is required (RatioWanted is empty), returns true
// - If RatioWanted is "portrait", returns true only if image is portrait
// - If RatioWanted is "landscape", returns true only if image is landscape
// - Otherwise returns false if orientations don't match
func (a *Asset) ratioCheck(wantedRatio ImageOrientation) bool {

	a.AddRatio()

	// specific ratio is not wanted
	if wantedRatio == "" {
		return true
	}

	if (wantedRatio == PortraitOrientation && a.IsPortrait) ||
		(wantedRatio == LandscapeOrientation && a.IsLandscape) {
		return true
	}

	return false
}

// AddRatio determines the ratio (portrait or landscape) of the image based on its EXIF information.
// It sets the Ratio field in ExifInfo and updates IsPortrait or IsLandscape accordingly.
// For orientations 5, 6, 7, and 8, it considers the image rotated by 90 degrees.
func (a *Asset) AddRatio() {

	switch a.ExifInfo.Orientation {
	case "5", "6", "7", "8":
		// For these orientations, the image is rotated, so we invert the height/width comparison
		if a.ExifInfo.ExifImageHeight < a.ExifInfo.ExifImageWidth {
			a.ExifInfo.ImageOrientation = PortraitOrientation
			a.IsPortrait = true
		} else {
			a.ExifInfo.ImageOrientation = LandscapeOrientation
			a.IsLandscape = true
		}
	default:
		// For all other orientations, including 1, 2, 3, 4, and any unknown orientations
		if a.ExifInfo.ExifImageHeight > a.ExifInfo.ExifImageWidth {
			a.ExifInfo.ImageOrientation = PortraitOrientation
			a.IsPortrait = true
		} else {
			a.ExifInfo.ImageOrientation = LandscapeOrientation
			a.IsLandscape = true
		}
	}
}

// mergeAssetInfo merges additional asset information into the current Asset.
// It uses reflection to examine each field of the current asset and updates
// field values based on the following rules:
// - For boolean fields: Always copies the value from additionalInfo
// - For all other fields: Only copies from additionalInfo if the current value is zero
//
// This allows for selective updating of asset information while preserving existing
// non-zero values for most fields, except booleans which are always updated.
//
// Parameters:
//   - additionalInfo: The Asset containing the additional information to merge
//
// Returns:
//   - error: If any field in additionalInfo is invalid during the merge process
func (a *Asset) mergeAssetInfo(additionalInfo Asset) error {

	v := reflect.ValueOf(a).Elem()
	d := reflect.ValueOf(additionalInfo)
	t := v.Type()

	for i := range v.NumField() {
		field := v.Field(i)
		fieldName := t.Field(i).Name
		if !field.CanSet() {
			continue
		}

		additionalField := d.FieldByName(fieldName)
		if !additionalField.IsValid() {
			return fmt.Errorf("invalid field: %s", fieldName)
		}

		if field.Kind() == reflect.Bool {
			field.Set(additionalField)
			continue
		}

		if field.Kind() == reflect.Slice {
			if field.Len() > 0 {
				continue // Don't overwrite non-empty slices
			}
			if !additionalField.IsNil() {
				field.Set(additionalField)
			}
			continue
		}

		if reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			field.Set(additionalField)
		}
	}

	return nil
}

// AssetInfo fetches the image information from Immich
func (a *Asset) AssetInfo(requestID, deviceID string) error {

	var immichAsset Asset

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", a.ID),
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, immichAsset)
	body, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		_, _, err = immichAPIFail(immichAsset, err, body, apiURL.String())
		return fmt.Errorf("fetching asset info: err %w", err)
	}

	err = json.Unmarshal(body, &immichAsset)
	if err != nil {
		_, _, err = immichAPIFail(immichAsset, err, body, apiURL.String())
		return fmt.Errorf("fetching asset info: err %w", err)
	}

	return a.mergeAssetInfo(immichAsset)
}

// ImagePreview fetches the raw image data from Immich
func (a *Asset) ImagePreview() ([]byte, error) {

	var bytes []byte

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		log.Error(err)
		return bytes, err
	}

	assetSize := AssetSizeThumbnail
	if a.requestConfig.UseOriginalImage && slices.Contains(supportedImageMimeTypes, a.OriginalMimeType) {
		assetSize = AssetSizeOriginal
	}

	apiURL := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     path.Join("api", "assets", a.ID, assetSize),
		RawQuery: "size=preview",
	}

	return a.immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
}

// FacesCenterPoint calculates the center point of all detected faces in an image as percentages.
// It analyzes both assigned (People) and unassigned faces, finding the bounding box that encompasses
// all faces and returning its center as x,y percentages relative to the image dimensions.
// Returns (0,0) if no faces are detected or if image dimensions are invalid.
func (a *Asset) FacesCenterPoint() (float64, float64) {
	if len(a.People) == 0 && len(a.UnassignedFaces) == 0 {
		return 0, 0
	}

	var minX, minY, maxX, maxY int
	initialized := false

	for _, person := range a.People {
		for _, face := range person.Faces {
			if face.BoundingBoxX1 == 0 && face.BoundingBoxY1 == 0 &&
				face.BoundingBoxX2 == 0 && face.BoundingBoxY2 == 0 {
				continue
			}

			if !initialized {
				minX, minY = face.BoundingBoxX1, face.BoundingBoxY1
				maxX, maxY = face.BoundingBoxX2, face.BoundingBoxY2
				initialized = true
				continue
			}

			minX = min(minX, face.BoundingBoxX1)
			minY = min(minY, face.BoundingBoxY1)
			maxX = max(maxX, face.BoundingBoxX2)
			maxY = max(maxY, face.BoundingBoxY2)
		}
	}

	for _, face := range a.UnassignedFaces {
		if face.BoundingBoxX1 == 0 && face.BoundingBoxY1 == 0 &&
			face.BoundingBoxX2 == 0 && face.BoundingBoxY2 == 0 {
			continue
		}

		if !initialized {
			minX, minY = face.BoundingBoxX1, face.BoundingBoxY1
			maxX, maxY = face.BoundingBoxX2, face.BoundingBoxY2
			initialized = true
			continue
		}

		minX = min(minX, face.BoundingBoxX1)
		minY = min(minY, face.BoundingBoxY1)
		maxX = max(maxX, face.BoundingBoxX2)
		maxY = max(maxY, face.BoundingBoxY2)
	}

	if !initialized {
		return 0, 0
	}

	centerX := float64(minX+maxX) / 2
	centerY := float64(minY+maxY) / 2

	var percentX, percentY float64
	var imageWidth, imageHeight int

	switch {
	case len(a.People) != 0 && len(a.People[0].Faces) != 0:
		imageWidth = a.People[0].Faces[0].ImageWidth
		imageHeight = a.People[0].Faces[0].ImageHeight
	case len(a.UnassignedFaces) != 0:
		imageWidth = a.UnassignedFaces[0].ImageWidth
		imageHeight = a.UnassignedFaces[0].ImageHeight
	default:
		return 0, 0
	}

	if imageWidth == 0 || imageHeight == 0 {
		return 0, 0
	}

	percentX = centerX / float64(imageWidth) * 100
	percentY = centerY / float64(imageHeight) * 100

	return percentX, percentY
}

// FacesCenterPointPX calculates the center point of all detected faces in an image in pixels.
// It analyzes both assigned (People) and unassigned faces, finding the bounding box that encompasses
// all faces and returning its center as x,y pixel coordinates.
// Returns (0,0) if no faces are detected or if all bounding boxes are empty.
func (a *Asset) FacesCenterPointPX() (float64, float64) {
	if len(a.People) == 0 && len(a.UnassignedFaces) == 0 {
		return 0, 0
	}

	var minX, minY, maxX, maxY int
	initialized := false

	for _, person := range a.People {
		for _, face := range person.Faces {
			if face.BoundingBoxX1 == 0 && face.BoundingBoxY1 == 0 &&
				face.BoundingBoxX2 == 0 && face.BoundingBoxY2 == 0 {
				continue
			}

			if !initialized {
				minX, minY = face.BoundingBoxX1, face.BoundingBoxY1
				maxX, maxY = face.BoundingBoxX2, face.BoundingBoxY2
				initialized = true
				continue
			}

			minX = min(minX, face.BoundingBoxX1)
			minY = min(minY, face.BoundingBoxY1)
			maxX = max(maxX, face.BoundingBoxX2)
			maxY = max(maxY, face.BoundingBoxY2)
		}
	}

	for _, face := range a.UnassignedFaces {
		if face.BoundingBoxX1 == 0 && face.BoundingBoxY1 == 0 &&
			face.BoundingBoxX2 == 0 && face.BoundingBoxY2 == 0 {
			continue
		}

		if !initialized {
			minX, minY = face.BoundingBoxX1, face.BoundingBoxY1
			maxX, maxY = face.BoundingBoxX2, face.BoundingBoxY2
			initialized = true
			continue
		}

		minX = min(minX, face.BoundingBoxX1)
		minY = min(minY, face.BoundingBoxY1)
		maxX = max(maxX, face.BoundingBoxX2)
		maxY = max(maxY, face.BoundingBoxY2)
	}

	if !initialized {
		return 0, 0
	}

	centerX := float64(minX+maxX) / 2
	centerY := float64(minY+maxY) / 2

	return centerX, centerY
}

// containsTag checks if an asset has a specific tag (case-insensitive).
// It iterates through the asset's tags and compares the given tagName
// with each tag's name, ignoring case.
//
// Parameters:
//   - tagName: The name of the tag to search for (case-insensitive)
//
// Returns:
//   - bool: true if the tag is found, false otherwise
func (a *Asset) containsTag(tagName string) bool {
	for _, tag := range a.Tags {
		if strings.EqualFold(tag.Name, tagName) {
			return true
		}
	}
	return false
}

// isValidAsset checks if an asset meets all the required criteria for processing.
// It performs a series of validation checks including basic properties, date filters,
// album membership, people detection, and tag validation.
//
// Parameters:
//   - requestID: Unique identifier for the request
//   - deviceID: ID of the device making the request
//   - allowedTypes: Slice of allowed asset types
//   - wantedRatio: Desired image orientation ratio
//
// Returns:
//   - bool: true if asset meets all criteria, false otherwise
func (a *Asset) isValidAsset(requestID, deviceID string, allowedTypes []AssetType, wantedRatio ImageOrientation) bool {
	return a.hasValidBasicProperties(allowedTypes, wantedRatio) &&
		a.hasValidDateFilter() &&
		a.hasValidAlbums(requestID, deviceID) &&
		a.hasValidPeople(requestID, deviceID) &&
		a.hasValidTags(requestID, deviceID)
}

// hasValidBasicProperties checks basic asset properties including type,
// trash status, archive status, aspect ratio and blacklist status.
//
// Parameters:
//   - allowedTypes: Slice of allowed asset types to check against
//   - wantedRatio: Desired image orientation ratio
//
// Returns:
//   - bool: true if basic properties are valid, false otherwise
func (a *Asset) hasValidBasicProperties(allowedTypes []AssetType, wantedRatio ImageOrientation) bool {
	if !slices.Contains(allowedTypes, a.Type) {
		return false
	}
	if a.IsTrashed {
		return false
	}
	if a.IsArchived && !a.requestConfig.ShowArchived {
		return false
	}
	if !a.ratioCheck(wantedRatio) {
		return false
	}
	if slices.Contains(a.requestConfig.Blacklist, a.ID) {
		return false
	}
	return true
}

// hasValidDateFilter validates if the asset's date matches the configured date filter criteria.
// Assets from Memories or DateRange buckets bypass the date filter check.
//
// Returns:
//   - bool: true if date is valid or no filter set, false if outside filter range
func (a *Asset) hasValidDateFilter() bool {
	if a.requestConfig.DateFilter == "" || (a.Bucket == kiosk.SourceMemories || a.Bucket == kiosk.SourceDateRange) {
		return true
	}

	dateStart, dateEnd, err := determineDateRange(a.requestConfig.DateFilter)
	if err != nil {
		log.Error("malformed filter", "err", err)
		return true // Continue processing if date filter is malformed
	}

	return utils.IsTimeBetween(a.LocalDateTime.Local(), dateStart, dateEnd)
}

// hasValidAlbums checks if the asset belongs to any excluded albums.
// If album data is missing, it fetches it first.
//
// Parameters:
//   - requestID: Unique identifier for the request
//   - deviceID: ID of the device making the request
//
// Returns:
//   - bool: true if asset is not in any excluded albums, false otherwise
func (a *Asset) hasValidAlbums(requestID, deviceID string) bool {
	if len(a.AppearsIn) == 0 {
		a.AlbumsThatContainAsset(requestID, deviceID)
	}

	return !slices.ContainsFunc(a.AppearsIn, func(album Album) bool {
		return slices.Contains(a.requestConfig.ExcludedAlbums, album.ID)
	})
}

// hasValidPeople checks if the asset contains any excluded people.
// If people data is missing and exclusions are configured, it fetches face data first.
//
// Parameters:
//   - requestID: Unique identifier for the request
//   - deviceID: ID of the device making the request
//
// Returns:
//   - bool: true if asset contains no excluded people, false otherwise
func (a *Asset) hasValidPeople(requestID, deviceID string) bool {
	if len(a.requestConfig.ExcludedPeople) > 0 && len(a.People) == 0 {
		a.CheckForFaces(requestID, deviceID)
	}

	return !slices.ContainsFunc(a.People, func(person Person) bool {
		return slices.Contains(a.requestConfig.ExcludedPeople, person.ID)
	})
}

// hasValidTags checks if the asset has any tags that would exclude it from processing.
// It first fetches additional asset metadata via AssetInfo if needed. After getting
// the metadata, it restores the asset's orientation ratio since AssetInfo can override
// those values. Finally it checks if the asset has the special "skip" tag that
// indicates it should be excluded.
//
// Parameters:
//   - requestID: Unique identifier for the request, used for logging and caching
//   - deviceID: ID of the device making the request, used for caching
//
// Returns:
//   - bool: true if asset has no excluding tags (like "skip"), false if it should be excluded
func (a *Asset) hasValidTags(requestID, deviceID string) bool {

	if err := a.AssetInfo(requestID, deviceID); err != nil {
		log.Error("Failed to get additional asset data", "error", err)
	}

	// AssetInfo overrides IsPortrait and IsLandscape so lets add them back
	a.AddRatio()

	return !a.containsTag(kiosk.TagSkip)
}

func (a *Asset) fetchPaginatedMetadata(u *url.URL, requestBody SearchRandomBody, requestID string, deviceID string) (int, error) {
	var totalCount int

	for {

		if requestBody.Page > MaxPages {
			log.Warn(requestID + " Reached maximum page count when fetching Metadata")
			break
		}

		var response SearchMetadataResponse

		// convert body to queries so url is unique and can be cached
		queries, _ := query.Values(requestBody)

		apiURL := url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     "api/search/metadata",
			RawQuery: queries.Encode(),
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			_, _, err = immichAPIFail(totalCount, err, nil, apiURL.String())
			return totalCount, err
		}

		immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, response)
		apiBody, err := immichAPICall(a.ctx, http.MethodPost, apiURL.String(), jsonBody)
		if err != nil {
			_, _, err = immichAPIFail(response, err, apiBody, apiURL.String())
			return totalCount, err
		}

		err = json.Unmarshal(apiBody, &response)
		if err != nil {
			_, _, err = immichAPIFail(response, err, apiBody, apiURL.String())
			return totalCount, err
		}

		totalCount += response.Assets.Total

		if response.Assets.NextPage == "" {
			break
		}

		requestBody.Page++
	}

	return totalCount, nil
}

func (a *Asset) updateAsset(deviceID string, requestBody UpdateAssetBody) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var res Asset

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(res, err, nil, "")
		return err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", a.ID),
	}

	jsonBody, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return fmt.Errorf("marshaling request body: %w", marshalErr)
	}

	apiBody, err := a.immichAPICall(a.ctx, http.MethodPut, apiURL.String(), jsonBody)
	if err != nil {
		_, _, err = immichAPIFail(res, err, apiBody, apiURL.String())
		return err
	}

	err = json.Unmarshal(apiBody, &res)
	if err != nil {
		_, _, err = immichAPIFail(res, err, apiBody, apiURL.String())
		return err
	}

	// remove asset data from cache as we've changed its data
	cacheErr := a.RemoveAssetCache(deviceID)
	if cacheErr != nil {
		log.Error("error removing asset from cache", "assetID", a.ID, "error", cacheErr)
	}

	mergErr := a.mergeAssetInfo(res)
	if mergErr != nil {
		log.Error("error merging asset info", "assetID", a.ID, "error", mergErr)
	}

	return nil
}
