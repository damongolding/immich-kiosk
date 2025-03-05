package immich

import (
	"bytes"
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
	return func(method, apiURL string, body []byte, headers ...map[string]string) ([]byte, error) {

		if !requestConfig.Kiosk.Cache {
			return immichAPICall(method, apiURL, body, headers...)
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

		apiBody, err := immichAPICall(method, apiURL, body)
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
func (i *Asset) immichAPICall(method, apiURL string, body []byte, headers ...map[string]string) ([]byte, error) {

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

		req, err := http.NewRequest(method, apiURL, bodyReader)
		if err != nil {
			log.Error(err)
			return responseBody, err
		}

		req.Header.Set("Accept", "application/json")
		apiKey := i.requestConfig.ImmichAPIKey
		if i.requestConfig.SelectedUser != "" {
			if key, ok := i.requestConfig.ImmichUsersAPIKeys[i.requestConfig.SelectedUser]; ok {
				apiKey = key
			} else {
				return responseBody, fmt.Errorf("no API key found for user %s in the config", i.requestConfig.SelectedUser)
			}
		}

		req.Header.Set("x-api-key", apiKey)

		if method == http.MethodPost || method == "PUT" || method == "PATCH" {
			req.Header.Set("Content-Type", "application/json")
		}

		// Add any additional headers
		for _, header := range headers {
			for key, value := range header {
				req.Header.Set(key, value)
			}
		}

		httpClient.Timeout = time.Second * time.Duration(i.requestConfig.Kiosk.HTTPTimeout)
		res, err := httpClient.Do(req)
		if err != nil {
			lastErr = err

			// Type assert to get more details about the error
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				log.Error("Request failed",
					"attempt", attempts,
					"URL", apiURL,
					"operation", urlErr.Op,
					"error_type", fmt.Sprintf("%T", urlErr.Err),
					"error", urlErr.Err)
			} else {
				log.Error("Request failed",
					"attempt", attempts,
					"URL", apiURL,
					"error_type", fmt.Sprintf("%T", err),
					"error", err)
			}
			time.Sleep(time.Duration(1<<attempts) * time.Second)
			continue
		}

		defer res.Body.Close()

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			err = fmt.Errorf("unexpected status code: %d", res.StatusCode)
			log.Error(err)
			_, _ = io.Copy(io.Discard, res.Body)

			if res.StatusCode == http.StatusUnauthorized {
				err = errors.New("received 401 (unauthorised) code from Immich. Please check your Immich API is correct")
			}

			return responseBody, err
		}

		responseBody, err = io.ReadAll(res.Body)
		if err != nil {
			log.Error("reading response body", "url", apiURL, "err", err)
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
func (i *Asset) ratioCheck(wantedRatio ImageOrientation) bool {

	i.addRatio()

	// specific ratio is not wanted
	if wantedRatio == "" {
		return true
	}

	if (wantedRatio == PortraitOrientation && i.IsPortrait) ||
		(wantedRatio == LandscapeOrientation && i.IsLandscape) {
		return true
	}

	return false
}

// addRatio determines the ratio (portrait or landscape) of the image based on its EXIF information.
// It sets the Ratio field in ExifInfo and updates IsPortrait or IsLandscape accordingly.
// For orientations 5, 6, 7, and 8, it considers the image rotated by 90 degrees.
func (i *Asset) addRatio() {

	switch i.ExifInfo.Orientation {
	case "5", "6", "7", "8":
		// For these orientations, the image is rotated, so we invert the height/width comparison
		if i.ExifInfo.ExifImageHeight < i.ExifInfo.ExifImageWidth {
			i.ExifInfo.ImageOrientation = PortraitOrientation
			i.IsPortrait = true
		} else {
			i.ExifInfo.ImageOrientation = LandscapeOrientation
			i.IsLandscape = true
		}
	default:
		// For all other orientations, including 1, 2, 3, 4, and any unknown orientations
		if i.ExifInfo.ExifImageHeight > i.ExifInfo.ExifImageWidth {
			i.ExifInfo.ImageOrientation = PortraitOrientation
			i.IsPortrait = true
		} else {
			i.ExifInfo.ImageOrientation = LandscapeOrientation
			i.IsLandscape = true
		}
	}
}

// mergeAssetInfo merges additional asset information into the current ImmichAsset.
// It uses reflection to examine each field of the current asset and if a field
// has its zero value, it copies the corresponding value from additionalInfo.
// This allows for selective updating of asset information while preserving
// existing non-zero values.
//
// Parameters:
//   - additionalInfo: The ImmichAsset containing the additional information to merge
//
// Returns an error if any field in additionalInfo is invalid during the merge process.
func (i *Asset) mergeAssetInfo(additionalInfo Asset) error {
	v := reflect.ValueOf(i).Elem()
	d := reflect.ValueOf(additionalInfo)

	for index := range v.NumField() {
		field := v.Field(index)
		if !field.CanSet() {
			continue
		}
		if reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			if !d.Field(index).IsValid() {
				return fmt.Errorf("invalid field at index %d", index)
			}
			field.Set(d.Field(index))
		}
	}
	return nil
}

// AssetInfo fetches the image information from Immich
func (i *Asset) AssetInfo(requestID, deviceID string) error {

	var immichAsset Asset

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", i.ID),
	}

	immichAPICall := withImmichAPICache(i.immichAPICall, requestID, deviceID, i.requestConfig, immichAsset)
	body, err := immichAPICall(http.MethodGet, apiURL.String(), nil)
	if err != nil {
		_, _, err = immichAPIFail(immichAsset, err, body, apiURL.String())
		return fmt.Errorf("fetching asset info: err %w", err)
	}

	err = json.Unmarshal(body, &immichAsset)
	if err != nil {
		_, _, err = immichAPIFail(immichAsset, err, body, apiURL.String())
		return fmt.Errorf("fetching asset info: err %w", err)
	}

	return i.mergeAssetInfo(immichAsset)
}

// ImagePreview fetches the raw image data from Immich
func (i *Asset) ImagePreview() ([]byte, error) {

	var bytes []byte

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		log.Error(err)
		return bytes, err
	}

	assetSize := AssetSizeThumbnail
	if i.requestConfig.UseOriginalImage && slices.Contains(supportedImageMimeTypes, i.OriginalMimeType) {
		assetSize = AssetSizeOriginal
	}

	apiURL := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     path.Join("api", "assets", i.ID, assetSize),
		RawQuery: "size=preview",
	}

	return i.immichAPICall(http.MethodGet, apiURL.String(), nil)
}

