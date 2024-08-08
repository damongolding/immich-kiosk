package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/labstack/echo/v4"
)

// NewImage new image endpoint
func NewImage(c echo.Context) error {

	if log.GetLevel() == log.DebugLevel {
		fmt.Println()
	}

	kioskVersionHeader := c.Request().Header.Get("kiosk-version")
	requestId := fmt.Sprintf("[%s]", c.Response().Header().Get(echo.HeaderXRequestID))
	requestingRawImage := c.Request().URL.Query().Has("raw")

	// create a copy of the global config to use with this instance
	instanceConfig := baseConfig

	// If kiosk version on client and server do not match refresh client. Pypass if requestingRawImage is set
	if !requestingRawImage && KioskVersion != kioskVersionHeader {
		c.Response().Header().Set("HX-Refresh", "true")
		return c.String(http.StatusTemporaryRedirect, "")
	}

	queries, err := utils.CombineQueries(c.Request().URL.Query(), c.Request().Referer())
	if err != nil {
		log.Error("err combining queries", "err", err)
	}

	if len(queries) > 0 {
		instanceConfig = instanceConfig.ConfigWithOverrides(queries)
	}

	log.Debug(requestId, "path", c.Request().URL.String(), "config", instanceConfig)

	immichImage := immich.NewImage(baseConfig)

	var peopleAndAlbums []immich.ImmichAsset

	for _, people := range instanceConfig.Person {
  // TODO whitelisting goes here
		peopleAndAlbums = append(peopleAndAlbums, immich.ImmichAsset{Type: "PERSON", ID: people})
	}

	for _, album := range instanceConfig.Album {
		// TODO whitelisting goes here
  peopleAndAlbums = append(peopleAndAlbums, immich.ImmichAsset{Type: "ALBUM", ID: album})
	}

	pickedImage := utils.RandomItem(peopleAndAlbums)

	if !pickedImage.IsOnWhitelist() {
		return Render(c, http.StatusOK, views.Error(views.ErrorData{Title: "ID is not allowed", Message: "Is ID on whitelist"}))
	}

	switch pickedImage.Type {
	case "ALBUM":
		randomAlbumImageErr := immichImage.GetRandomImageFromAlbum(pickedImage.ID, requestId)
		if randomAlbumImageErr != nil {
			log.Error("err getting image from album", "err", randomAlbumImageErr)
			return Render(c, http.StatusOK, views.Error(views.ErrorData{Title: "Error getting image from album", Message: "Is album ID correct?"}))
		}
		break
	case "PERSON":
		randomPersonImageErr := immichImage.GetRandomImageOfPerson(pickedImage.ID, requestId)
		if randomPersonImageErr != nil {
			log.Error("err getting image of person", "err", randomPersonImageErr)
			return Render(c, http.StatusOK, views.Error(views.ErrorData{Title: "Error getting image of person", Message: "Is person ID correct?"}))
		}
		break
	default:
		randomImageErr := immichImage.GetRandomImage(requestId)
		if randomImageErr != nil {
			log.Error("err getting random image", "err", randomImageErr)
			return Render(c, http.StatusOK, views.Error(views.ErrorData{Title: "Error getting random image", Message: "Is Immich running? Are your config settings correct?"}))
		}
		break
	}

	imageGet := time.Now()
	imgBytes, err := immichImage.GetImagePreview()
	if err != nil {
		return err
	}
	log.Debug(requestId, "Got image in", time.Since(imageGet).Seconds())

	// if user wants the raw image data send it
	if requestingRawImage {
		return c.Blob(http.StatusOK, immichImage.OriginalMimeType, imgBytes)
	}

	imageConvertTime := time.Now()
	img, err := utils.ImageToBase64(imgBytes)
	if err != nil {
		return err
	}
	log.Debug(requestId, "Converted image in", time.Since(imageConvertTime).Seconds())

	var imgBlur string

	if instanceConfig.BackgroundBlur && strings.ToLower(instanceConfig.ImageFit) != "cover" {
		imageBlurTime := time.Now()
		imgBlurBytes, err := utils.BlurImage(imgBytes)
		if err != nil {
			log.Error("err blurring image", "err", err)
			return err
		}
		imgBlur, err = utils.ImageToBase64(imgBlurBytes)
		if err != nil {
			log.Error("err converting blurred image to base", "err", err)
			return err
		}
		log.Debug(requestId, "Blurred image in", time.Since(imageBlurTime).Seconds())
	}

	// Image METADATA
	var imageDate string

	var imageTimeFormat string
	if instanceConfig.ImageTimeFormat == "12" {
		imageTimeFormat = time.Kitchen
	} else {
		imageTimeFormat = time.TimeOnly
	}

	imageDateFormat := instanceConfig.ImageDateFormat
	if imageDateFormat == "" {
		imageDateFormat = "02/01/2006"
	}

	switch {
	case (instanceConfig.ShowImageDate && instanceConfig.ShowImageTime):
		imageDate = fmt.Sprintf("%s %s", immichImage.LocalDateTime.Format(imageDateFormat), immichImage.LocalDateTime.Format(imageTimeFormat))
		break
	case instanceConfig.ShowImageDate:
		imageDate = fmt.Sprintf("%s", immichImage.LocalDateTime.Format(imageDateFormat))
		break
	case instanceConfig.ShowImageTime:
		imageDate = fmt.Sprintf("%s", immichImage.LocalDateTime.Format(imageTimeFormat))
		break
	}

	data := views.PageData{
		ImageData:     img,
		ImageBlurData: imgBlur,
		ImageDate:     imageDate,
		Config:        instanceConfig,
	}

	return Render(c, http.StatusOK, views.Image(data))

}
