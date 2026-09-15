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

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich_open_api"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
)

type (
	ImageOrientation string
	AssetType        string
	AssetOrder       string
	AssetVisibility  string
)

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

	Archive  AssetVisibility = "archive"
	Hidden   AssetVisibility = "hidden"
	Locked   AssetVisibility = "locked"
	Timeline AssetVisibility = "timeline"

	SearchRandomEndpoint   = "api/search/random"
	SearchMetadataEndpoint = "api/search/metadata"
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

	ImageOnlyAssetTypes = []AssetType{ImageType}
	VideoOnlyAssetTypes = []AssetType{VideoType}
	AllAssetTypes       = []AssetType{ImageType, VideoType}
)

type PersonStatistics struct {
	Assets int `json:"assets"`
}

type Error struct {
	Path    []string `json:"path"`
	Message string   `json:"message"`
}

type ErrorResponse struct {
	Message string  `json:"message"`
	Errors  []Error `json:"errors"`
}

type Owner struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type ExifInfo struct {
	City             string    `json:"city"`
	Country          string    `json:"country"`
	DateTimeOriginal time.Time `json:"dateTimeOriginal"`
	Description      string    `json:"description"`
	ExifImageHeight  int       `json:"exifImageHeight"`
	ExifImageWidth   int       `json:"exifImageWidth"`
	ExposureTime     string    `json:"exposureTime"`
	FileSizeInByte   int       `json:"fileSizeInByte"`
	FNumber          float64   `json:"fNumber"`
	FocalLength      float64   `json:"focalLength"`
	ImageOrientation ImageOrientation
	Iso              int       `json:"iso"`
	Latitude         float64   `json:"latitude"`
	LensModel        string    `json:"lensModel"`
	Longitude        float64   `json:"longitude"`
	Make             string    `json:"make"`
	Model            string    `json:"model"`
	ModifyDate       time.Time `json:"modifyDate"`
	Orientation      string    `json:"orientation"`
	ProjectionType   any       `json:"-"` // `json:"projectionType"`
	Rating           int       `json:"rating"`
	State            string    `json:"state"`
	TimeZone         string    `json:"timeZone"`
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
	Name      string    `json:"name"`  // e.g "child"
	Value     string    `json:"value"` // e.g "parent/child"
	CreatedAt time.Time `json:"-"`     // `json:"createdAt"`
	UpdatedAt time.Time `json:"-"`     // `json:"updatedAt"`
	Color     string    `json:"color,omitempty"`
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
	Checksum         string          `json:"checksum"`
	DuplicateID      any             `json:"-"`        // `json:"duplicateId"`
	Duration         int64           `json:"duration"` // milliseconds
	ExifInfo         ExifInfo        `json:"exifInfo"`
	FileCreatedAt    time.Time       `json:"-"` // `json:"fileCreatedAt"`
	FileModifiedAt   time.Time       `json:"-"` // `json:"fileModifiedAt"`
	HasMetadata      bool            `json:"-"` // `json:"hasMetadata"`
	ID               string          `json:"id"`
	IsArchived       bool            `json:"isArchived"`
	IsEdited         bool            `json:"isEdited"`
	IsFavorite       bool            `json:"isFavorite"`
	IsOffline        bool            `json:"-"` // `json:"isOffline"`
	IsTrashed        bool            `json:"isTrashed"`
	LibraryID        string          `json:"-"` // `json:"libraryId"`
	LivePhotoVideoID string          `json:"livePhotoVideoId"`
	LocalDateTime    time.Time       `json:"localDateTime"`
	OriginalFileName string          `json:"originalFileName"`
	OriginalMimeType string          `json:"originalMimeType"`
	OriginalPath     string          `json:"-"` // `json:"originalPath"`
	Owner            Owner           `json:"owner"`
	OwnerID          string          `json:"ownerId"`
	People           []Person        `json:"people"`
	StackCount       any             `json:"-"` // `json:"stackCount"`
	Tags             Tags            `json:"tags"`
	Thumbhash        string          `json:"-"` // `json:"thumbhash"`
	Type             AssetType       `json:"type"`
	UpdatedAt        time.Time       `json:"-"` // `json:"updatedAt"`
	Visibility       AssetVisibility `json:"visibility"`

	// Kiosk specific fields
	AppearsIn       Albums          `json:"kioskAppearsIn"`
	Bucket          kiosk.Source    `json:"kioskBucket"`
	BucketID        string          `json:"kioskBucketId"`
	ctx             context.Context `json:"-" msgpack:"-"`
	DeviceID        string          `json:"-"`
	IsLandscape     bool            `json:"isLandscape"`
	IsPortrait      bool            `json:"isPortrait"`
	MemoryTitle     string          `json:"-"`
	mu              *sync.Mutex
	RatioWanted     ImageOrientation `json:"-"`
	requestConfig   config.Config    `json:"-"`
	ServedMimeType  string           `json:"servedMimeType"` // mime type served from the Immich server
	UnassignedFaces []Face           `json:"unassignedFaces"`
}