// FacesCenterPoint calculates the center point of all detected faces in an image as percentages.
// It analyzes both assigned (People) and unassigned faces, finding the bounding box that encompasses
// all faces and returning its center as x,y percentages relative to the image dimensions.
// Returns (0,0) if no faces are detected or if image dimensions are invalid.
func (i *Asset) FacesCenterPoint() (float64, float64) {
	if len(i.People) == 0 && len(i.UnassignedFaces) == 0 {
		return 0, 0
	}

	var minX, minY, maxX, maxY int
	initialized := false

	for _, person := range i.People {
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

	for _, face := range i.UnassignedFaces {
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
	case len(i.People) != 0 && len(i.People[0].Faces) != 0:
		imageWidth = i.People[0].Faces[0].ImageWidth
		imageHeight = i.People[0].Faces[0].ImageHeight
	case len(i.UnassignedFaces) != 0:
		imageWidth = i.UnassignedFaces[0].ImageWidth
		imageHeight = i.UnassignedFaces[0].ImageHeight
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
func (i *Asset) FacesCenterPointPX() (float64, float64) {
	if len(i.People) == 0 && len(i.UnassignedFaces) == 0 {
		return 0, 0
	}

	var minX, minY, maxX, maxY int
	initialized := false

	for _, person := range i.People {
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

	for _, face := range i.UnassignedFaces {
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
func (i *Asset) containsTag(tagName string) bool {
	for _, tag := range i.Tags {
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
func (i *Asset) isValidAsset(requestID, deviceID string, allowedTypes []AssetType, wantedRatio ImageOrientation) bool {
	return i.hasValidBasicProperties(allowedTypes, wantedRatio) &&
		i.hasValidDateFilter() &&
		i.hasValidAlbums(requestID, deviceID) &&
		i.hasValidPeople(requestID, deviceID) &&
		i.hasValidTags(requestID, deviceID)
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
func (i *Asset) hasValidBasicProperties(allowedTypes []AssetType, wantedRatio ImageOrientation) bool {
	if !slices.Contains(allowedTypes, i.Type) {
		return false
	}
	if i.IsTrashed {
		return false
	}
	if i.IsArchived && !i.requestConfig.ShowArchived {
		return false
	}
	if !i.ratioCheck(wantedRatio) {
		return false
	}
	if slices.Contains(i.requestConfig.Blacklist, i.ID) {
		return false
	}
	return true
}

// hasValidDateFilter validates if the asset's date matches the configured date filter criteria.
// Assets from Memories or DateRange buckets bypass the date filter check.
//
// Returns:
//   - bool: true if date is valid or no filter set, false if outside filter range
func (i *Asset) hasValidDateFilter() bool {
	if i.requestConfig.DateFilter == "" || (i.Bucket == kiosk.SourceMemories || i.Bucket == kiosk.SourceDateRange) {
		return true
	}

	dateStart, dateEnd, err := determineDateRange(i.requestConfig.DateFilter)
	if err != nil {
		log.Error("malformed filter", "err", err)
		return true // Continue processing if date filter is malformed
	}

	return utils.IsTimeBetween(i.LocalDateTime.Local(), dateStart, dateEnd)
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
func (i *Asset) hasValidAlbums(requestID, deviceID string) bool {
	if len(i.AppearsIn) == 0 {
		i.AlbumsThatContainAsset(requestID, deviceID)
	}

	return !slices.ContainsFunc(i.AppearsIn, func(album Album) bool {
		return slices.Contains(i.requestConfig.ExcludedAlbums, album.ID)
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
func (i *Asset) hasValidPeople(requestID, deviceID string) bool {
	if len(i.requestConfig.ExcludedPeople) > 0 && len(i.People) == 0 {
		i.CheckForFaces(requestID, deviceID)
	}

	return !slices.ContainsFunc(i.People, func(person Person) bool {
		return slices.Contains(i.requestConfig.ExcludedPeople, person.ID)
	})
}

// hasValidTags checks if the asset has any tags that would exclude it from processing.
// Fetches asset info if needed and checks for the skip tag.
//
// Parameters:
//   - requestID: Unique identifier for the request
//   - deviceID: ID of the device making the request
//
// Returns:
//   - bool: true if asset has no excluding tags, false otherwise
func (i *Asset) hasValidTags(requestID, deviceID string) bool {
	if err := i.AssetInfo(requestID, deviceID); err != nil {
		log.Error("Failed to get additional asset data", "error", err)
	}

	return !i.containsTag(kiosk.TagSkip)
}

func (i *Asset) fetchPaginatedMetadata(u *url.URL, requestBody interface{}, requestID string, deviceID string) (int, error) {
	var totalCount int
	pageCount := 1

	for {
		var response SearchMetadataResponse

		// Update page number in request body
		reflect.ValueOf(requestBody).Elem().FieldByName("Page").Set(reflect.ValueOf(pageCount))

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

		immichAPICall := withImmichAPICache(i.immichAPICall, requestID, deviceID, i.requestConfig, response)
		apiBody, err := immichAPICall(http.MethodPost, apiURL.String(), jsonBody)
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

		pageCount++
	}

	return totalCount, nil
}
