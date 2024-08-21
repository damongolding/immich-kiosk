package immich

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"math/rand/v2"

	"github.com/charmbracelet/log"
	"github.com/patrickmn/go-cache"

	"github.com/damongolding/immich-kiosk/config"
)

// maxRetries the maximum amount of retries to find a IMAGE type
const maxRetries int = 10

var (
	// baseConfig the base config i.e config.yaml or ENV
	baseConfig config.Config
	// apiCache cache store for immich api call(s)
	apiCache *cache.Cache
)

type ImmichError struct {
	Message    []string `json:"message"`
	Error      string   `json:"error"`
	StatusCode int      `json:"statusCode"`
}

type ImmichAsset struct {
	Retries          int
	ID               string    `json:"id"`
	DeviceAssetID    string    `json:"deviceAssetId"`
	OwnerID          string    `json:"ownerId"`
	DeviceID         string    `json:"deviceId"`
	LibraryID        string    `json:"libraryId"`
	Type             string    `json:"type"`
	OriginalPath     string    `json:"originalPath"`
	OriginalFileName string    `json:"originalFileName"`
	OriginalMimeType string    `json:"originalMimeType"`
	Resized          bool      `json:"resized"`
	Thumbhash        string    `json:"thumbhash"`
	FileCreatedAt    time.Time `json:"fileCreatedAt"`
	FileModifiedAt   time.Time `json:"fileModifiedAt"`
	LocalDateTime    time.Time `json:"localDateTime"`
	UpdatedAt        time.Time `json:"updatedAt"`
	IsFavorite       bool      `json:"isFavorite"`
	IsArchived       bool      `json:"isArchived"`
	IsTrashed        bool      `json:"isTrashed"`
	Duration         string    `json:"duration"`
	ExifInfo         struct {
		Make             string    `json:"make"`
		Model            string    `json:"model"`
		ExifImageWidth   int       `json:"exifImageWidth"`
		ExifImageHeight  int       `json:"exifImageHeight"`
		FileSizeInByte   int       `json:"fileSizeInByte"`
		Orientation      any       `json:"orientation"`
		DateTimeOriginal time.Time `json:"dateTimeOriginal"`
		ModifyDate       time.Time `json:"modifyDate"`
		TimeZone         string    `json:"timeZone"`
		LensModel        string    `json:"lensModel"`
		FNumber          float64   `json:"fNumber"`
		FocalLength      float64   `json:"focalLength"`
		Iso              int       `json:"iso"`
		ExposureTime     string    `json:"exposureTime"`
		Latitude         float64   `json:"latitude"`
		Longitude        float64   `json:"longitude"`
		City             string    `json:"city"`
		State            string    `json:"state"`
		Country          string    `json:"country"`
		Description      string    `json:"description"`
		ProjectionType   any       `json:"projectionType"`
	} `json:"exifInfo"`
	LivePhotoVideoID any `json:"livePhotoVideoId"`
	People           []struct {
		ID            string    `json:"id"`
		Name          string    `json:"name"`
		BirthDate     any       `json:"birthDate"`
		ThumbnailPath string    `json:"thumbnailPath"`
		IsHidden      bool      `json:"isHidden"`
		UpdatedAt     time.Time `json:"updatedAt"`
		Faces         []struct {
			ID            string `json:"id"`
			ImageHeight   int    `json:"imageHeight"`
			ImageWidth    int    `json:"imageWidth"`
			BoundingBoxX1 int    `json:"boundingBoxX1"`
			BoundingBoxX2 int    `json:"boundingBoxX2"`
			BoundingBoxY1 int    `json:"boundingBoxY1"`
			BoundingBoxY2 int    `json:"boundingBoxY2"`
		} `json:"faces"`
	} `json:"people"`
	Checksum    string `json:"checksum"`
	StackCount  any    `json:"stackCount"`
	IsOffline   bool   `json:"isOffline"`
	HasMetadata bool   `json:"hasMetadata"`
	DuplicateID any    `json:"duplicateId"`
}

type ImmichAlbum struct {
	Assets []ImmichAsset `json:"assets"`
}

func init() {
	// Setting up Immich api cache
	apiCache = cache.New(5*time.Minute, 10*time.Minute)
}

// NewImage returns a new image instance
func NewImage(base config.Config) ImmichAsset {
	baseConfig = base
	return ImmichAsset{}
}

type ImmichApiCall func(string) ([]byte, error)

