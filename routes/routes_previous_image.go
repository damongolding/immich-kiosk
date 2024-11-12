package routes

import (
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
)

// NewImage returns an echo.HandlerFunc that handles requests for new images.
// It manages image processing, caching, and prefetching based on the configuration.
func PreviousImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		kioskDeviceVersion := c.Request().Header.Get("kiosk-version")
		kioskDeviceID := c.Request().Header.Get("kiosk-device-id")
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
			"deviceID", kioskDeviceID,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		if isSleepMode(requestConfig) || len(requestConfig.History) < 2 {
			return c.NoContent(http.StatusOK)
		}

		lastHistoryEntry := requestConfig.History[len(requestConfig.History)-2]
		prevImages := strings.Split(lastHistoryEntry, ",")
		requestConfig.History = requestConfig.History[:len(requestConfig.History)-2]

		ViewData := views.ViewData{
			KioskVersion: kioskDeviceVersion,
			DeviceID:     kioskDeviceID,
			Images:       make([]views.ImageData, len(prevImages)),
			Queries:      c.QueryParams(),
			Config:       requestConfig,
		}

		for i, imageID := range prevImages {
			image := immich.NewImage(requestConfig)
			image.ID = imageID
			imgBytes, err := image.ImagePreview()
			if err != nil {
				return RenderError(c, err, "retrieving image")
			}

			img, err := imageToBase64(imgBytes, requestConfig, requestID, kioskDeviceID, "Converted", false)
			if err != nil {
				return RenderError(c, err, "converting image to base64")
			}

			imgBlur, err := processBlurredImage(imgBytes, requestConfig, requestID, kioskDeviceID, false)
			if err != nil {
				return RenderError(c, err, "converting blurred image to base64")
			}

			ViewData.Images[i] = views.ImageData{
				ImmichImage:   image,
				ImageData:     img,
				ImageBlurData: imgBlur,
			}

		}

		return Render(c, http.StatusOK, views.Image(ViewData))
	}
}
