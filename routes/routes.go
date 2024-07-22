package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
)

var (
	ExampleConfig []byte
	baseConfig    config.Config
)

type PageData struct {
	// ImageData image as base64 data
	ImageData string
	// ImageData blurred image as base64 data
	ImageBlurData string
	// Date image date
	Date string
	// instance config
	config.Config
}

type ErrorData struct {
	Title   string
	Message string
}

func init() {
	err := baseConfig.Load()
	if err != nil {
		log.Fatal(err)
	}
}

// Home home endpoint
func Home(c echo.Context) error {

	if log.GetLevel() == log.DebugLevel {
		fmt.Println()
	}

	requestId := fmt.Sprintf("[%s]", c.Response().Header().Get(echo.HeaderXRequestID))

	// create a copy of the global config to use with this instance
	instanceConfig := baseConfig

	queries := c.Request().URL.Query()

	if len(queries) > 0 {
		instanceConfig = instanceConfig.ConfigWithOverrides(queries)
	}

	log.Debug(requestId, "instanceConfig", instanceConfig)

	pageData := PageData{
		Config: instanceConfig,
	}

	return c.Render(http.StatusOK, "index.html", pageData)

}

// NewImage new image endpoint
func NewImage(c echo.Context) error {

	if log.GetLevel() == log.DebugLevel {
		fmt.Println()
	}

	requestId := fmt.Sprintf("[%s]", c.Response().Header().Get(echo.HeaderXRequestID))

	// create a copy of the global config to use with this instance
	instanceConfig := baseConfig

	referer, err := url.Parse(c.Request().Referer())
	if err != nil {
		log.Error(err)
		return c.Render(http.StatusOK, "error.html", ErrorData{Title: "Error with URL", Message: err.Error()})
	}

	queries := referer.Query()

	if len(queries) > 0 {
		instanceConfig = instanceConfig.ConfigWithOverrides(queries)
	}

	log.Debug(requestId, "config", instanceConfig)

	immichImage := immich.NewImage(baseConfig)

	switch {
	case instanceConfig.Album != "":
		randomAlbumImageErr := immichImage.GetRandomImageFromAlbum(instanceConfig.Album, requestId)
		if randomAlbumImageErr != nil {
			return c.Render(http.StatusOK, "error.html", ErrorData{Title: "Error getting image from album", Message: randomAlbumImageErr.Error()})
		}
		break
	case instanceConfig.Person != "":
		randomPersonImageErr := immichImage.GetRandomImageOfPerson(instanceConfig.Person, requestId)
		if randomPersonImageErr != nil {
			return c.Render(http.StatusOK, "error.html", ErrorData{Title: "Error getting image of person", Message: randomPersonImageErr.Error()})
		}
		break
	default:
		randomImageErr := immichImage.GetRandomImage(requestId)
		if randomImageErr != nil {
			return c.Render(http.StatusOK, "error.html", ErrorData{Title: "Error getting random image", Message: randomImageErr.Error()})
		}
	}

	imageGet := time.Now()
	imgBytes, err := immichImage.GetImagePreview()
	if err != nil {
		return err
	}
	log.Debug(requestId, "Got image in", time.Since(imageGet).Seconds())

	imageConvertTime := time.Now()
	img, err := utils.ImageToBase64(imgBytes)
	if err != nil {
		return err
	}
	log.Debug(requestId, "Converted image in", time.Since(imageConvertTime).Seconds())

	var imgBlur string

	if instanceConfig.BackgroundBlur {
		imageBlurTime := time.Now()
		imgBlurBytes, err := utils.BlurImage(imgBytes)
		if err != nil {
			return err
		}
		imgBlur, err = utils.ImageToBase64(imgBlurBytes)
		if err != nil {
			return err
		}
		log.Debug(requestId, "Blurred image in", time.Since(imageBlurTime).Seconds())
	}

	var date string

	dateFormat := instanceConfig.DateFormat
	if dateFormat == "" {
		dateFormat = "02/01/2006"
	}

	var timeFormat string
	if instanceConfig.TimeFormat == "12" {
		timeFormat = time.Kitchen
	} else {
		timeFormat = time.TimeOnly
	}

	switch {
	case (instanceConfig.ShowDate && instanceConfig.ShowTime):
		date = fmt.Sprintf("%s %s", immichImage.LocalDateTime.Format(dateFormat), immichImage.LocalDateTime.Format(timeFormat))
		break
	case instanceConfig.ShowDate:
		date = fmt.Sprintf("%s", immichImage.LocalDateTime.Format(dateFormat))
		break
	case instanceConfig.ShowTime:
		date = fmt.Sprintf("%s", immichImage.LocalDateTime.Format(timeFormat))
		break
	}

	data := PageData{
		ImageData:     img,
		ImageBlurData: imgBlur,
		Date:          date,
		Config:        instanceConfig,
	}

	return c.Render(http.StatusOK, "image.html", data)
}