// immichApiCallDecorator Decorator to impliment cache for the immichApiCall func
func immichApiCallDecorator(immichApiCall ImmichApiCall, requestId string) ImmichApiCall {
	return func(apiUrl string) ([]byte, error) {

		if baseConfig.Kiosk.Cache {
			apiData, found := apiCache.Get(apiUrl)
			if found {
				log.Debug(requestId+" Cache hit", "url", apiUrl)
				return apiData.([]byte), nil
			}

			log.Debug(requestId+" Cache miss", "url", apiUrl)
			body, err := immichApiCall(apiUrl)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			apiCache.Set(apiUrl, body, cache.DefaultExpiration)
			log.Debug(requestId+" Cache saved", "url", apiUrl)
			return body, nil

		}

		return immichApiCall(apiUrl)
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
	req.Header.Add("x-api-key", baseConfig.ImmichApiKey)

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

// GetRandomImage retrieve a random image from Immich
func (i *ImmichAsset) GetRandomImage(requestId string) error {

	log.Debug(requestId + " Getting Random image")

	var immichAssets []ImmichAsset

	u, err := url.Parse(baseConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "api/assets/random",
		RawQuery: "count=6",
	}

	body, err := i.immichApiCall(apiUrl.String())
	if err != nil {
		log.Error(err)
		return err
	}

	err = json.Unmarshal(body, &immichAssets)
	if err != nil {
		var immichError ImmichError
		errorUnmarshalErr := json.Unmarshal(body, &immichError)
		if errorUnmarshalErr != nil {
			log.Error("couldn't read error", "body", string(body))
			return err
		}
		return fmt.Errorf("%s : %v", immichError.Error, immichError.Message)

	}

	if len(immichAssets) == 0 {
		log.Error("no assets found")
		return fmt.Errorf("no assets found")
	}

	for _, img := range immichAssets {
		// We only want images and that are not archived or trashed
		if img.Type != "IMAGE" || img.IsArchived || img.IsTrashed {
			return nil
		}

		*i = img
		return nil
	}

	// No images found
	i.Retries++
	log.Debug(requestId+" Not a image. Trying again", "retry", i.Retries)

	if i.Retries >= maxRetries {
		log.Error("No images found")
		return fmt.Errorf("No images found")
	}

	return i.GetRandomImage(requestId)
}

// GetRandomImageOfPerson retrieve random image of person from Immich
func (i *ImmichAsset) GetRandomImageOfPerson(personId, requestId string) error {

	var images []ImmichAsset

	u, err := url.Parse(baseConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personId + "/assets",
	}

	immichApiCal := immichApiCallDecorator(i.immichApiCall, requestId)
	body, err := immichApiCal(apiUrl.String())
	if err != nil {
		log.Error(err)
		return err
	}

	err = json.Unmarshal(body, &images)
	if err != nil {
		var immichError ImmichError
		errorUnmarshalErr := json.Unmarshal(body, &immichError)
		if errorUnmarshalErr != nil {
			log.Error("couln't read error", "body", string(body))
			return err
		}
		log.Errorf("%s : %v", immichError.Error, immichError.Message)
		return fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
	}

	if len(images) == 0 {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	rand.Shuffle(len(images), func(i, j int) {
		images[i], images[j] = images[j], images[i]
	})

	for _, pick := range images {
		// We only want images and that are not archived or trashed
		if pick.Type != "IMAGE" || pick.IsArchived || pick.IsTrashed {
			continue
		}

		*i = pick
		break
	}

	if i.ID == "" {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	if log.GetLevel() == log.DebugLevel {
		for _, per := range i.People {
			if per.ID == personId {
				log.Debug(requestId+" Got image of", "person", per.Name)
				break
			}
		}
	}

	return nil
}

// GetRandomImageFromAlbum retrieve random image within a specified album from Immich
func (i *ImmichAsset) GetRandomImageFromAlbum(albumId, requestId string) error {
	var album ImmichAlbum

	u, err := url.Parse(baseConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/albums/" + albumId,
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestId)
	body, err := immichApiCall(apiUrl.String())
	if err != nil {
		log.Error(err)
		return err
	}

	err = json.Unmarshal(body, &album)
	if err != nil {
		var immichError ImmichError
		errorUnmarshalErr := json.Unmarshal(body, &immichError)
		if errorUnmarshalErr != nil {
			log.Error("couln't read error", "body", string(body))
			return err
		}
		log.Errorf("%s : %v", immichError.Error, immichError.Message)
		return fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
	}

	if len(album.Assets) == 0 {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	rand.Shuffle(len(album.Assets), func(i, j int) {
		album.Assets[i], album.Assets[j] = album.Assets[j], album.Assets[i]
	})

	for _, pick := range album.Assets {
		// We only want images and that should not be archived or in trashed
		if pick.Type != "IMAGE" || pick.IsArchived || pick.IsTrashed {
			continue
		}

		*i = pick
		break
	}

	if i.ID == "" {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	return nil
}

// GetImagePreview fetches the raw image data from Immich
func (i *ImmichAsset) GetImagePreview() ([]byte, error) {

	var bytes []byte

	u, err := url.Parse(baseConfig.ImmichUrl)
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

	return i.immichApiCall(apiUrl.String())
}
