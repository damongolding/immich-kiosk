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

	"github.com/damongolding/immich-kiosk/config"
)

var baseConfig config.Config

func init() {
	baseConfig.Load()
}

type ImmichError struct {
	Message    []string `json:"message"`
	Error      string   `json:"error"`
	StatusCode int      `json:"statusCode"`
}

type ImmichImage struct {
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
	Assets []ImmichImage `json:"assets"`
}

// immichApiCall bootstrap from immich api call
func (i *ImmichImage) immichApiCall(apiUrl string) ([]byte, error) {

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

// NewImage returns a new image instance
func NewImage() ImmichImage {
	return ImmichImage{}
}

// GetRandomImage retrieve a random image from Immich
func (i *ImmichImage) GetRandomImage(requestId string) error {

	log.Debug(requestId + " Getting Random image")

	var images []ImmichImage

	u, err := url.Parse(baseConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "api/assets/random",
		RawQuery: "count=1",
	}

	body, err := i.immichApiCall(apiUrl.String())
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
		return fmt.Errorf(immichError.Error, immichError.Message)
	}

	if len(images) == 0 {
		return fmt.Errorf("no images found")
	}

	// We only want images
	if images[0].Type != "IMAGE" {
		log.Debug("Not a image. Trying again")
		return i.GetRandomImage(requestId)
	}

	*i = images[0]

	return nil
}

// GetRandomImageOfPerson retrieve random image of person from Immich
func (i *ImmichImage) GetRandomImageOfPerson(personId, requestId string) error {

	var images []ImmichImage

	u, err := url.Parse(baseConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personId + "/assets",
	}

	body, err := i.immichApiCall(apiUrl.String())
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
		return fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
	}

	if len(images) == 0 {
		return fmt.Errorf("no images found")
	}

	rand.Shuffle(len(images), func(i, j int) {
		images[i], images[j] = images[j], images[i]
	})

	for _, pick := range images {
		// We only want images
		if pick.Type != "IMAGE" {
			continue
		}

		*i = pick
		break
	}

	if i.ID == "" {
		return fmt.Errorf("no images found")
	}

	if log.GetLevel() == log.DebugLevel {
		for _, per := range i.People {
			if per.ID == personId {
				log.Debug(requestId+" Got image of", "perople", per.Name)
				break
			}
		}
	}

	return nil
}

// GetRandomImageOfPersonFromAlbum retrieve random image of person within a specified album from Immich
func (i *ImmichImage) GetRandomImageOfPersonFromAlbum(personId, albumId, requestId string) error {
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

	body, err := i.immichApiCall(apiUrl.String())
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
		return fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
	}

	if len(album.Assets) == 0 {
		return fmt.Errorf("no images found")
	}

	rand.Shuffle(len(album.Assets), func(i, j int) {
		album.Assets[i], album.Assets[j] = album.Assets[j], album.Assets[i]
	})

	for _, pick := range album.Assets {
		// We only want images
		if pick.Type != "IMAGE" {
			continue
		}

		log.Debug("people", "peeps", pick.People)

		for _, person := range pick.People {
			if person.ID == personId {
				*i = pick
				break
			}
		}
	}

	if i.ID == "" {
		return fmt.Errorf("no images found")
	}

	return nil
}

// GetRandomImageFromAlbum retrieve random image within a specified album from Immich
func (i *ImmichImage) GetRandomImageFromAlbum(albumId, requestId string) error {
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

	body, err := i.immichApiCall(apiUrl.String())
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
		return fmt.Errorf("%s : %v", immichError.Error, immichError.Message)
	}

	if len(album.Assets) == 0 {
		return fmt.Errorf("no images found")
	}

	rand.Shuffle(len(album.Assets), func(i, j int) {
		album.Assets[i], album.Assets[j] = album.Assets[j], album.Assets[i]
	})

	for _, pick := range album.Assets {
		// We only want images
		if pick.Type != "IMAGE" {
			continue
		}

		*i = pick
		break
	}

	if i.ID == "" {
		return fmt.Errorf("no images found")
	}

	return nil
}

// GetImagePreview fetches the raw image data from Immich
func (i *ImmichImage) GetImagePreview() ([]byte, error) {

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
