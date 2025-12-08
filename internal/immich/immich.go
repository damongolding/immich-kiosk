// Package immich provides functions to interact with the Immich API.
//
// It includes functionality for retrieving random images, fetching images
// associated with specific people or albums, and getting image statistics.
// The package also implements caching mechanisms to optimize API calls.
package immich

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich_open_api"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
)

type ImageOrientation string
type AssetType string
type AssetOrder string

const (
	MaxRetries = 3
	MaxPages   = 100

	PortraitOrientation  ImageOrientation = "PORTRAIT"
	LandscapeOrientation ImageOrientation = "LANDSCAPE"
	SquareOrientation    ImageOrientation = "SQUARE"

	ImageType AssetType = "IMAGE"
	VideoType AssetType = "VIDEO"
	AudioType AssetType = "AUDIO"
	OtherType AssetType = "OTHER"

	AssetSizeThumbnail string = "thumbnail"
	AssetSizeOriginal  string = "original"

	Asc  AssetOrder = "asc"
	Desc AssetOrder = "desc"
	Rand AssetOrder = "rand"
)

var (
	// httpTransport defines the transport layer configuration for HTTP requests to the Immich API.
	// It manages connection pooling, keepalive settings, and connection timeouts.
	httpTransport = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 100,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	// httpClient default http client for Immich api calls
	HTTPClient = &http.Client{
		Transport: httpTransport,
	}

	supportedImageMimeTypes = []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	ImageOnlyAssetTypes = []AssetType{ImageType}
	VideoOnlyAssetTypes = []AssetType{VideoType}
	AllAssetTypes       = []AssetType{ImageType, VideoType}
)

type PersonStatistics struct {
	Assets int `json:"assets"`
}

type Error struct {
	Error      string   `json:"error"`
	Message    []string `json:"message"`
	StatusCode int      `json:"statusCode"`
}

type Owner struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type ExifInfo struct {
	DateTimeOriginal time.Time `json:"dateTimeOriginal"`
	ModifyDate       time.Time `json:"modifyDate"`
	ProjectionType   any       `json:"-"` // `json:"projectionType"`
	Make             string    `json:"make"`
	Model            string    `json:"model"`
	Orientation      string    `json:"orientation"`
	TimeZone         string    `json:"timeZone"`
	LensModel        string    `json:"lensModel"`
	ExposureTime     string    `json:"exposureTime"`
	City             string    `json:"city"`
	State            string    `json:"state"`
	Country          string    `json:"country"`
	Description      string    `json:"description"`
	ImageOrientation ImageOrientation
	ExifImageWidth   int     `json:"exifImageWidth"`
	ExifImageHeight  int     `json:"exifImageHeight"`
	FileSizeInByte   int     `json:"fileSizeInByte"`
	FNumber          float64 `json:"fNumber"`
	FocalLength      float64 `json:"focalLength"`
	Iso              int     `json:"iso"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
}

type BirthDate string

func (bd BirthDate) Time() (time.Time, error) {
	if string(bd) == "" {
		return time.Time{}, errors.New("empty birth date")
	}
	return time.Parse("2006-01-02", string(bd))
}

type Person struct {
	UpdatedAt     time.Time `json:"-"` // `json:"updatedAt"`
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	BirthDate     BirthDate `json:"birthDate"`
	ThumbnailPath string    `json:"-"` // `json:"thumbnailPath"`
	Faces         []Face    `json:"faces"`
	IsHidden      bool      `json:"-"` // `json:"isHidden"`
}

type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"-"` // `json:"createdAt"`
	UpdatedAt time.Time `json:"-"` // `json:"updatedAt"`
	Color     string    `json:"color"`
}

type Face struct {
	ID            string `json:"id"`
	SourceType    string `json:"sourceType"`
	ImageHeight   int    `json:"imageHeight"`
	ImageWidth    int    `json:"imageWidth"`
	BoundingBoxX1 int    `json:"boundingBoxX1"`
	BoundingBoxX2 int    `json:"boundingBoxX2"`
	BoundingBoxY1 int    `json:"boundingBoxY1"`
	BoundingBoxY2 int    `json:"boundingBoxY2"`
}