type AlbumUsers struct {
	User Owner  `json:"user"`
	Role string `json:"role"`
}

type Album struct {
	AlbumName                  string       `json:"albumName"`
	Description                string       `json:"description"`
	AlbumThumbnailAssetID      string       `json:"albumThumbnailAssetId"`
	CreatedAt                  string       `json:"createdAt"`
	UpdatedAt                  string       `json:"updatedAt"`
	ID                         string       `json:"id"`
	AlbumUsers                 []AlbumUsers `json:"albumUsers"`
	Shared                     bool         `json:"shared"`
	HasSharedLink              bool         `json:"hasSharedLink"`
	StartDate                  string       `json:"startDate"`
	EndDate                    string       `json:"endDate"`
	AssetCount                 int          `json:"assetCount"`
	IsActivityEnabled          bool         `json:"isActivityEnabled"`
	Order                      string       `json:"order"`
	LastModifiedAssetTimestamp string       `json:"lastModifiedAssetTimestamp"`

	// Kiosk specific fields
	Assets []Asset `json:"assets"`
}

type Albums []Album

type StringFilter struct {
	Eq    string   `json:"eq,omitempty,omitzero"`
	In    []string `json:"in,omitempty,omitzero"`
	Ne    string   `json:"ne,omitempty,omitzero"`
	NotIn []string `json:"notIn,omitempty,omitzero"`
}

type StringFilterNullable struct {
	Eq    string   `json:"eq,omitempty,omitzero"`
	In    []string `json:"in,omitempty,omitzero"`
	Ne    string   `json:"ne,omitempty,omitzero"`
	NotIn []string `json:"notIn,omitempty,omitzero"`
}

type BoolFilter struct {
	Eq bool `json:"eq,omitempty,omitzero"`
}

type DateFilter struct {
	Eq  time.Time `json:"eq,omitempty,omitzero"`
	Gt  time.Time `json:"gt,omitempty,omitzero"`
	Gte time.Time `json:"gte,omitempty,omitzero"`
	Lt  time.Time `json:"lt,omitempty,omitzero"`
	Lte time.Time `json:"lte,omitempty,omitzero"`
	Ne  time.Time `json:"ne,omitempty,omitzero"`
}

type IDsFilter struct {
	All  []string `json:"all,omitempty,omitzero"`
	Any  []string `json:"any,omitempty,omitzero"`
	None []string `json:"none,omitempty,omitzero"`
}

type StringPatternFilter struct {
	EndsWith   string   `json:"endsWith,omitempty,omitzero"`
	Eq         string   `json:"eq,omitempty,omitzero"`
	In         []string `json:"in,omitempty,omitzero"`
	Like       string   `json:"like,omitempty,omitzero"`
	Ne         string   `json:"ne,omitempty,omitzero"`
	NotIn      []string `json:"notIn,omitempty,omitzero"`
	NotLike    string   `json:"notLike,omitempty,omitzero"`
	StartsWith string   `json:"startsWith,omitempty,omitzero"`
}

type NumberFilter struct {
	Eq    float64   `json:"eq,omitempty,omitzero"`
	Gt    float64   `json:"gt,omitempty,omitzero"`
	Gte   float64   `json:"gte,omitempty,omitzero"`
	In    []float64 `json:"in,omitempty,omitzero"`
	Lt    float64   `json:"lt,omitempty,omitzero"`
	Lte   float64   `json:"lte,omitempty,omitzero"`
	Ne    float64   `json:"ne,omitempty,omitzero"`
	NotIn []float64 `json:"notIn,omitempty,omitzero"`
}

type IDFilter struct {
	Eq string `json:"eq,omitempty,omitzero"`
	Ne string `json:"ne,omitempty,omitzero"`
}

type NumberFilterNullable struct {
	Eq    float64   `json:"eq,omitempty,omitzero"`
	Gt    float64   `json:"gt,omitempty,omitzero"`
	Gte   float64   `json:"gte,omitempty,omitzero"`
	In    []float64 `json:"in,omitempty,omitzero"`
	Lt    float64   `json:"lt,omitempty,omitzero"`
	Lte   float64   `json:"lte,omitempty,omitzero"`
	Ne    float64   `json:"ne,omitempty,omitzero"`
	NotIn []float64 `json:"notIn,omitempty,omitzero"`
}

