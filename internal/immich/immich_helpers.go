package immich

import (
	"bytes"
	"encoding/json"
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
	"github.com/damongolding/immich-kiosk/internal/kiosk"
)

// immichApiFail handles failures in Immich API calls by unmarshaling the error response,
// logging the error, and returning a formatted error along with the original value.
func immichApiFail[T ImmichApiResponse](value T, err error, body []byte, apiUrl string) (T, string, error) {
	var immichError ImmichError
	errorUnmarshalErr := json.Unmarshal(body, &immichError)
	if errorUnmarshalErr != nil {
		log.Error("Couldn't read error", "body", string(body), "url", apiUrl)
		return value, apiUrl, err
	}
	log.Errorf("%s : %v", immichError.Error, immichError.Message)
	return value, apiUrl, fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
}

// withImmichApiCache Decorator to implement cache for the immichApiCall func
func withImmichApiCache[T ImmichApiResponse](immichApiCall ImmichApiCall, requestID, deviceID string, jsonShape T) ImmichApiCall {
	return func(method, apiUrl string, body []byte, headers ...map[string]string) ([]byte, error) {

		if !requestConfig.Kiosk.Cache {
			return immichApiCall(method, apiUrl, body, headers...)
		}

		apiCacheKey := cache.ApiCacheKey(apiUrl, deviceID, requestConfig.SelectedUser)

		if apiData, found := cache.Get(apiCacheKey); found {
			log.Debug(requestID+" Cache hit", "url", apiUrl)
			return apiData.([]byte), nil
		}

		if requestConfig.Kiosk.DebugVerbose {
			log.Debug(requestID+" Cache miss", "url", apiUrl)
		}

		apiBody, err := immichApiCall(method, apiUrl, body)
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
			log.Debug(requestID+" Cache saved", "url", apiUrl)
		}

		return jsonBytes, nil
	}
}

