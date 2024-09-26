// Package immich provides functions to interact with the Immich API.
//
// It includes functionality for retrieving random images, fetching images
// associated with specific people or albums, and getting image statistics.
// The package also implements caching mechanisms to optimize API calls.
package immich

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"math/rand/v2"

	"github.com/charmbracelet/log"
	"github.com/patrickmn/go-cache"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/utils"
)

const (
	AllAlbumsID    = "all"
	SharedAlbumsID = "shared"
)

var (
	// requestConfig the config for this request
	requestConfig config.Config
	// apiCache cache store for immich api call(s)
	apiCache *cache.Cache
	// apiCacheLock is used to synchronize access to the apiCache
	apiCacheLock sync.Mutex
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

type ImmichBuckets []struct {
	Count      int       `json:"count"`
	TimeBucket time.Time `json:"timeBucket"`
}

type ImmichAlbum struct {
	ID         string        `json:"id"`
	Assets     []ImmichAsset `json:"assets"`
	AssetCount int           `json:"assetCount"`
}

type ImmichAlbums []ImmichAlbum

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
	ImmichAsset | []ImmichAsset | ImmichAlbum | ImmichAlbums | ImmichPersonStatistics | int
}

// immichApiFail handles failures in Immich API calls by unmarshaling the error response,
// logging the error, and returning a formatted error along with the original value.
func immichApiFail[T ImmichApiResponse](value T, err error, body []byte, apiUrl string) (T, error) {
	var immichError ImmichError
	errorUnmarshalErr := json.Unmarshal(body, &immichError)
	if errorUnmarshalErr != nil {
		log.Error("Couldn't ready error", "body", string(body), "url", apiUrl)
		return value, fmt.Errorf(`
			No data or error returned from Immich API.
			<ul>
				<li>Are your data source ID's correct (albumID, personID)?</li>
				<li>Do those data sources have assets?</li>
				<li>Is Immich online?</li>
			</ul>
			<p>
				Full error:<br/><br/>
				<code>%w</code>
			</p>`, err)
	}
	log.Errorf("%s : %v", immichError.Error, immichError.Message)
	return value, fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
}

// immichApiCallDecorator Decorator to impliment cache for the immichApiCall func
func immichApiCallDecorator[T ImmichApiResponse](immichApiCall ImmichApiCall, requestID string, jsonShape T) ImmichApiCall {
	return func(apiUrl string) ([]byte, error) {

		if !requestConfig.Kiosk.Cache {
			return immichApiCall(apiUrl)
		}

		apiCacheLock.Lock()
		defer apiCacheLock.Unlock()

		if apiData, found := apiCache.Get(apiUrl); found {
			if requestConfig.Kiosk.DebugVerbose {
				log.Debug(requestID+" Cache hit", "url", apiUrl)
			}
			return apiData.([]byte), nil
		}

		if requestConfig.Kiosk.DebugVerbose {
			log.Debug(requestID+" Cache miss", "url", apiUrl)
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
			log.Debug(requestID+" Cache saved", "url", apiUrl)
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

func (i *ImmichAsset) people(requestID string, shared bool) (ImmichAlbums, error) {
	var albums ImmichAlbums

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people",
	}

	if shared {
		apiUrl.RawQuery = "shared=true"
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, albums)
	body, err := immichApiCall(apiUrl.String())
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &albums)
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	return albums, nil
}

// personAssets retrieves all assets associated with a specific person from Immich.
func (i *ImmichAsset) personAssets(personID, requestID string) ([]ImmichAsset, error) {

	var images []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personID + "/assets",
	}

	immichApiCal := immichApiCallDecorator(i.immichApiCall, requestID, images)
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

// albums retrieves albums from Immich based on the shared parameter.
// It constructs the API URL, makes the API call, and returns the albums.
func (i *ImmichAsset) albums(requestID string, shared bool) (ImmichAlbums, error) {
	var albums ImmichAlbums

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/albums/",
	}

	if shared {
		apiUrl.RawQuery = "shared=true"
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, albums)
	body, err := immichApiCall(apiUrl.String())
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &albums)
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	return albums, nil
}

// allSharedAlbums retrieves all shared albums from Immich.
func (i *ImmichAsset) allSharedAlbums(requestID string) (ImmichAlbums, error) {
	return i.albums(requestID, true)
}

// allAlbums retrieves all non-shared albums from Immich.
func (i *ImmichAsset) allAlbums(requestID string) (ImmichAlbums, error) {
	return i.albums(requestID, false)
}