type IDFilterNullable struct {
	Eq string `json:"eq,omitempty,omitzero"`
	Ne string `json:"ne,omitempty,omitzero"`
}

type StringSimilarityFilter struct {
	Matches string `json:"matches,omitempty,omitzero"`
}

type FilterAssetVisibility struct {
	// Eq Asset visibility
	Eq AssetVisibility   `json:"eq,omitempty,omitzero"`
	In []AssetVisibility `json:"in,omitempty,omitzero"`

	// Ne Asset visibility
	Ne    AssetVisibility   `json:"ne,omitempty,omitzero"`
	NotIn []AssetVisibility `json:"notIn,omitempty,omitzero"`
}

type FilterAssetType struct {
	// Eq Asset type
	Eq AssetType   `json:"eq,omitempty,omitzero"`
	In []AssetType `json:"in,omitempty,omitzero"`

	// Ne Asset type
	Ne    AssetType   `json:"ne,omitempty,omitzero"`
	NotIn []AssetType `json:"notIn,omitempty,omitzero"`
}

type SearchOrder struct {
	// Direction Asset sort order
	Direction AssetOrder `json:"direction,omitempty,omitzero"`
	Field     string     `json:"field,omitempty,omitzero"`
}

type SearchFilter struct {
	AlbumIDs         IDsFilter              `url:"albumIds,omitempty,omitzero" json:"albumIds,omitempty,omitzero"`
	Checksum         StringFilter           `url:"checksum,omitempty,omitzero" json:"checksum,omitempty,omitzero"`
	City             StringFilterNullable   `url:"city,omitempty,omitzero" json:"city,omitempty,omitzero"`
	Country          StringFilterNullable   `url:"country,omitempty,omitzero" json:"country,omitempty,omitzero"`
	CreatedAt        DateFilter             `url:"createdAt,omitempty,omitzero" json:"createdAt,omitempty,omitzero"`
	Description      StringPatternFilter    `url:"description,omitempty,omitzero" json:"description,omitempty,omitzero"`
	EncodedVideoPath StringFilter           `url:"encodedVideoPath,omitempty,omitzero" json:"encodedVideoPath,omitempty,omitzero"`
	FileSizeInBytes  NumberFilter           `url:"fileSizeInBytes,omitempty,omitzero" json:"fileSizeInBytes,omitempty,omitzero"`
	HasAlbums        BoolFilter             `url:"hasAlbums,omitempty,omitzero" json:"hasAlbums,omitempty,omitzero"`
	HasPeople        BoolFilter             `url:"hasPeople,omitempty,omitzero" json:"hasPeople,omitempty,omitzero"`
	HasTags          BoolFilter             `url:"hasTags,omitempty,omitzero" json:"hasTags,omitempty,omitzero"`
	ID               IDFilter               `url:"id,omitempty,omitzero" json:"id,omitempty,omitzero"`
	IsEncoded        BoolFilter             `url:"isEncoded,omitempty,omitzero" json:"isEncoded,omitempty,omitzero"`
	IsFavorite       BoolFilter             `url:"isFavorite,omitempty,omitzero" json:"isFavorite,omitempty,omitzero"`
	IsMotion         BoolFilter             `url:"isMotion,omitempty,omitzero" json:"isMotion,omitempty,omitzero"`
	IsOffline        BoolFilter             `url:"isOffline,omitempty,omitzero" json:"isOffline,omitempty,omitzero"`
	LensModel        StringFilterNullable   `url:"lensModel,omitempty,omitzero" json:"lensModel,omitempty,omitzero"`
	LibraryID        IDFilterNullable       `url:"libraryId,omitempty,omitzero" json:"libraryId,omitempty,omitzero"`
	Make             StringFilterNullable   `url:"make,omitempty,omitzero" json:"make,omitempty,omitzero"`
	Model            StringFilterNullable   `url:"model,omitempty,omitzero" json:"model,omitempty,omitzero"`
	Ocr              StringSimilarityFilter `url:"ocr,omitempty,omitzero" json:"ocr,omitempty,omitzero"`
	// Or               *[]SearchFilterBranch      `json:"or,omitempty,omitzero"`
	OriginalFileName StringPatternFilter  `url:"originalFileName,omitempty,omitzero" json:"originalFileName,omitempty,omitzero"`
	OriginalPath     StringPatternFilter  `url:"originalPath,omitempty,omitzero" json:"originalPath,omitempty,omitzero"`
	PersonIDs        IDsFilter            `url:"personIds,omitempty,omitzero" json:"personIds,omitempty,omitzero"`
	Rating           NumberFilterNullable `url:"rating,omitempty,omitzero" json:"rating,omitempty,omitzero"`
	State            StringFilterNullable `json:"state,omitempty,omitzero"`
	TagIDs           IDsFilter            `url:"tagIds,omitempty,omitzero" json:"tagIds,omitempty,omitzero"`
	TakenAt          DateFilter           `url:"takenAt,omitempty,omitzero" json:"takenAt,omitempty,omitzero"`
	// TrashedAt        *DateFilterNullable        `json:"trashedAt,omitempty,omitzero"`
	Type       FilterAssetType       `url:"type,omitempty,omitzero" json:"type,omitempty,omitzero"`
	UpdatedAt  DateFilter            `url:"updatedAt,omitempty,omitzero" json:"updatedAt,omitempty,omitzero"`
	Visibility FilterAssetVisibility `url:"visibility,omitempty,omitzero" json:"visibility,omitempty,omitzero"`
}