type Asset struct {
	FileCreatedAt  time.Time `json:"-"` // `json:"fileCreatedAt"`
	FileModifiedAt time.Time `json:"-"` // `json:"fileModifiedAt"`
	LocalDateTime  time.Time `json:"localDateTime"`
	UpdatedAt      time.Time `json:"-"` // `json:"updatedAt"`
	StackCount     any       `json:"-"` // `json:"stackCount"`
	DuplicateID    any       `json:"-"` // `json:"duplicateId"`

	ctx context.Context `json:"-"`

	// Data added and used by Kiosk
	mu               *sync.Mutex
	Owner            Owner     `json:"owner"`
	ID               string    `json:"id"`
	DeviceAssetID    string    `json:"-"` // `json:"deviceAssetId"`
	OwnerID          string    `json:"ownerId"`
	DeviceID         string    `json:"-"` // `json:"deviceId"`
	LibraryID        string    `json:"-"` // `json:"libraryId"`
	Type             AssetType `json:"type"`
	OriginalPath     string    `json:"-"` // `json:"originalPath"`
	OriginalFileName string    `json:"originalFileName"`
	OriginalMimeType string    `json:"originalMimeType"`
	Thumbhash        string    `json:"-"` // `json:"thumbhash"`
	Duration         string    `json:"duration"`
	LivePhotoVideoID string    `json:"livePhotoVideoId"`
	Checksum         string    `json:"checksum"`
	Visibility       string    `json:"-"` // `json:"visibility"`

	RatioWanted ImageOrientation `json:"-"`
	MemoryTitle string           `json:"-"`
	Bucket      kiosk.Source     `json:"kioskBucket"`
	BucketID    string           `json:"kioskBucketId"`

	People          []Person `json:"people"`
	Tags            Tags     `json:"tags"`
	UnassignedFaces []Face   `json:"unassignedFaces"`
	AppearsIn       Albums   `json:"kioskAppearsIn"`
	ExifInfo        ExifInfo `json:"exifInfo"`

	requestConfig config.Config `json:"-"`
	IsFavorite    bool          `json:"isFavorite"`
	IsArchived    bool          `json:"isArchived"`
	IsTrashed     bool          `json:"isTrashed"`
	IsOffline     bool          `json:"-"` // `json:"isOffline"`
	HasMetadata   bool          `json:"-"` // `json:"hasMetadata"`
	IsPortrait    bool          `json:"isPortrait"`
	IsLandscape   bool          `json:"isLandscape"`
}

type Album struct {
	ID            string  `json:"id"`
	AlbumName     string  `json:"albumName"`
	Assets        []Asset `json:"assets"`
	AssetCount    int     `json:"assetCount"`
	AssetsOrdered bool    `json:"assetsOrdered"`
}

type Albums []Album

type SearchRandomBody struct {
	City          string   `url:"city,omitempty" json:"city,omitempty"`
	Country       string   `url:"country,omitempty" json:"country,omitempty"`
	CreatedAfter  string   `url:"createdAfter,omitempty" json:"createdAfter,omitempty"`
	CreatedBefore string   `url:"createdBefore,omitempty" json:"createdBefore,omitempty"`
	DeviceID      string   `url:"deviceId,omitempty" json:"deviceId,omitempty"`
	LensModel     string   `url:"lensModel,omitempty" json:"lensModel,omitempty"`
	LibraryID     string   `url:"libraryId,omitempty" json:"libraryId,omitempty"`
	Make          string   `url:"make,omitempty" json:"make,omitempty"`
	Model         string   `url:"model,omitempty" json:"model,omitempty"`
	State         string   `url:"state,omitempty" json:"state,omitempty"`
	TakenAfter    string   `url:"takenAfter,omitempty" json:"takenAfter,omitempty"`
	TakenBefore   string   `url:"takenBefore,omitempty" json:"takenBefore,omitempty"`
	TrashedAfter  string   `url:"trashedAfter,omitempty" json:"trashedAfter,omitempty"`
	TrashedBefore string   `url:"trashedBefore,omitempty" json:"trashedBefore,omitempty"`
	Type          string   `url:"type,omitempty" json:"type,omitempty"`
	UpdatedAfter  string   `url:"updatedAfter,omitempty" json:"updatedAfter,omitempty"`
	UpdatedBefore string   `url:"updatedBefore,omitempty" json:"updatedBefore,omitempty"`
	PersonIDs     []string `url:"personIds,omitempty" json:"personIds,omitempty"`
	TagIDs        []string `url:"tagIds,omitempty" json:"tagIds,omitempty"`
	Size          int      `url:"size,omitempty" json:"size,omitempty"`
	Page          int      `url:"page,omitempty" json:"page,omitempty"`
	IsArchived    bool     `url:"isArchived,omitempty" json:"isArchived,omitempty"`
	IsEncoded     bool     `url:"isEncoded,omitempty" json:"isEncoded,omitempty"`
	IsFavorite    bool     `url:"isFavorite,omitempty" json:"isFavorite,omitempty"`
	IsMotion      bool     `url:"isMotion,omitempty" json:"isMotion,omitempty"`
	IsNotInAlbum  bool     `url:"isNotInAlbum,omitempty" json:"isNotInAlbum,omitempty"`
	IsOffline     bool     `url:"isOffline,omitempty" json:"isOffline,omitempty"`
	IsVisible     bool     `url:"isVisible,omitempty" json:"isVisible,omitempty"`
	WithArchived  bool     `url:"withArchived,omitempty" json:"withArchived,omitempty"`
	WithDeleted   bool     `url:"withDeleted,omitempty" json:"withDeleted,omitempty"`
	WithExif      bool     `url:"withExif,omitempty" json:"withExif,omitempty"`
	WithPeople    bool     `url:"withPeople,omitempty" json:"withPeople,omitempty"`
	WithStacked   bool     `url:"withStacked,omitempty" json:"withStacked,omitempty"`
}