// albumAssets retrieves all assets associated with a specific album from Immich.
func (i *ImmichAsset) albumAssets(albumID, requestID string) (ImmichAlbum, error) {
	var album ImmichAlbum

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/albums/" + albumID,
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, album)
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
func (i *ImmichAsset) PersonImageCount(personID, requestID string) (int, error) {

	var personStatistics ImmichPersonStatistics

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personID + "/statistics",
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, personStatistics)
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
func (i *ImmichAsset) RandomImage(requestID, kioskDeviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", kioskDeviceID, "Getting Random image", true)
	} else {
		log.Debug(requestID + " Getting Random image")
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
		RawQuery: "count=1000",
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, immichAssets)
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
		log.Debug(requestID + " No images left in cache. Refreshing and trying again")
		apiCache.Delete(apiUrl.String())
		return i.RandomImage(requestID, kioskDeviceID, isPrefetch)
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

	log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
	apiCache.Delete(apiUrl.String())
	return i.RandomImage(requestID, kioskDeviceID, isPrefetch)
}

// RandomImageOfPerson retrieve random image of person from Immich
func (i *ImmichAsset) RandomImageOfPerson(personID, requestID, kioskDeviceID string, isPrefetch bool) error {

	images, err := i.personAssets(personID, requestID)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		log.Error("no images found", "for person", personID)
		return fmt.Errorf("no images found for person %s", personID)
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
		log.Error("no images found", "for person", personID)
		return fmt.Errorf("no images found for person %s", personID)
	}

	if log.GetLevel() == log.DebugLevel {
		for _, per := range i.People {
			if per.ID == personID {

				if isPrefetch {
					log.Debug(requestID, "PREFETCH", kioskDeviceID, "Got image of person", per.Name)
				} else {
					log.Debug(requestID, "Got image of person", per.Name)
				}

				break
			}
		}
	}

	return nil
}

func (i *ImmichAsset) RandomAlbumFromAllAlbums(requestID string) (string, error) {
	albums, err := i.allAlbums(requestID)
	if err != nil {
		return "", err
	}

	albumsWithWeighting := []utils.AssetWithWeighting{}

	for _, album := range albums {
		albumsWithWeighting = append(albumsWithWeighting, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: "ALBUM", ID: album.ID},
			Weight: album.AssetCount,
		})
	}

	pickedAlbum := utils.PickRandomImageType(requestConfig.Kiosk.AssetWeighting, albumsWithWeighting)

	return pickedAlbum.ID, nil
}

func (i *ImmichAsset) RandomAlbumFromSharedAlbums(requestID string) (string, error) {
	albums, err := i.allSharedAlbums(requestID)
	if err != nil {
		return "", err
	}

	albumsWithWeighting := []utils.AssetWithWeighting{}

	for _, album := range albums {
		albumsWithWeighting = append(albumsWithWeighting, utils.AssetWithWeighting{
			Asset:  utils.WeightedAsset{Type: "ALBUM", ID: album.ID},
			Weight: album.AssetCount,
		})
	}

	pickedAlbum := utils.PickRandomImageType(requestConfig.Kiosk.AssetWeighting, albumsWithWeighting)

	return pickedAlbum.ID, nil
}

// RandomImageFromAlbum retrieve random image within a specified album from Immich
func (i *ImmichAsset) RandomImageFromAlbum(albumID, requestID, kioskDeviceID string, isPrefetch bool) error {
	album, err := i.albumAssets(albumID, requestID)
	if err != nil {
		return err
	}

	if len(album.Assets) == 0 {
		log.Error("no images found", "for album", albumID)
		return fmt.Errorf("no images found for album %s", albumID)
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
		log.Error("no images found", "for album", albumID)
		return fmt.Errorf("no images found for album %s", albumID)
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

func (i *ImmichAsset) countAssetsInAlbums(albums ImmichAlbums) int {
	total := 0
	for _, album := range albums {
		total += album.AssetCount
	}
	return total
}

func (i *ImmichAsset) AlbumImageCount(albumID, requestID string) (int, error) {
	switch albumID {
	case AllAlbumsID:
		albums, err := i.allAlbums(requestID)
		if err != nil {
			return 0, fmt.Errorf("failed to get all albums: %w", err)
		}
		return i.countAssetsInAlbums(albums), nil
	case SharedAlbumsID:
		albums, err := i.allSharedAlbums(requestID)
		if err != nil {
			return 0, fmt.Errorf("failed to get shared albums: %w", err)
		}
		return i.countAssetsInAlbums(albums), nil
	default:
		album, err := i.albumAssets(albumID, requestID)
		if err != nil {
			return 0, fmt.Errorf("failed to get album assets for album %s: %w", albumID, err)
		}
		return len(album.Assets), nil
	}
}
