package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"math/rand/v2"

	"github.com/charmbracelet/log"
)

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

func NewImage() ImmichImage {
	return ImmichImage{}
}

func (i *ImmichImage) GetRandomImage() error {

	log.Info("Getting Random image")

	var image []ImmichImage

	u, err := url.Parse(immichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "api/assets/random",
		RawQuery: "count=1",
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl.String(), nil)
	if err != nil {
		log.Error(err)
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-api-key", immichApiKey)

	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return err
	}

	err = json.Unmarshal(body, &image)
	if err != nil {
		log.Error(err, "body", string(body))
		return err
	}

	if len(image) == 0 {
		return fmt.Errorf("no images found")
	}

	// We only want images
	if image[0].Type != "IMAGE" {
		log.Info("Not a image. Trying again")
		return i.GetRandomImage()
	}

	*i = image[0]

	return nil
}

func (i *ImmichImage) GetRandomImageOfPerson(personId string) error {

	var images []ImmichImage

	u, err := url.Parse(immichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personId + "/assets",
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl.String(), nil)
	if err != nil {
		log.Error(err)
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-api-key", immichApiKey)

	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return err
	}

	err = json.Unmarshal(body, &images)
	if err != nil {
		log.Error(err, "body", string(body))
		return err
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

	for _, per := range i.People {
		if per.ID == personId {
			log.Info("Got image of", "perople", per.Name)
			break
		}
	}

	return nil
}

func (i *ImmichImage) GetImagePreview() ([]byte, error) {

	var img []byte

	u, err := url.Parse(immichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "/api/assets/" + i.ID + "/thumbnail",
		RawQuery: "size=preview",
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl.String(), nil)
	if err != nil {
		log.Error(err)
		return img, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-api-key", immichApiKey)

	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return img, err
	}
	defer res.Body.Close()

	img, err = io.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return img, err
	}

	return img, err
}
