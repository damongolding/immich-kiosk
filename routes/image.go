package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
)

// NewImage returns an echo.HandlerFunc that handles requests for new images.
// It manages image processing, caching, and prefetching based on the configuration.
func NewImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		kioskDeviceVersion := c.Request().Header.Get("kiosk-version")
		kioskDeviceID := c.Request().Header.Get("kiosk-device-id")
		requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		// If kiosk version on client and server do not match refresh client.
		if kioskDeviceVersion != "" && KioskVersion != kioskDeviceVersion {
			c.Response().Header().Set("HX-Refresh", "true")
			return c.String(http.StatusTemporaryRedirect, "")
		}

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"deviceID", kioskDeviceID,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		if isSleepTime, _ := utils.IsSleepTime(requestConfig.SleepStart, requestConfig.SleepEnd, time.Now()); isSleepTime {
			return nil
		}

		// get and use prefetch data (if found)
		if requestConfig.Kiosk.PreFetch {
			if pageData := fromCache(c, kioskDeviceID); pageData != nil {
				return renderCachedPageData(c, pageData, &requestConfig, requestID, kioskDeviceID)
			}
			log.Debug(requestID, "deviceID", kioskDeviceID, "cache miss for new image")
		}

		pageData, err := generatePageData(requestConfig, c)
		if err != nil {
			RenderError(c, err, "processing image")
		}

		if requestConfig.Kiosk.PreFetch {
			go imagePreFetch(2, requestConfig, c, kioskDeviceID)
		}

		return Render(c, http.StatusOK, views.Image(pageData...))
	}
}

// NewRawImage returns an echo.HandlerFunc that handles requests for raw images.
// It processes the image without any additional transformations and returns it as a blob.
func NewRawImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestId,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		immichImage := immich.NewImage(requestConfig)

		imgBytes, err := processImage(&immichImage, requestConfig, requestId, "", false)
		if err != nil {
			return err
		}

		return c.Blob(http.StatusOK, immichImage.OriginalMimeType, imgBytes)
	}
}
