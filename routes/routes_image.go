package routes

import (
	"net/http"

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

		kioskDeviceVersion := c.Request().Header.Get("kiosk-version")
		kioskDeviceID := c.Request().Header.Get("kiosk-device-id")
		requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		// If kiosk version on client and server do not match refresh client.
		if kioskDeviceVersion != "" && KioskVersion != kioskDeviceVersion {
			c.Response().Header().Set("HX-Refresh", "true")
			return c.NoContent(http.StatusOK)
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

		if isSleepMode(requestConfig) {
			return c.NoContent(http.StatusOK)
		}

		// get and use prefetch data (if found)
		if requestConfig.Kiosk.PreFetch {
			if viewData := fromCache(c, kioskDeviceID); viewData != nil {
				go imagePreFetch(requestConfig, c, kioskDeviceID)
				return renderCachedViewData(c, viewData, &requestConfig, requestID, kioskDeviceID)
			}
			log.Debug(requestID, "deviceID", kioskDeviceID, "cache miss for new image")
		}

		ViewData, err := generateViewData(requestConfig, c, kioskDeviceID, false)
		if err != nil {
			RenderError(c, err, "processing image")
		}

		if requestConfig.Kiosk.PreFetch {
			go imagePreFetch(requestConfig, c, kioskDeviceID)
		}

		return Render(c, http.StatusOK, views.Image(ViewData))
	}
}

// NewRawImage returns an echo.HandlerFunc that handles requests for raw images.
// It processes the image without any additional transformations and returns it as a blob.
func NewRawImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		immichImage := immich.NewImage(requestConfig)

		imgBytes, err := processImage(&immichImage, requestConfig, requestID, "", false)
		if err != nil {
			return err
		}

		return c.Blob(http.StatusOK, immichImage.OriginalMimeType, imgBytes)
	}
}
