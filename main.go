package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

type PageData struct {
	ImageUrl   string
	Date       string
	FillScreen bool
}

type ErrorData struct {
	Message string
}

var (
	immichApiKey string
	immichUrl    string

	config Config
)

func ImageToBase64(imgBtyes []byte) (string, error) {

	var base64Encoding string

	mimeType := http.DetectContentType(imgBtyes)

	base64Encoding += fmt.Sprintf("data:%s;base64,", mimeType)

	base64Encoding += base64.StdEncoding.EncodeToString(imgBtyes)

	return base64Encoding, nil
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

	config.Load()

	log.Info("Config loaded", "config", config)

	http.Handle("/css/*", http.FileServer(http.Dir("./assets/")))

	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println()

		instanceConfig := config

		referer, err := url.Parse(r.Referer())
		if err != nil {
			log.Fatal(err)
		}

		queries := referer.Query()

		if len(queries) > 0 {
			instanceConfig = instanceConfig.ConfigWithOverrides(queries)
		}

		immichImage := NewImage()

		if instanceConfig.People != "" {
			randomPersonImageErr := immichImage.GetRandomImageOfPerson(instanceConfig.People)
			if randomPersonImageErr != nil {
				showErrorTemplate(w, randomPersonImageErr)
				return
			}
		} else {
			randomImageErr := immichImage.GetRandomImage()
			if randomImageErr != nil {
				showErrorTemplate(w, randomImageErr)
				return
			}
		}

		// imageGet := time.Now()
		imgBytes, err := immichImage.GetImagePreview()
		if err != nil {
			log.Fatal(err)
		}

		// fmt.Println(immichImage.OriginalFileName, ": Got image in", time.Since(imageGet).Seconds())

		// imageConvert := time.Now()
		img, err := ImageToBase64(imgBytes)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(immichImage.OriginalFileName, ": Converted image in", time.Since(imageConvert).Seconds())

		date := fmt.Sprintf("%s %s", immichImage.LocalDateTime.Format("02-01-2006"), immichImage.LocalDateTime.Format(time.Kitchen))

		data := PageData{
			ImageUrl:   img,
			Date:       date,
			FillScreen: instanceConfig.FillScreen,
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

		fmt.Println()

		// create a copy of the global config to use with this instance
		instanceConfig := config

		queries := r.URL.Query()

		if len(queries) > 0 {
			instanceConfig = instanceConfig.ConfigWithOverrides(queries)
		}

		templateFile := "templates/index.html"
		tmpl, err := template.New("index.html").ParseFiles(templateFile)
		if err != nil {
			log.Fatal(err)
		}

		err = tmpl.Execute(w, instanceConfig)
		if err != nil {
			log.Fatal(err)
		}
	})

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
