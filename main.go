package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/joho/godotenv"
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
	LivePhotoVideoID any    `json:"livePhotoVideoId"`
	People           []any  `json:"people"`
	Checksum         string `json:"checksum"`
	StackCount       any    `json:"stackCount"`
	IsOffline        bool   `json:"isOffline"`
	HasMetadata      bool   `json:"hasMetadata"`
	DuplicateID      any    `json:"duplicateId"`
}

type HomeData struct {
	ImageUrl string
	Date     string
}

type ErrorData struct {
	Message string
}

var (
	immichApiKey string
	immichUrl    string
)

func ImageToBase64(imgBtyes []byte) (string, error) {

	var base64Encoding string

	mimeType := http.DetectContentType(imgBtyes)

	base64Encoding += fmt.Sprintf("data:%s;base64,", mimeType)

	base64Encoding += base64.StdEncoding.EncodeToString(imgBtyes)

	return base64Encoding, nil
}

func getRandomImage() (ImmichImage, error) {

	var image []ImmichImage

	url := immichUrl + "/api/assets/random?count=1"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return image[0], err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-api-key", "")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return image[0], err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return image[0], err
	}

	err = json.Unmarshal(body, &image)
	if err != nil {
		fmt.Println(err)
		return image[0], err
	}

	// We only want images
	if image[0].Type != "IMAGE" {
		return getRandomImage()
	}

	return image[0], nil
}

func getImagePreview(id string) ([]byte, error) {

	var img []byte

	method := "GET"
	apiUrl, err := url.JoinPath(immichUrl, "/api/assets/"+id+"/thumbnail?size=preview")
	if err != nil {
		fmt.Println(err)
		return img, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, apiUrl, nil)
	if err != nil {
		fmt.Println(err)
		return img, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("x-api-key", "")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return img, err
	}
	defer res.Body.Close()

	img, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return img, err
	}

	return img, err
}

func getImage(id string) ([]byte, error) {

	var img []byte

	url := immichUrl + "/api/assets/" + id + "/original"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return img, err
	}

	req.Header.Add("Accept", "application/octet-stream")
	req.Header.Add("x-api-key", "")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return img, err
	}
	defer res.Body.Close()

	img, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return img, err
	}

	return img, err

}

func showErrorTemplate(w io.Writer, errToShow error) {
	templateFile := "templates/error.html"
	tmpl, err := template.New("error.html").ParseFiles(templateFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = tmpl.Execute(w, ErrorData{Message: errToShow.Error()})
	if err != nil {
		log.Fatal(err)
	}

}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	immichApiKey = os.Getenv("IMMICH_API_KEY")
	immichUrl = os.Getenv("IMMICH_URL")
}

func main() {

	http.Handle("/assets/", http.FileServer(http.Dir("./")))

	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {

		referer, err := url.Parse(r.Referer())
		if err != nil {

		}

		person := referer.Query().Get("people")
		if person != "" {
			fmt.Println(person)
		}

		randomImage, randomImageErr := getRandomImage()
		if randomImageErr != nil {
			showErrorTemplate(w, randomImageErr)
			return
		}

		imageGet := time.Now()
		imgBytes, err := getImagePreview(randomImage.ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(randomImage.OriginalFileName, ": Got image in", time.Since(imageGet).Seconds())

		imageConvert := time.Now()
		img, err := ImageToBase64(imgBytes)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(randomImage.OriginalFileName, ": Converted image in", time.Since(imageConvert).Seconds())

		date := fmt.Sprintf("%s %s", randomImage.LocalDateTime.Format("02-01-2006"), randomImage.LocalDateTime.Format(time.Kitchen))

		data := HomeData{
			ImageUrl: img,
			Date:     date,
		}

		templateFile := "templates/image.html"
		tmpl, err := template.New("image.html").ParseFiles(templateFile)
		if err != nil {
			log.Fatal(err)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		data := HomeData{
			ImageUrl: "/assets/dog.jpg",
			Date:     "13-11-2023 09:12am",
		}

		templateFile := "templates/index.html"
		tmpl, err := template.New("index.html").ParseFiles(templateFile)
		if err != nil {
			log.Fatal(err)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Fatal(err)
		}
	})

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
