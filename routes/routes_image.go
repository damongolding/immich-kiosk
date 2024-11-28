package routes

import (
	"net/http"
	"path"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/damongolding/immich-kiosk/webhooks"
)

// NewImage returns an echo.HandlerFunc that handles requests for new images.
// It manages image processing, caching, and prefetching based on the configuration.
func NewImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID
		deviceID := requestData.DeviceID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"deviceID", deviceID,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		if isSleepMode(requestConfig) {
			return c.NoContent(http.StatusNoContent)
		}

		// get and use prefetch data (if found)
		if requestConfig.Kiosk.PreFetch {
			if cachedViewData := fromCache(c, deviceID); cachedViewData != nil {
				go imagePreFetch(requestConfig, c)
				go webhooks.Trigger(requestData, KioskVersion, webhooks.NewAsset, cachedViewData[0])
				return renderCachedViewData(c, cachedViewData, &requestConfig, requestID, deviceID)
			}
			log.Debug(requestID, "deviceID", deviceID, "cache miss for new image")
		}

		viewData, err := generateViewData(requestConfig, c, deviceID, false)
		if err != nil {
			return RenderError(c, err, "retrieving image")
		}

		if requestConfig.Kiosk.PreFetch {
			go imagePreFetch(requestConfig, c)
		}

		go webhooks.Trigger(requestData, KioskVersion, webhooks.NewAsset, viewData)
		return Render(c, http.StatusOK, views.Image(viewData))
	}
}

// NewRawImage returns an echo.HandlerFunc that handles requests for raw images.
// It processes the image without any additional transformations and returns it as a blob.
func NewRawImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		immichImage := immich.NewImage(requestConfig)

		img, err := processImage(&immichImage, requestConfig, requestID, "", false)
		if err != nil {
			return err
		}

		imgBytes, err := utils.ImageToBytes(img)
		if err != nil {
			return err
		}

		return c.Blob(http.StatusOK, immichImage.OriginalMimeType, imgBytes)
	}
}

type ImageIDParam struct {
	ImageID string `param:"id"`
}

func ServeImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		var imageIDParam ImageIDParam
		if err := c.Bind(&imageIDParam); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid image request")
		}

		imgPath := path.Join("tmp", imageIDParam.ImageID)

		return c.File(imgPath)

	}
}
