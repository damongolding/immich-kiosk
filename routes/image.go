package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
)

func processImage(requestConfig config.Config, c echo.Context, isPrefetch bool) (views.PageData, error) {
	requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))
	kioskDeviceId := c.Request().Header.Get("kiosk-device-id")
	immichImage := immich.NewImage(requestConfig)

	peopleAndAlbums := []immich.ImmichAsset{}
	for _, people := range requestConfig.Person {
		peopleAndAlbums = append(peopleAndAlbums, immich.ImmichAsset{Type: "PERSON", ID: people})
	}
	for _, album := range requestConfig.Album {
		peopleAndAlbums = append(peopleAndAlbums, immich.ImmichAsset{Type: "ALBUM", ID: album})
	}

	pickedImage := utils.RandomItem(peopleAndAlbums)

	var err error
	switch pickedImage.Type {
	case "ALBUM":
		err = immichImage.GetRandomImageFromAlbum(pickedImage.ID, requestId)
	case "PERSON":
		err = immichImage.GetRandomImageOfPerson(pickedImage.ID, requestId)
	default:
		err = immichImage.GetRandomImage(requestId)
	}

	if err != nil {
		return views.PageData{}, fmt.Errorf("error getting image: %w", err)
	}

	imageGet := time.Now()
	imgBytes, err := immichImage.GetImagePreview()
	if err != nil {
		return views.PageData{}, fmt.Errorf("error getting image preview: %w", err)
	}
	if isPrefetch {
		log.Debug(requestId, "PREFETCH", kioskDeviceId, "Got image in", time.Since(imageGet).Seconds())
	} else {
		log.Debug(requestId, "Got image in", time.Since(imageGet).Seconds())
	}

	imageConvertTime := time.Now()
	img, err := utils.ImageToBase64(imgBytes)
	if err != nil {
		return views.PageData{}, fmt.Errorf("error converting image to base64: %w", err)
	}
	if isPrefetch {
		log.Debug(requestId, "PREFETCH", kioskDeviceId, "Converted image in", time.Since(imageConvertTime).Seconds())
	} else {
		log.Debug(requestId, "Converted image in", time.Since(imageConvertTime).Seconds())
	}

	var imgBlur string
	if requestConfig.BackgroundBlur && strings.ToLower(requestConfig.ImageFit) != "cover" {
		imageBlurTime := time.Now()
		imgBlurBytes, err := utils.BlurImage(imgBytes)
		if err != nil {
			return views.PageData{}, fmt.Errorf("error blurring image: %w", err)
		}
		imgBlur, err = utils.ImageToBase64(imgBlurBytes)
		if err != nil {
			return views.PageData{}, fmt.Errorf("error converting blurred image to base64: %w", err)
		}
		if isPrefetch {
			log.Debug(requestId, "PREFETCH", kioskDeviceId, "Blurred image in", time.Since(imageBlurTime).Seconds())
		} else {
			log.Debug(requestId, "Blurred image in", time.Since(imageBlurTime).Seconds())
		}
	}

	if len(requestConfig.History) > 10 {
		requestConfig.History = requestConfig.History[len(requestConfig.History)-10:]
	}

	return views.PageData{
		ImmichImage:   immichImage,
		ImageData:     img,
		ImageBlurData: imgBlur,
		Config:        requestConfig,
	}, nil
}

func imagePreFetch(requestConfig config.Config, c echo.Context, kioskDeviceId string) {
	data, err := processImage(requestConfig, c, true)
	if err != nil {
		log.Error("Error in prefetch", "err", err)
		return
	}
	pageDataCache.Set(c.Request().URL.String()+kioskDeviceId, data, cache.DefaultExpiration)
}

func NewImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		kioskDeviceVersion := c.Request().Header.Get("kiosk-version")
		kioskDeviceId := c.Request().Header.Get("kiosk-device-id")
		requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		// If kiosk version on client and server do not match refresh client.
		if kioskDeviceVersion != "" && KioskVersion != kioskDeviceVersion {
			c.Response().Header().Set("HX-Refresh", "true")
			return c.String(http.StatusTemporaryRedirect, "")
		}

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("err overriding config", "err", err)
		}

		log.Debug(
			requestId,
			"method", c.Request().Method,
			"deviceID", kioskDeviceId,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		// get and use prefetch data (if found)
		if requestConfig.Kiosk.PreFetch {
			if data, found := pageDataCache.Get(c.Request().URL.String() + kioskDeviceId); found {
				log.Debug(
					requestId,
					"deviceID", kioskDeviceId,
					"cache hit for new image", true,
				)
				cachedPageData := data.(views.PageData)
				pageDataCache.Delete(c.Request().URL.String())
				go imagePreFetch(requestConfig, c, kioskDeviceId)
				return Render(c, http.StatusOK, views.Image(cachedPageData))
			}
			log.Debug(
				requestId,
				"deviceID", kioskDeviceId,
				"cache miss for new image", false,
			)
		}

		// If it's a GET request for raw image data
		if c.Request().Method == http.MethodGet {
			immichImage := immich.NewImage(requestConfig)
			imgBytes, err := immichImage.GetImagePreview()
			if err != nil {
				return err
			}
			return c.Blob(http.StatusOK, immichImage.OriginalMimeType, imgBytes)
		}

		pageData, err := processImage(requestConfig, c, false)
		if err != nil {
			log.Error("Error processing image", "err", err)
			return Render(c, http.StatusOK, views.Error(views.ErrorData{Title: "Error processing image", Message: err.Error()}))
		}

		if requestConfig.Kiosk.PreFetch {
			go imagePreFetch(requestConfig, c, kioskDeviceId)
		}

		return Render(c, http.StatusOK, views.Image(pageData))
	}
}
