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

type ImmichPersonStatistics struct {
	Assets int `json:"assets"`
}

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
	DateTimeOriginal time.Time `json:"dateTimeOriginal"`
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
	IsFavorite       bool      `json:"isFavorite"`
	IsArchived       bool      `json:"isArchived"`
	IsTrashed        bool      `json:"isTrashed"`
	Duration         string    `json:"-"` // `json:"duration"`
	ExifInfo         ExifInfo  `json:"exifInfo"`
	LivePhotoVideoID any       `json:"-"`        // `json:"livePhotoVideoId"`
	People           People    `json:"people"`   // `json:"people"`
	Checksum         string    `json:"checksum"` // `json:"checksum"`
	StackCount       any       `json:"-"`        // `json:"stackCount"`
	IsOffline        bool      `json:"-"`        // `json:"isOffline"`
	HasMetadata      bool      `json:"-"`        // `json:"hasMetadata"`
	DuplicateID      any       `json:"-"`        // `json:"duplicateId"`
}

type WeightedAsset struct {
	Type string
	ID   string
}

type AssetWithWeighting struct {
	Asset  WeightedAsset
	Weight int
}

type ImmichBuckets []struct {
	Count      int       `json:"count"`
	TimeBucket time.Time `json:"timeBucket"`
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

type ImmichApiResponse interface {
	ImmichAsset | []ImmichAsset | ImmichAlbum | ImmichPersonStatistics | int
}

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

// personAssets retrieves all assets associated with a specific person from Immich.
func (i *ImmichAsset) personAssets(personId, requestId string) ([]ImmichAsset, error) {

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
		return immichApiFail(images, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &images)
	if err != nil {
		return immichApiFail(images, err, body, apiUrl.String())
	}

	return images, nil
}

// albumAssets retrieves all assets associated with a specific album from Immich.
func (i *ImmichAsset) albumAssets(albumId, requestId string) (ImmichAlbum, error) {
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
		return immichApiFail(album, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &album)
	if err != nil {
		return immichApiFail(album, err, body, apiUrl.String())
	}

	return album, nil
}

// PersonImageCount returns the number of images associated with a specific person in Immich.
func (i *ImmichAsset) PersonImageCount(personId, requestId string) (int, error) {

	var personStatistics ImmichPersonStatistics

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personId + "/statistics",
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestId, personStatistics)
	body, err := immichApiCall(apiUrl.String())
	if err != nil {
		_, err = immichApiFail(personStatistics, err, body, apiUrl.String())
		return 0, err
	}

	err = json.Unmarshal(body, &personStatistics)
	if err != nil {
		_, err = immichApiFail(personStatistics, err, body, apiUrl.String())
		return 0, err
	}

	return personStatistics.Assets, err
}

// GetRandomImage retrieve a random image from Immich
func (i *ImmichAsset) RandomImage(requestId, kioskDeviceId string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestId, "PREFETCH", kioskDeviceId, "Getting Random image", true)
	} else {
		log.Debug(requestId + " Getting Random image")
	}

	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal("parsing url", err)
	}

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "api/assets/random",
		RawQuery: "count=100",
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestId, immichAssets)
	body, err := immichApiCall(apiUrl.String())
	if err != nil {
		_, err = immichApiFail(immichAssets, err, body, apiUrl.String())
		return err
	}

	err = json.Unmarshal(body, &immichAssets)
	if err != nil {
		_, err = immichApiFail(immichAssets, err, body, apiUrl.String())
		return err
	}

	if len(immichAssets) == 0 {
		log.Debug(requestId + " No images left in cache. Refreshing and trying again")
		apiCache.Delete(apiUrl.String())
		return i.RandomImage(requestId, kioskDeviceId, isPrefetch)
	}

	for immichAssetIndex, img := range immichAssets {
		// We only want images and that are not trashed or archived (unless wanted by user)
		if img.Type != "IMAGE" || img.IsTrashed || (img.IsArchived && !requestConfig.ShowArchived) {
			continue
		}

		if requestConfig.Kiosk.Cache {
			// Remove the current image from the slice
			immichAssetsToCache := append(immichAssets[:immichAssetIndex], immichAssets[immichAssetIndex+1:]...)
			jsonBytes, err := json.Marshal(immichAssetsToCache)
			if err != nil {
				log.Error("Failed to marshal immichAssetsToCache", "error", err)
				return err
			}
			// replace cwith cache minus used image
			err = apiCache.Replace(apiUrl.String(), jsonBytes, cache.DefaultExpiration)
			if err != nil {
				log.Debug("cache not found!")
			}
		}

		*i = img
		return nil
	}

	log.Debug(requestId + " No viable images left in cache. Refreshing and trying again")
	apiCache.Delete(apiUrl.String())
	return i.RandomImage(requestId, kioskDeviceId, isPrefetch)
}

// RandomImageOfPerson retrieve random image of person from Immich
func (i *ImmichAsset) RandomImageOfPerson(personId, requestId, kioskDeviceId string, isPrefetch bool) error {

	images, err := i.personAssets(personId, requestId)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	rand.Shuffle(len(images), func(i, j int) {
		images[i], images[j] = images[j], images[i]
	})

	for _, pick := range images {
		// We only want images and that are not trashed or archived (unless wanted by user)
		if pick.Type != "IMAGE" || pick.IsTrashed || (pick.IsArchived && !requestConfig.ShowArchived) {
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

				if isPrefetch {
					log.Debug(requestId, "PREFETCH", kioskDeviceId, "Got image of person", per.Name)
				} else {
					log.Debug(requestId, "Got image of person", per.Name)
				}

				break
			}
		}
	}

	return nil
}

// RandomImageFromAlbum retrieve random image within a specified album from Immich
func (i *ImmichAsset) RandomImageFromAlbum(albumId, requestId, kioskDeviceId string, isPrefetch bool) error {
	album, err := i.albumAssets(albumId, requestId)
	if err != nil {
		return err
	}

	if len(album.Assets) == 0 {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	rand.Shuffle(len(album.Assets), func(i, j int) {
		album.Assets[i], album.Assets[j] = album.Assets[j], album.Assets[i]
	})

	for _, pick := range album.Assets {
		// We only want images and that are not trashed or archived (unless wanted by user)
		if pick.Type != "IMAGE" || pick.IsTrashed || (pick.IsArchived && !requestConfig.ShowArchived) {
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

	return i.immichApiCall(apiUrl.String())
}

func (i *ImmichAsset) AlbumImageCount(albumId, requestId string) (int, error) {
	album, err := i.albumAssets(albumId, requestId)
	return len(album.Assets), err
}