// immichApiCall bootstrap for immich api call
func (i *ImmichAsset) immichApiCall(method, apiUrl string, body []byte, headers ...map[string]string) ([]byte, error) {

	var responseBody []byte
	var lastErr error

	_, err := url.Parse(apiUrl)
	if err != nil {
		log.Error("Invalid URL", "url", apiUrl, "err", err)
		return responseBody, err
	}

	for attempts := range 3 {

		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequest(method, apiUrl, bodyReader)
		if err != nil {
			log.Error(err)
			return responseBody, err
		}

		req.Header.Set("Accept", "application/json")
		apiKey := requestConfig.ImmichApiKey
		if requestConfig.SelectedUser != "" {
			if key, ok := requestConfig.ImmichUsersApiKeys[requestConfig.SelectedUser]; ok {
				apiKey = key
			} else {
				return responseBody, fmt.Errorf("no API key found for user %s in the config", requestConfig.SelectedUser)
			}
		}

		req.Header.Set("x-api-key", apiKey)

		if method == "POST" || method == "PUT" || method == "PATCH" {
			req.Header.Set("Content-Type", "application/json")
		}

		// Add any additional headers
		for _, header := range headers {
			for key, value := range header {
				req.Header.Set(key, value)
			}
		}

		res, err := httpClient.Do(req)
		if err != nil {
			lastErr = err

			// Type assert to get more details about the error
			if urlErr, ok := err.(*url.Error); ok {
				log.Error("Request failed",
					"attempt", attempts,
					"URL", apiUrl,
					"operation", urlErr.Op,
					"error_type", fmt.Sprintf("%T", urlErr.Err),
					"error", urlErr.Err)
			} else {
				log.Error("Request failed",
					"attempt", attempts,
					"URL", apiUrl,
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

			if res.StatusCode == 401 {
				err = fmt.Errorf("received 401 (unauthorised) code from Immich. Please check your Immich API is correct")
			}

			return responseBody, err
		}

		responseBody, err = io.ReadAll(res.Body)
		if err != nil {
			log.Error("reading response body", "url", apiUrl, "err", err)
			return responseBody, err
		}

		return responseBody, nil
	}

	return responseBody, fmt.Errorf("Request failed: max retries exceeded. last err=%v", lastErr)
}

// ratioCheck checks if an image's orientation matches a desired ratio.
// First, it computes the image's ratio (portrait/landscape) by calling addRatio().
// Then it checks if the image matches any desired ratio requirements:
// - If no specific ratio is required (RatioWanted is empty), returns true
// - If RatioWanted is "portrait", returns true only if image is portrait
// - If RatioWanted is "landscape", returns true only if image is landscape
// - Otherwise returns false if orientations don't match
func (i *ImmichAsset) ratioCheck(wantedRatio ImageOrientation) bool {

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
func (i *ImmichAsset) addRatio() {

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
func (i *ImmichAsset) mergeAssetInfo(additionalInfo ImmichAsset) error {
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
func (i *ImmichAsset) AssetInfo(requestID, deviceID string) error {

	var immichAsset ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		return err
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", i.ID),
	}

	immichApiCall := withImmichApiCache(i.immichApiCall, requestID, deviceID, immichAsset)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		_, _, err = immichApiFail(immichAsset, err, body, apiUrl.String())
		return fmt.Errorf("fetching asset info: err %v", err)
	}

	err = json.Unmarshal(body, &immichAsset)
	if err != nil {
		_, _, err = immichApiFail(immichAsset, err, body, apiUrl.String())
		return fmt.Errorf("fetching asset info: err %v", err)
	}

	return i.mergeAssetInfo(immichAsset)
}

// ImagePreview fetches the raw image data from Immich
func (i *ImmichAsset) ImagePreview() ([]byte, error) {

	var bytes []byte

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Error(err)
		return bytes, err
	}

	assetSize := AssetSizeThumbnail
	if requestConfig.UseOriginalImage && slices.Contains(supportedImageMimeTypes, i.OriginalMimeType) {
		assetSize = AssetSizeOriginal
	}

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     path.Join("api", "assets", i.ID, assetSize),
		RawQuery: "size=preview",
	}

	return i.immichApiCall("GET", apiUrl.String(), nil)
}

// FacesCenterPoint calculates the center point of all detected faces in an image as percentages.
// It analyzes both assigned (People) and unassigned faces, finding the bounding box that encompasses
// all faces and returning its center as x,y percentages relative to the image dimensions.
// Returns (0,0) if no faces are detected or if image dimensions are invalid.
func (i *ImmichAsset) FacesCenterPoint() (float64, float64) {
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
			} else {
				minX = min(minX, face.BoundingBoxX1)
				minY = min(minY, face.BoundingBoxY1)
				maxX = max(maxX, face.BoundingBoxX2)
				maxY = max(maxY, face.BoundingBoxY2)
			}
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
		} else {
			minX = min(minX, face.BoundingBoxX1)
			minY = min(minY, face.BoundingBoxY1)
			maxX = max(maxX, face.BoundingBoxX2)
			maxY = max(maxY, face.BoundingBoxY2)
		}
	}

	if !initialized {
		return 0, 0
	}

	centerX := float64(minX+maxX) / 2
	centerY := float64(minY+maxY) / 2

	var percentX, percentY float64
	var imageWidth, imageHeight int

	if len(i.People) != 0 && len(i.People[0].Faces) != 0 {
		imageWidth = i.People[0].Faces[0].ImageWidth
		imageHeight = i.People[0].Faces[0].ImageHeight
	} else if len(i.UnassignedFaces) != 0 {
		imageWidth = i.UnassignedFaces[0].ImageWidth
		imageHeight = i.UnassignedFaces[0].ImageHeight
	} else {
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
func (i *ImmichAsset) FacesCenterPointPX() (float64, float64) {
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
			} else {
				minX = min(minX, face.BoundingBoxX1)
				minY = min(minY, face.BoundingBoxY1)
				maxX = max(maxX, face.BoundingBoxX2)
				maxY = max(maxY, face.BoundingBoxY2)
			}
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
		} else {
			minX = min(minX, face.BoundingBoxX1)
			minY = min(minY, face.BoundingBoxY1)
			maxX = max(maxX, face.BoundingBoxX2)
			maxY = max(maxY, face.BoundingBoxY2)
		}
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
func (i *ImmichAsset) containsTag(tagName string) bool {
	for _, tag := range i.Tags {
		if strings.EqualFold(tag.Name, tagName) {
			return true
		}
	}
	return false
}

// isValidAsset checks if an asset meets all the required criteria for processing.
// It validates:
// - Asset type against allowed types
// - Trash and archive status
// - Image ratio requirements
// - Blacklist status
// - Album exclusions
// - Skip tags
//
// Parameters:
//
//	requestID: unique identifier for the current request
//	deviceID: identifier for the device making the request
//	allowedTypes: slice of allowed asset types
//	wantedRatio: required image orientation
//
// Returns:
//
//	bool: true if the asset is valid, false otherwise
func (i *ImmichAsset) isValidAsset(requestID, deviceID string, allowedTypes []ImmichAssetType, wantedRatio ImageOrientation) bool {
	if !slices.Contains(allowedTypes, i.Type) {
		return false
	}

	if i.IsTrashed {
		return false
	}

	if i.IsArchived && !requestConfig.ShowArchived {
		return false
	}

	if !i.ratioCheck(wantedRatio) {
		return false
	}

	if slices.Contains(requestConfig.Blacklist, i.ID) {
		return false
	}

	// Album validation
	if len(i.AppearsIn) == 0 {
		i.AlbumsThatContainAsset(requestID, deviceID)
	}

	isInExcludedAlbum := slices.ContainsFunc(i.AppearsIn, func(album ImmichAlbum) bool {
		return slices.Contains(requestConfig.ExcludedAlbums, album.ID)
	})
	if isInExcludedAlbum {
		return false
	}

	// Get more info for tag validation
	if err := i.AssetInfo(requestID, deviceID); err != nil {
		log.Error("Failed to get additional asset data", "error", err)
	}

	// Tag validation
	if i.containsTag(kiosk.TagSkip) {
		return false
	}

	return true

}
