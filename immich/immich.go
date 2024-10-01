// Package immich provides functions to interact with the Immich API.
//
// It includes functionality for retrieving random images, fetching images
// associated with specific people or albums, and getting image statistics.
// The package also implements caching mechanisms to optimize API calls.
package immich

import (
	"net/url"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/patrickmn/go-cache"

	"github.com/damongolding/immich-kiosk/config"
)

const (
	Portrait       = "PORTRAIT"
	Landscape      = "LANDSCAPE"
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
	ExifImageWidth   int       `json:"exifImageWidth"`
	ExifImageHeight  int       `json:"exifImageHeight"`
	FileSizeInByte   int       `json:"-"` // `json:"fileSizeInByte"`
	Orientation      string    `json:"orientation"`
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
	Ratio            string
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
	RatioWanted      string
	IsPortrait       bool
	IsLandscape      bool
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
