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
	// requestConfig the config for this request
	requestConfig config.Config
	// apiCache cache store for immich api call(s)
	apiCache *cache.Cache
)

type ImmichError struct {
	Message    []string `json:"message"`
	Error      string   `json:"error"`
	StatusCode int      `json:"statusCode"`
}

type ExifInfo struct {
	Make             string    `json:"-"` // `json:"make"`
	Model            string    `json:"-"` // `json:"model"`
	ExifImageWidth   int       `json:"-"` // `json:"exifImageWidth"`
	ExifImageHeight  int       `json:"-"` // `json:"exifImageHeight"`
	FileSizeInByte   int       `json:"-"` // `json:"fileSizeInByte"`
	Orientation      any       `json:"-"` // `json:"orientation"`
	DateTimeOriginal time.Time `json:"-"` // `json:"dateTimeOriginal"`
	ModifyDate       time.Time `json:"-"` // `json:"modifyDate"`
	TimeZone         string    `json:"-"` // `json:"timeZone"`
	LensModel        string    `json:"-"` // `json:"lensModel"`
	FNumber          float64   `json:"fNumber"`
	FocalLength      float64   `json:"focalLength"`
	Iso              int       `json:"iso"`
	ExposureTime     string    `json:"-"` // `json:"exposureTime"`
	Latitude         float64   `json:"-"` // `json:"latitude"`
	Longitude        float64   `json:"-"` // `json:"longitude"`
	City             string    `json:"city"`
	State            string    `json:"state"`
	Country          string    `json:"country"`
	Description      string    `json:"-"` // `json:"description"`
	ProjectionType   any       `json:"-"` // `json:"projectionType"`
}

type People []struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	BirthDate     any       `json:"-"` // `json:"birthDate"`
	ThumbnailPath string    `json:"-"` // `json:"thumbnailPath"`
	IsHidden      bool      `json:"-"` // `json:"isHidden"`
	UpdatedAt     time.Time `json:"-"` // `json:"updatedAt"`
	Faces         Faces     `json:"-"` // `json:"faces"`
}

type Faces []struct {
	ID            string `json:"-"` // `json:"id"`
	ImageHeight   int    `json:"-"` // `json:"imageHeight"`
	ImageWidth    int    `json:"-"` // `json:"imageWidth"`
	BoundingBoxX1 int    `json:"-"` // `json:"boundingBoxX1"`
	BoundingBoxX2 int    `json:"-"` // `json:"boundingBoxX2"`
	BoundingBoxY1 int    `json:"-"` // `json:"boundingBoxY1"`
	BoundingBoxY2 int    `json:"-"` // `json:"boundingBoxY2"`
}

type ImmichAsset struct {
	Retries          int
	ID               string    `json:"id"`
	DeviceAssetID    string    `json:"-"` // `json:"deviceAssetId"`
	OwnerID          string    `json:"-"` // `json:"ownerId"`
	DeviceID         string    `json:"-"` // `json:"deviceId"`
	LibraryID        string    `json:"-"` // `json:"libraryId"`
	Type             string    `json:"type"`
	OriginalPath     string    `json:"-"`                // `json:"originalPath"`
	OriginalFileName string    `json:"-"`                // `json:"originalFileName"`
	OriginalMimeType string    `json:"originalMimeType"` // `json:"originalMimeType"`
	Resized          bool      `json:"-"`                // `json:"resized"`
	Thumbhash        string    `json:"-"`                // `json:"thumbhash"`
	FileCreatedAt    time.Time `json:"-"`                // `json:"fileCreatedAt"`
	FileModifiedAt   time.Time `json:"-"`                // `json:"fileModifiedAt"`
	LocalDateTime    time.Time `json:"localDateTime"`    // `json:"localDateTime"`
	UpdatedAt        time.Time `json:"-"`                // `json:"updatedAt"`
	IsFavorite       bool      `json:"isFavorite"`       // `json:"isFavorite"`
	IsArchived       bool      `json:"isArchived"`       // `json:"isArchived"`
	IsTrashed        bool      `json:"isTrashed"`        // `json:"isTrashed"`
	Duration         string    `json:"-"`                // `json:"duration"`
	ExifInfo         ExifInfo  `json:"exifInfo"`
	LivePhotoVideoID any       `json:"-"`        // `json:"livePhotoVideoId"`
	People           People    `json:"people"`   // `json:"people"`
	Checksum         string    `json:"checksum"` // `json:"checksum"`
	StackCount       any       `json:"-"`        // `json:"stackCount"`
	IsOffline        bool      `json:"-"`        // `json:"isOffline"`
	HasMetadata      bool      `json:"-"`        // `json:"hasMetadata"`
	DuplicateID      any       `json:"-"`        // `json:"duplicateId"`
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
	requestConfig = base
	return ImmichAsset{}
}

type ImmichApiCall func(string) ([]byte, error)

// immichApiCallDecorator Decorator to impliment cache for the immichApiCall func
func immichApiCallDecorator[T []ImmichAsset | ImmichAlbum](immichApiCall ImmichApiCall, requestId string, jsonShape T) ImmichApiCall {
	return func(apiUrl string) ([]byte, error) {

		if !requestConfig.Kiosk.Cache {
			return immichApiCall(apiUrl)
		}

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
		log.Debug(requestId+" Cache saved", "url", apiUrl)

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

// GetRandomImage retrieve a random image from Immich
func (i *ImmichAsset) GetRandomImage(requestId string) error {

	log.Debug(requestId + " Getting Random image")

	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
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
			continue
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

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personId + "/assets",
	}

	immichApiCal := immichApiCallDecorator(i.immichApiCall, requestId, images)
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

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/albums/" + albumId,
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestId, album)
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

	return i.immichApiCall(apiUrl.String())
}
