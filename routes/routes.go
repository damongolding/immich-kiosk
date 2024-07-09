package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-frame/config"
	"github.com/damongolding/immich-frame/immich"
	"github.com/damongolding/immich-frame/utils"
)

var baseConfig config.Config

type PageData struct {
	ImageUrl       string
	Date           string
	FillScreen     bool
	ShowDate       bool
	BackgroundBlur bool
}

type ErrorData struct {
	Message string
}

func init() {
	baseConfig.Load()
}

func Home(c echo.Context) error {
	fmt.Println()

	// create a copy of the global config to use with this instance
	instanceConfig := baseConfig

	queries := c.Request().URL.Query()

	if len(queries) > 0 {
		instanceConfig = instanceConfig.ConfigWithOverrides(queries)
	}

	return c.Render(http.StatusOK, "index.html", instanceConfig)

}

func NewImage(c echo.Context) error {
	fmt.Println()

	// log.Debug("in", "employeeNumber", c.FormValue("employeeNumber"))
	// log.Debug("in", "test", c.FormValue("TEST"))

	instanceConfig := baseConfig

	referer, err := url.Parse(c.Request().Referer())
	if err != nil {
		log.Error(err)
		return c.Render(http.StatusOK, "error.html", ErrorData{Message: err.Error()})
	}

	queries := referer.Query()

	if len(queries) > 0 {
		instanceConfig = instanceConfig.ConfigWithOverrides(queries)
	}

	log.Debug("config used", "config", instanceConfig)

	immichImage := immich.NewImage()

	if instanceConfig.Person != "" {
		randomPersonImageErr := immichImage.GetRandomImageOfPerson(instanceConfig.Person)
		if randomPersonImageErr != nil {
			return c.Render(http.StatusOK, "error.html", ErrorData{Message: randomPersonImageErr.Error()})
		}
	} else {
		randomImageErr := immichImage.GetRandomImage()
		if randomImageErr != nil {
			return c.Render(http.StatusOK, "error.html", ErrorData{Message: randomImageErr.Error()})
		}
	}

	imageGet := time.Now()
	imgBytes, err := immichImage.GetImagePreview()
	if err != nil {
		return err
	}
	log.Debug(immichImage.OriginalFileName, "Got image in", time.Since(imageGet).Seconds())

	imageConvertTime := time.Now()
	img, err := utils.ImageToBase64(imgBytes)
	if err != nil {
		return err
	}
	log.Debug(immichImage.OriginalFileName, "Converted image in", time.Since(imageConvertTime).Seconds())

	date := fmt.Sprintf("%s %s", immichImage.LocalDateTime.Format("02-01-2006"), immichImage.LocalDateTime.Format(time.Kitchen))

	data := PageData{
		ImageUrl:       img,
		Date:           date,
		FillScreen:     instanceConfig.FillScreen,
		ShowDate:       instanceConfig.ShowDate,
		BackgroundBlur: instanceConfig.BackgroundBlur,
	}

	return c.Render(http.StatusOK, "image.html", data)
}