type TagAssetsBody struct {
	IDs []string `url:"ids,omitempty" json:"ids,omitempty"`
}

type AddAssetsToAlbumBody TagAssetsBody

type UpsertTagBody struct {
	Tags []string `url:"tags,omitempty" json:"tags,omitempty"`
}

type UpsertTagResponse []struct {
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  string    `json:"parentId"`
	UpdatedAt time.Time `json:"updatedAt"`
	Value     string    `json:"value"`
}

type SearchMetadataResponse struct {
	Assets struct {
		NextPage string `json:"nextPage"`
		Total    int    `json:"total"`
	} `json:"assets"`
}

type Memory struct {
	CreatedAt time.Time                  `json:"createdAt"`
	UpdatedAt time.Time                  `json:"updatedAt"`
	MemoryAt  time.Time                  `json:"memoryAt"`
	ShowAt    time.Time                  `json:"showAt"`
	HideAt    time.Time                  `json:"hideAt"`
	ID        string                     `json:"id"`
	OwnerID   string                     `json:"ownerId"`
	Type      immich_open_api.MemoryType `json:"type"`
	Assets    []Asset                    `json:"assets"`
	Data      struct {
		Year int `json:"year"`
	} `json:"data"`
	IsSaved bool `json:"isSaved"`
}

type MemoriesResponse []Memory

type AssetFaceResponse struct {
	ID            string `json:"id"`
	Person        Person `json:"person"`
	BoundingBoxX1 int    `json:"boundingBoxX1"`
	BoundingBoxX2 int    `json:"boundingBoxX2"`
	BoundingBoxY1 int    `json:"boundingBoxY1"`
	BoundingBoxY2 int    `json:"boundingBoxY2"`
	ImageHeight   int    `json:"imageHeight"`
	ImageWidth    int    `json:"imageWidth"`
}

type TagAssetsResponse []struct {
	Error   immich_open_api.BulkIdResponseDtoError `json:"error"`
	ID      string                                 `json:"id"`
	Success bool                                   `json:"success"`
}

type AlbumCreateResponse TagAssetsResponse

type AlbumCreateBody struct {
	AlbumName   string `json:"albumName"`
	Description string `json:"description,omitempty"`
}

type UpdateAssetBody struct {
	DateTimeOriginal string  `json:"dateTimeOriginal,omitempty"`
	Description      string  `json:"description,omitempty"`
	LivePhotoVideoID string  `json:"livePhotoVideoId,omitempty"`
	Visibility       string  `json:"visibility,omitempty"`
	Latitude         float64 `json:"latitude,omitempty"`
	Longitude        float64 `json:"longitude,omitempty"`
	Rating           int     `json:"rating,omitempty"`
	IsArchived       bool    `json:"isArchived"`
	IsFavorite       bool    `json:"isFavorite"`
}

// UserAvatarColor defines model for UserAvatarColor.
type UserAvatarColor string

// UserLicense defines model for UserLicense.
type UserLicense struct {
	ActivatedAt   time.Time `json:"activatedAt"`
	ActivationKey string    `json:"activationKey"`
	LicenseKey    string    `json:"licenseKey"`
}

// UserStatus defines model for UserStatus.
type UserStatus string

type UserResponse struct {
	CreatedAt            time.Time       `json:"createdAt"`
	ProfileChangedAt     time.Time       `json:"profileChangedAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
	DeletedAt            *time.Time      `json:"deletedAt"`
	License              *UserLicense    `json:"license"`
	QuotaSizeInBytes     *int64          `json:"quotaSizeInBytes"`
	QuotaUsageInBytes    *int64          `json:"quotaUsageInBytes"`
	StorageLabel         *string         `json:"storageLabel"`
	AvatarColor          UserAvatarColor `json:"avatarColor"`
	Email                string          `json:"email"`
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	OauthID              string          `json:"oauthId"`
	ProfileImagePath     string          `json:"profileImagePath"`
	Status               UserStatus      `json:"status"`
	IsAdmin              bool            `json:"isAdmin"`
	ShouldChangePassword bool            `json:"shouldChangePassword"`
}

type AllPeopleResponse struct {
	People      []Person `json:"people"`
	Hidden      int      `json:"hidden"`
	Total       int      `json:"total"`
	HasNextPage bool     `json:"hasNextPage"`
}

type apiCall func(context.Context, string, string, []byte, ...map[string]string) ([]byte, string, error)

type APIResponse interface {
	Asset |
		[]Asset |
		Album |
		Albums |
		PersonStatistics |
		int |
		SearchMetadataResponse |
		[]Face |
		[]Person |
		[]Tag |
		[]AssetFaceResponse |
		immich_open_api.PersonResponseDto |
		MemoriesResponse |
		TagAssetsResponse |
		AlbumCreateResponse |
		UpsertTagResponse |
		UserResponse |
		AllPeopleResponse |
		[]byte
}

// New returns a new asset instance
func New(ctx context.Context, base config.Config) Asset {
	return Asset{
		requestConfig: base,
		mu:            &sync.Mutex{},
		ctx:           ctx,
	}
}