type SearchRandomBody struct {
	Cursor      string       `url:"cursor,omitempty" json:"cursor,omitempty"`
	Filter      SearchFilter `url:"filter,omitempty" json:"filter,omitempty"`
	OrderBy     SearchOrder  `url:"orderBy,omitempty,omitzero" json:"orderBy,omitempty,omitzero"`
	Size        int          `url:"size,omitempty" json:"size,omitempty"`
	WithExif    bool         `url:"withExif,omitempty" json:"withExif,omitempty"`
	WithPeople  bool         `url:"withPeople,omitempty" json:"withPeople,omitempty"`
	WithStacked bool         `url:"withStacked,omitempty" json:"withStacked,omitempty"`

	// Kiosk specific fields
	PaginationComplete bool `url:"paginationComplete,omitempty" json:"paginationComplete,omitempty"`
}

func NewSearchFilterBuilder() *SearchFilter {
	return &SearchFilter{
		Visibility: FilterAssetVisibility{In: []AssetVisibility{Timeline}},
		Type:       FilterAssetType{In: []AssetType{ImageType}},
	}
}

func (b *SearchFilter) WithAlbumsAny(albumIDs ...string) *SearchFilter {
	b.AlbumIDs.Any = albumIDs
	return b
}

func (b *SearchFilter) WithPeopleAll(personIDs ...string) *SearchFilter {
	b.PersonIDs.All = personIDs
	return b
}

func (b *SearchFilter) WithArchived(enabled bool) *SearchFilter {
	if enabled {
		b.Visibility.In = append(b.Visibility.In, Archive)
	}
	return b
}

func (b *SearchFilter) WithVideos(enabled bool) *SearchFilter {
	if enabled {
		b.Type.In = append(b.Type.In, VideoType)
	}
	return b
}

func (b *SearchFilter) WithFilterFavorites(enabled bool) *SearchFilter {
	if enabled {
		b.IsFavorite = BoolFilter{Eq: true}
	}
	return b
}

func (b *SearchFilter) WithRating(r *float32) *SearchFilter {
	if r != nil {
		b.Rating = NumberFilterNullable{Eq: float64(*r)}
	}
	return b
}

func (b *SearchFilter) WithTagsAll(tags ...string) *SearchFilter {
	if len(tags) > 0 {
		b.TagIDs.All = tags
	}
	return b
}

func (b *SearchFilter) WithFilterDate(dateFilter string) *SearchFilter {
	if dateFilter == "" {
		return b
	}

	dateStart, dateEnd, err := determineDateRange(dateFilter)
	if err != nil {
		log.Error("malformed filter", "err", err)
	} else {
		b.TakenAt = DateFilter{
			Gte: dateStart,
			Lte: dateEnd,
		}
	}
	return b
}

func (b *SearchFilter) ExcludePeople(p []string) *SearchFilter {
	if len(p) > 0 {
		b.PersonIDs.None = p
	}
	return b
}

func (b *SearchFilter) ExcludeAlbums(a []string) *SearchFilter {
	if len(a) > 0 {
		b.AlbumIDs.None = a
	}
	return b
}

func (b *SearchFilter) ExcludeTags(t []string) *SearchFilter {
	if len(t) > 0 {
		b.TagIDs.None = t
	}
	return b
}

func (b *SearchFilter) Build() SearchFilter {
	return *b
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
		Items      []Asset `json:"items"`
		NextPage   string  `json:"nextPage"`
		NextCursor string  `json:"nextCursor"`
		Total      int     `json:"total"`
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
	Error   string `json:"error"`
	ID      string `json:"id"`
	Success bool   `json:"success"`
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

type apiCall func(context.Context, string, string, []byte, ...map[string]string) ([]byte, string, bool, error)

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
		StatisticsResponse |
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
