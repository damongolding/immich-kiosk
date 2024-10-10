package immich

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/charmbracelet/log"
	"github.com/patrickmn/go-cache"
)

// immichApiFail handles failures in Immich API calls by unmarshaling the error response,
// logging the error, and returning a formatted error along with the original value.
func immichApiFail[T ImmichApiResponse](value T, err error, body []byte, apiUrl string) (T, error) {
	var immichError ImmichError
	errorUnmarshalErr := json.Unmarshal(body, &immichError)
	if errorUnmarshalErr != nil {
		log.Error("Couldn't read error", "body", string(body), "url", apiUrl)
		return value, err
	}
	log.Errorf("%s : %v", immichError.Error, immichError.Message)
	return value, fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
}

// immichApiCallDecorator Decorator to impliment cache for the immichApiCall func
func immichApiCallDecorator[T ImmichApiResponse](immichApiCall ImmichApiCall, requestID string, jsonShape T) ImmichApiCall {
	return func(method, apiUrl string, body io.Reader) ([]byte, error) {

		if !requestConfig.Kiosk.Cache {
			return immichApiCall(method, apiUrl, body)
		}

		apiCacheLock.Lock()
		defer apiCacheLock.Unlock()

		if apiData, found := apiCache.Get(apiUrl); found {
			if requestConfig.Kiosk.DebugVerbose {
				log.Debug(requestID+" Cache hit", "url", apiUrl)
			}
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

		apiCache.Set(apiUrl, jsonBytes, cache.DefaultExpiration)
		if requestConfig.Kiosk.DebugVerbose {
			log.Debug(requestID+" Cache saved", "url", apiUrl)
		}

		return jsonBytes, nil
	}
}

// immichApiCall bootstrap for immich api call
func (i *ImmichAsset) immichApiCall(method, apiUrl string, body io.Reader) ([]byte, error) {

	var responseBody []byte

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, apiUrl, body)
	if err != nil {
		log.Error(err)
		return responseBody, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-api-key", requestConfig.ImmichApiKey)

	if method == "POST" || method == "PUT" || method == "PATCH" {
		req.Header.Add("Content-Type", "application/json")
	}

	var res *http.Response
	for attempts := 0; attempts < 3; attempts++ {
		res, err = client.Do(req)
		if err == nil {
			break
		}
		log.Error("Request failed, retrying", "attempt", attempts, "URL", apiUrl, "err", err)
		time.Sleep(time.Duration(attempts) * time.Second)
	}
	if err != nil {
		log.Error("Request failed after retries", "err", err)
		return responseBody, err
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		err = fmt.Errorf("unexpected status code: %d", res.StatusCode)
		log.Error(err)
		return responseBody, err
	}

	responseBody, err = io.ReadAll(res.Body)
	if err != nil {
		log.Error("reading response body", "url", apiUrl, "err", err)
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

	if (i.RatioWanted == PortraitOrientation && img.IsPortrait) ||
		(i.RatioWanted == LandscapeOrientation && img.IsLandscape) {
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

// ImagePreview fetches the raw image data from Immich
func (i *ImmichAsset) ImagePreview() ([]byte, error) {

	var bytes []byte

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Error(err)
		return bytes, err
	}

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "/api/assets/" + i.ID + "/thumbnail",
		RawQuery: "size=preview",
	}

	return i.immichApiCall("GET", apiUrl.String(), nil)
}
