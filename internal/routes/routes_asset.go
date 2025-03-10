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
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
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

		viewData, err := generateViewData(requestConfig, requestCtx, requestID, deviceID, false)
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

// TagAsset returns an echo.HandlerFunc that handles requests to add tags to assets.
// It validates the asset ID and tag name parameters, then adds the specified tag to the asset.
// Returns HTTP 200 on success or appropriate error codes if validation fails or tag addition fails.
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

		return c.String(http.StatusOK, "SUCCESS")
	}
}

// FavouriteAsset returns an echo.HandlerFunc that handles requests to favorite/unfavorite assets.
// It validates the asset ID parameter and updates the favorite status of the specified asset.
// Parameters:
//   - baseConfig: Pointer to the global configuration
//   - com: Common utility functions and dependencies
//   - favouriteAsset: Boolean indicating whether to favorite (true) or unfavorite (false) the asset
//
// Returns:
//   - An echo.HandlerFunc that processes the favorite/unfavorite request
//   - HTTP 200 with updated like button HTML on success
//   - HTTP 400 if asset ID is missing
//   - HTTP 500 if favorite operation fails
func FavouriteAsset(baseConfig *config.Config, com *common.Common, favouriteAsset bool) echo.HandlerFunc {
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

		if assetID == "" {
			log.Error("Asset ID is required")
			return echo.NewHTTPError(http.StatusBadRequest, "Asset ID is required")
		}

		immichAsset := immich.New(com.Context(), requestConfig)
		immichAsset.ID = assetID

		favouriteErr := immichAsset.FavouriteStatus(favouriteAsset)
		if favouriteErr != nil {
			log.Error(requestID+" error favouriting asset", "assetID", assetID, "error", favouriteErr)
			return echo.NewHTTPError(http.StatusInternalServerError, "unable to favourite asset")
		}

		cacheErr := removeAssetCache(requestConfig.ImmichURL, requestData.DeviceID, assetID, requestConfig.SelectedUser)
		if cacheErr != nil {
			log.Error(requestID+" error removing asset from cache", "assetID", assetID, "error", cacheErr)
		}

		return Render(c, http.StatusOK, partials.LikeButton(assetID, favouriteAsset, true))
	}
}

// HideAsset returns an echo.HandlerFunc that handles requests to hide/unhide assets via tags.
// It adds or removes a tag from an asset based on the hideAsset parameter.
// Parameters:
//   - baseConfig: Pointer to the global configuration
//   - com: Common utility functions and dependencies
//   - hideAsset: Boolean indicating whether to hide (true) or unhide (false) the asset
//
// Returns:
//   - An echo.HandlerFunc that processes the hide/unhide request
//   - HTTP 200 with updated hide button HTML on success
//   - HTTP 400 if asset ID or tag name is missing
//   - HTTP 500 if tag addition/removal fails
func HideAsset(baseConfig *config.Config, com *common.Common, hideAsset bool) echo.HandlerFunc {
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

		if hideAsset {
			err := immichAsset.AddTag(tag)
			if err != nil {
				log.Error(requestID+" error adding tag to asset", "assetID", assetID, "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "unable to add tag to asset")
			}

		} else {
			err := immichAsset.RemoveTag(tag)
			if err != nil {
				log.Error(requestID+" error removing tag from asset", "assetID", assetID, "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "unable to remove tag from asset")
			}
		}

		cacheErr := removeAssetCache(requestConfig.ImmichURL, requestData.DeviceID, assetID, requestConfig.SelectedUser)
		if cacheErr != nil {
			log.Error(requestID+" error removing asset from cache", "assetID", assetID, "error", cacheErr)
		}

		return Render(c, http.StatusOK, partials.HideButton(assetID, hideAsset))
	}
}
