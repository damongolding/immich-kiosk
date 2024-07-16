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

var baseConfig config.Config

type PageData struct {
	ImageData      string
	ImageBlurData  string
	Date           string
	FillScreen     bool
	ShowDate       bool
	ShowTime       bool
	BackgroundBlur bool
	Transition     string
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

	return c.Render(http.StatusOK, "index.html", instanceConfig)

}

func NewImage(c echo.Context) error {

	if log.GetLevel() == log.DebugLevel {
		fmt.Println()
		// return c.Render(http.StatusOK, "error.html", ErrorData{Title: "Error title", Message: "Vitae purus pharetra montes metus venenatis ligula. Elit nisl netus tincidunt tempus dictumst eleifend ridiculus. Tempus varius sed duis, nunc proin curae id ligula velit."})
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

	immichImage := immich.NewImage()

	if instanceConfig.Person != "" {
		randomPersonImageErr := immichImage.GetRandomImageOfPerson(instanceConfig.Person, requestId)
		if randomPersonImageErr != nil {
			return c.Render(http.StatusOK, "error.html", ErrorData{Message: randomPersonImageErr.Error()})
		}
	} else {
		randomImageErr := immichImage.GetRandomImage(requestId)
		if randomImageErr != nil {
			return c.Render(http.StatusOK, "error.html", ErrorData{Message: randomImageErr.Error()})
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
	switch {
	case (instanceConfig.ShowDate && instanceConfig.ShowTime):
		date = fmt.Sprintf("%s %s", immichImage.LocalDateTime.Format("02/01/2006"), immichImage.LocalDateTime.Format(time.Kitchen))
		break
	case instanceConfig.ShowDate:
		date = fmt.Sprintf("%s", immichImage.LocalDateTime.Format("02/01/2006"))
		break
	case instanceConfig.ShowTime:
		date = fmt.Sprintf("%s", immichImage.LocalDateTime.Format(time.Kitchen))
		break
	}

	data := PageData{
		ImageData:      img,
		ImageBlurData:  imgBlur,
		Date:           date,
		FillScreen:     instanceConfig.FillScreen,
		BackgroundBlur: instanceConfig.BackgroundBlur,
		Transition:     instanceConfig.Transition,
	}

	return c.Render(http.StatusOK, "image.html", data)
}
