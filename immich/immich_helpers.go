package immich

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/patrickmn/go-cache"
)

// immichApiFail handles failures in Immich API calls by unmarshaling the error response,
// logging the error, and returning a formatted error along with the original value.
func immichApiFail[T ImmichApiResponse](value T, err error, body []byte, apiUrl string) (T, error) {
	var immichError ImmichError
	errorUnmarshalErr := json.Unmarshal(body, &immichError)
	if errorUnmarshalErr != nil {
		log.Error("couln't read error", "body", string(body), "url", apiUrl)
		return value, err
	}
	log.Errorf("%s : %v", immichError.Error, immichError.Message)
	return value, fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
}

// immichApiCallDecorator Decorator to impliment cache for the immichApiCall func
func immichApiCallDecorator[T ImmichApiResponse](immichApiCall ImmichApiCall, requestId string, jsonShape T) ImmichApiCall {
	return func(apiUrl string) ([]byte, error) {

		if !requestConfig.Kiosk.Cache {
			return immichApiCall(apiUrl)
		}

		apiCacheLock.Lock()
		defer apiCacheLock.Unlock()

		if apiData, found := apiCache.Get(apiUrl); found {
			if requestConfig.Kiosk.DebugVerbose {
				log.Debug(requestId+" Cache hit", "url", apiUrl)
			}
			log.Debug(requestId+" Cache hit", "url", apiUrl)
			return apiData.([]byte), nil
		}

		if requestConfig.Kiosk.DebugVerbose {
			log.Debug(requestId+" Cache miss", "url", apiUrl)
		}
		body, err := immichApiCall(apiUrl)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// Unpack api json into struct which discards data we don't use (for smaller cache size)
		err = json.Unmarshal(body, &jsonShape)
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

		apiCache.Set(apiUrl, jsonBytes, cache.DefaultExpiration)
		if requestConfig.Kiosk.DebugVerbose {
			log.Debug(requestId+" Cache saved", "url", apiUrl)
		}

		return jsonBytes, nil
	}
}

// immichApiCall bootstrap for immich api call
func (i *ImmichAsset) immichApiCall(apiUrl string) ([]byte, error) {

	var responseBody []byte

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		log.Error(err)
		return responseBody, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-api-key", requestConfig.ImmichApiKey)

	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return responseBody, err
	}
	defer res.Body.Close()

	responseBody, err = io.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return responseBody, err
	}

	return responseBody, err
}

// ratioCheck checks if the given image matches the desired ratio.
// It first adds the ratio information to the image, then checks if the ratio
// matches the desired ratio (Portrait or Landscape) if specified.
// If no specific ratio is wanted, it returns true.
func (i *ImmichAsset) ratioCheck(img *ImmichAsset) bool {

	img.addRatio()

	// specific ratio is not wanted
	if i.RatioWanted == "" {
		return true
	}

	if (i.RatioWanted == Portrait && img.IsPortrait) ||
		(i.RatioWanted == Landscape && img.IsLandscape) {
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
			i.ExifInfo.Ratio = Portrait
			i.IsPortrait = true
		} else {
			i.ExifInfo.Ratio = Landscape
			i.IsLandscape = true
		}
	default:
		// For all other orientations, including 1, 2, 3, 4, and any unknown orientations
		if i.ExifInfo.ExifImageHeight > i.ExifInfo.ExifImageWidth {
			i.ExifInfo.Ratio = Portrait
			i.IsPortrait = true
		} else {
			i.ExifInfo.Ratio = Landscape
			i.IsLandscape = true
		}
	}
}
