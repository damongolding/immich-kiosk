package routes

import (
	"bytes"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	videoComponent "github.com/damongolding/immich-kiosk/internal/templates/components/video"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
)

// NewAsset returns an echo.HandlerFunc that handles requests for new assets.
// It manages image processing, caching, and prefetching based on the configuration.
func NewAsset(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
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

		requestCtx := common.CopyContext(c)

		// get and use prefetch data (if found)
		if requestConfig.Kiosk.PreFetch {
			if cachedViewData := fromCache(requestCtx.URL.String(), deviceID); cachedViewData != nil {
				go assetPreFetch(com, requestData, requestCtx)
				go webhooks.Trigger(com.Context(), requestData, KioskVersion, webhooks.NewAsset, cachedViewData[0])

				return renderCachedViewData(c, cachedViewData, &requestConfig, requestID, deviceID, com.Secret())
			}
			log.Debug(requestID, "deviceID", deviceID, "cache miss for new image")
		}

		viewData, err := generateViewData(requestConfig, requestCtx, deviceID, false)
		if err != nil {
			return RenderError(c, err, "retrieving image")
		}

		if requestConfig.Kiosk.PreFetch {
			go assetPreFetch(com, requestData, requestCtx)
		}

		go webhooks.Trigger(com.Context(), requestData, KioskVersion, webhooks.NewAsset, viewData)

		if len(viewData.Assets) > 0 && requestConfig.ExperimentalAlbumVideo && viewData.Assets[0].ImmichAsset.Type == immich.VideoType {
			return Render(c, http.StatusOK, videoComponent.Video(viewData, com.Secret()))
		}

		return Render(c, http.StatusOK, imageComponent.Image(viewData, com.Secret()))

	}
}

// NewRawImage returns an echo.HandlerFunc that handles requests for raw images.
// It processes the image without any additional transformations and returns it as a blob.
func NewRawImage(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
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

		immichAsset := immich.New(com.Context(), requestConfig)

		img, err := processAsset(&immichAsset, immich.ImageOnlyAssetTypes, requestConfig, requestID, "", "", false)
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

// ImageWithID returns an echo.HandlerFunc that handles requests for images by ID.
// It retrieves the image preview based on the provided imageID and returns it as a blob with the appropriate MIME type.
func ImageWithID(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
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

		imageID := c.Param("imageID")
		if imageID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Image ID is required")
		}

		immichAsset := immich.New(com.Context(), requestConfig)
		immichAsset.ID = imageID

		if requestConfig.UseOriginalImage {
			if assetInfoErr := immichAsset.AssetInfo(requestID, ""); assetInfoErr != nil {
				log.Error(requestID, "error getting asset info", "imageID", imageID, "error", assetInfoErr)
				return assetInfoErr
			}
		}

		imgBytes, previewErr := immichAsset.ImagePreview()
		if previewErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "unable to retrieve image")
		}

		imageMime := utils.ImageMimeType(bytes.NewReader(imgBytes))

		return c.Blob(http.StatusOK, imageMime, imgBytes)
	}
}

func TagAsset(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
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

		assetID := c.FormValue("assetID")
		tagName := c.FormValue("tagName")

		if assetID == "" {
			log.Error("Asset ID is required")
			return echo.NewHTTPError(http.StatusBadRequest, "Asset ID is required")
		}

		if tagName == "" {
			log.Error("Tag name is required")
			return echo.NewHTTPError(http.StatusBadRequest, "Tag name is required")
		}

		immichAsset := immich.New(com.Context(), requestConfig)
		immichAsset.ID = assetID

		tag := immich.Tag{
			Name: tagName,
		}

		addTagErr := immichAsset.AddTag(tag)
		if addTagErr != nil {
			log.Error(requestID+" error adding tag", "assetID", assetID, "tagName", tagName, "error", addTagErr)
			return echo.NewHTTPError(http.StatusInternalServerError, "unable to add tag")
		}

		return c.NoContent(http.StatusOK)
	}
}
