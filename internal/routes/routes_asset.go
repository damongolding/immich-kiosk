package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	videoComponent "github.com/damongolding/immich-kiosk/internal/templates/components/video"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
)

// NewAsset returns an echo.HandlerFunc that handles requests for new images.
// It manages image processing, caching, and prefetching based on the configuration.
func NewAsset(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
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

		if !requestConfig.DisableSleep && isSleepMode(requestConfig) {
			return c.NoContent(http.StatusNoContent)
		}

		// get and use prefetch data (if found)
		if requestConfig.Kiosk.PreFetch {
			if cachedViewData := fromCache(c.Request().URL.String(), deviceID); cachedViewData != nil {
				requestEchoCtx := c
				go imagePreFetch(requestData, requestEchoCtx)
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
			requestEchoCtx := c
			go imagePreFetch(requestData, requestEchoCtx)
		}

		go webhooks.Trigger(requestData, KioskVersion, webhooks.NewAsset, viewData)

		if requestConfig.ExperimentalAlbumVideo && viewData.Assets[0].ImmichAsset.Type == immich.VideoType {
			return Render(c, http.StatusOK, videoComponent.Video(viewData))
		}

		return Render(c, http.StatusOK, imageComponent.Image(viewData))

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

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
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

		allowedAssetTypes := []immich.ImmichAssetType{immich.ImageType}

		img, err := processAsset(&immichImage, allowedAssetTypes, requestConfig, requestID, "", "", false)
		if err != nil {
			return err
		}

		imgBytes, err := utils.ImageToBytes(img)
		if err != nil {
			return err
		}

		return c.Blob(http.StatusOK, "image/jpeg", imgBytes)
	}
}
