// Package immich provides functions to interact with the Immich API.
//
// It includes functionality for retrieving random images, fetching images
// associated with specific people or albums, and getting image statistics.
// The package also implements caching mechanisms to optimize API calls.
package immich

import (
	"io"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/damongolding/immich-kiosk/config"
)

type ImageOrientation string
type ImmichAssetType string
type ImmichAlbumKeyword string

const (
	PortraitOrientation  ImageOrientation = "PORTRAIT"
	LandscapeOrientation ImageOrientation = "LANDSCAPE"
	SquareOrientation    ImageOrientation = "SQUARE"

	ImageType ImmichAssetType = "IMAGE"
	VideoType ImmichAssetType = "VIDEO"
	AudioType ImmichAssetType = "AUDIO"
	OtherType ImmichAssetType = "OTHER"

	AllAlbumsID    ImmichAlbumKeyword = "all"
	SharedAlbumsID ImmichAlbumKeyword = "shared"
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
	ImageOrientation ImageOrientation
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
	ID               string          `json:"id"`
	DeviceAssetID    string          `json:"-"` // `json:"deviceAssetId"`
	OwnerID          string          `json:"-"` // `json:"ownerId"`
	DeviceID         string          `json:"-"` // `json:"deviceId"`
	LibraryID        string          `json:"-"` // `json:"libraryId"`
	Type             ImmichAssetType `json:"type"`
	OriginalPath     string          `json:"-"`                // `json:"originalPath"`
	OriginalFileName string          `json:"-"`                // `json:"originalFileName"`
	OriginalMimeType string          `json:"originalMimeType"` // `json:"originalMimeType"`
	Resized          bool            `json:"-"`                // `json:"resized"`
	Thumbhash        string          `json:"-"`                // `json:"thumbhash"`
	FileCreatedAt    time.Time       `json:"-"`                // `json:"fileCreatedAt"`
	FileModifiedAt   time.Time       `json:"-"`                // `json:"fileModifiedAt"`
	LocalDateTime    time.Time       `json:"localDateTime"`    // `json:"localDateTime"`
	UpdatedAt        time.Time       `json:"-"`                // `json:"updatedAt"`
	IsFavorite       bool            `json:"isFavorite"`
	IsArchived       bool            `json:"isArchived"`
	IsTrashed        bool            `json:"isTrashed"`
	Duration         string          `json:"-"` // `json:"duration"`
	ExifInfo         ExifInfo        `json:"exifInfo"`
	LivePhotoVideoID any             `json:"-"`        // `json:"livePhotoVideoId"`
	People           People          `json:"people"`   // `json:"people"`
	Checksum         string          `json:"checksum"` // `json:"checksum"`
	StackCount       any             `json:"-"`        // `json:"stackCount"`
	IsOffline        bool            `json:"-"`        // `json:"isOffline"`
	HasMetadata      bool            `json:"-"`        // `json:"hasMetadata"`
	DuplicateID      any             `json:"-"`        // `json:"duplicateId"`
	RatioWanted      ImageOrientation
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

type ImmichSearchBody struct {
	City          string   `url:"city,omitempty" json:"city,omitempty"`
	Country       string   `url:"country,omitempty" json:"country,omitempty"`
	CreatedAfter  string   `url:"createdAfter,omitempty" json:"createdAfter,omitempty"`
	CreatedBefore string   `url:"createdBefore,omitempty" json:"createdBefore,omitempty"`
	DeviceID      string   `url:"deviceId,omitempty" json:"deviceId,omitempty"`
	IsArchived    bool     `url:"isArchived,omitempty" json:"isArchived,omitempty"`
	IsEncoded     bool     `url:"isEncoded,omitempty" json:"isEncoded,omitempty"`
	IsFavorite    bool     `url:"isFavorite,omitempty" json:"isFavorite,omitempty"`
	IsMotion      bool     `url:"isMotion,omitempty" json:"isMotion,omitempty"`
	IsNotInAlbum  bool     `url:"isNotInAlbum,omitempty" json:"isNotInAlbum,omitempty"`
	IsOffline     bool     `url:"isOffline,omitempty" json:"isOffline,omitempty"`
	IsVisible     bool     `url:"isVisible,omitempty" json:"isVisible,omitempty"`
	LensModel     string   `url:"lensModel,omitempty" json:"lensModel,omitempty"`
	LibraryID     string   `url:"libraryId,omitempty" json:"libraryId,omitempty"`
	Make          string   `url:"make,omitempty" json:"make,omitempty"`
	Model         string   `url:"model,omitempty" json:"model,omitempty"`
	PersonIds     []string `url:"personIds,omitempty" json:"personIds,omitempty"`
	Size          int      `url:"size,omitempty" json:"size,omitempty"`
	State         string   `url:"state,omitempty" json:"state,omitempty"`
	TakenAfter    string   `url:"takenAfter,omitempty" json:"takenAfter,omitempty"`
	TakenBefore   string   `url:"takenBefore,omitempty" json:"takenBefore,omitempty"`
	TrashedAfter  string   `url:"trashedAfter,omitempty" json:"trashedAfter,omitempty"`
	TrashedBefore string   `url:"trashedBefore,omitempty" json:"trashedBefore,omitempty"`
	Type          string   `url:"type,omitempty" json:"type,omitempty"`
	UpdatedAfter  string   `url:"updatedAfter,omitempty" json:"updatedAfter,omitempty"`
	UpdatedBefore string   `url:"updatedBefore,omitempty" json:"updatedBefore,omitempty"`
	WithArchived  bool     `url:"withArchived,omitempty" json:"withArchived,omitempty"`
	WithDeleted   bool     `url:"withDeleted,omitempty" json:"withDeleted,omitempty"`
	WithExif      bool     `url:"withExif,omitempty" json:"withExif,omitempty"`
	WithPeople    bool     `url:"withPeople,omitempty" json:"withPeople,omitempty"`
	WithStacked   bool     `url:"withStacked,omitempty" json:"withStacked,omitempty"`
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

type ImmichApiCall func(string, string, io.Reader) ([]byte, error)

type ImmichApiResponse interface {
	ImmichAsset | []ImmichAsset | ImmichAlbum | ImmichAlbums | ImmichPersonStatistics | int
}

func FluchApiCache() {
	apiCache.Flush()
}

func ApiCacheCount() int {
	return apiCache.ItemCount()
}
