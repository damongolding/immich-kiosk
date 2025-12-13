package routes

import (
	"bytes"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/i18n"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	videoComponent "github.com/damongolding/immich-kiosk/internal/templates/components/video"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/templates/views"
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

		// use history
		if len(requestConfig.History) > 1 && !strings.HasPrefix(requestConfig.History[len(requestConfig.History)-1], "*") {
			return NextHistoryAsset(baseConfig, com, c)
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
			t := i18n.T()
			return RenderError(c, err, t("retrieving_asset"), requestConfig.Duration)
		}

		if requestConfig.Kiosk.PreFetch {
			go assetPreFetch(com, requestData, requestCtx)
		}

		go webhooks.Trigger(com.Context(), requestData, KioskVersion, webhooks.NewAsset, viewData)

		if len(viewData.Assets) > 0 && requestConfig.ShowVideos && viewData.Assets[0].ImmichAsset.Type == immich.VideoType {
			return Render(c, http.StatusOK, videoComponent.Video(viewData, com.Secret()))
		}

		return Render(c, http.StatusOK, imageComponent.Image(viewData, com.Secret()))

	}
}

// Image returns an echo.HandlerFunc that handles requests for raw images.
// It processes the image without any additional transformations and returns it as a blob.
func Image(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
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

		layout := strings.ToLower(strings.TrimSpace(c.QueryParam("layout")))

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		immichAsset := immich.New(com.Context(), requestConfig)

		switch layout {
		case kiosk.PortraitOrientation:
			immichAsset.RatioWanted = immich.PortraitOrientation
		case kiosk.LandscapeOrientation:
			immichAsset.RatioWanted = immich.LandscapeOrientation
		default:
		}

		img, err := processAsset(&immichAsset, requestConfig, requestID, "", "", false)
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

func ImageWithReload(baseConfig *config.Config) echo.HandlerFunc {
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

		queries := c.Request().URL.Query().Encode()

		return Render(c, http.StatusOK, views.ImageWithReload(requestConfig, queries, KioskVersion))
	}
}

// ImageWithID handles HTTP requests to retrieve an image preview by its image ID and returns the image as a blob with the correct MIME type.
// Returns HTTP 400 if the image ID is missing or if the image cannot be retrieved.
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

		imgBytes, _, previewErr := immichAsset.ImagePreview()
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

		// remove asset data from cache as we've changed its tags
		cacheErr := immichAsset.RemoveAssetCache(requestData.DeviceID)
		if cacheErr != nil {
			log.Error(requestID+" error removing asset from cache", "assetID", assetID, "error", cacheErr)
		}

		return c.String(http.StatusOK, "SUCCESS")
	}
}

// LikeAsset returns an echo.HandlerFunc that handles requests to favorite/unfavorite assets.
// It validates the asset ID parameter and updates the favorite status of the specified asset
// based on the configured favorite button action (either mark as favorite or add to album).
//
// Parameters:
//   - baseConfig: Pointer to the global configuration object containing core settings
//   - com: Common module containing context and utility functions
//   - setAssetAsLiked: If true, marks the asset as favorite/adds to album. If false, unfavorites/removes from album
//
// Returns:
//   - An echo.HandlerFunc that processes the favorite/unfavorite request and handles errors
//   - HTTP 200 with updated like button HTML on success
//   - HTTP 400 if required asset ID parameter is missing
//   - HTTP 500 if favorite/album operations fail
//   - Fresh like button HTML is returned regardless of success/failure
func LikeAsset(baseConfig *config.Config, com *common.Common, setAssetAsLiked bool) echo.HandlerFunc {
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
		if u := strings.TrimSpace(c.FormValue("user")); u != "" {
			requestConfig.SelectedUser = u
		}

		if assetID == "" {
			log.Error("Asset ID is required")
			return echo.NewHTTPError(http.StatusBadRequest, "Asset ID is required")
		}

		if baseConfig.Kiosk.DemoMode {
			return Render(c, http.StatusOK, partials.LikeButton(assetID, true, true, true, com.Secret()))
		}

		immichAsset := immich.New(com.Context(), requestConfig)
		immichAsset.ID = assetID
		infoErr := immichAsset.AssetInfo(requestID, requestData.DeviceID)
		if infoErr != nil {
			log.Error(requestID+" error getting asset info", "assetID", assetID, "error", infoErr)
			return infoErr
		}

		var eg error

		// Favourite Asset
		if slices.Contains(requestConfig.LikeButtonAction, kiosk.LikeButtonActionFavorite) {
			favouriteErr := immichAsset.FavouriteStatus(requestData.DeviceID, setAssetAsLiked)
			if favouriteErr != nil {
				log.Error(requestID+" error favouriting asset", "assetID", assetID, "error", favouriteErr)
				eg = errors.Join(eg, favouriteErr)
			}
		}

		// add asset to kiosk liked album
		if slices.Contains(requestConfig.LikeButtonAction, kiosk.LikeButtonActionAlbum) {
			switch setAssetAsLiked {
			case true:
				addErr := immichAsset.AddToKioskLikedAlbum(requestID, requestData.DeviceID)
				if addErr != nil {
					log.Error(requestID+" error adding asset to kiosk liked album", "assetID", assetID, "error", addErr)
					eg = errors.Join(eg, addErr)
				}
			case false:
				rmErr := immichAsset.RemoveFromKioskLikedAlbum(requestID, requestData.DeviceID)
				if rmErr != nil {
					log.Error(requestID+" error removing asset from kiosk liked album", "assetID", assetID, "error", rmErr)
					eg = errors.Join(eg, rmErr)
				}
			}
		}

		// handle error
		if eg != nil {
			return Render(c, http.StatusInternalServerError, partials.LikeButton(assetID, !setAssetAsLiked, false, true, com.Secret()))
		}

		return Render(c, http.StatusOK, partials.LikeButton(assetID, setAssetAsLiked, setAssetAsLiked, true, com.Secret()))
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
		if u := strings.TrimSpace(c.FormValue("user")); u != "" {
			requestConfig.SelectedUser = u
		}

		if assetID == "" {
			log.Error("Asset ID is required")
			return echo.NewHTTPError(http.StatusBadRequest, "Asset ID is required")
		}

		if tagName == "" {
			log.Error("Tag name is required")
			return echo.NewHTTPError(http.StatusBadRequest, "Tag name is required")
		}

		if baseConfig.Kiosk.DemoMode {
			return Render(c, http.StatusOK, partials.HideButton(assetID, !hideAsset, true, com.Secret()))
		}

		immichAsset := immich.New(com.Context(), requestConfig)
		immichAsset.ID = assetID
		infoErr := immichAsset.AssetInfo(requestID, requestData.DeviceID)
		if infoErr != nil {
			log.Error(requestID+" error getting asset info", "assetID", assetID, "error", infoErr)
			return infoErr
		}

		var eg error

		if slices.Contains(requestConfig.HideButtonAction, kiosk.HideButtonActionTag) {
			tag := immich.Tag{
				Name: tagName,
			}

			switch hideAsset {
			case true:
				addTagErr := immichAsset.AddTag(tag)
				if addTagErr != nil {
					log.Error(requestID+" error adding tag to asset", "assetID", assetID, "error", addTagErr)
					eg = errors.Join(eg, addTagErr)
				}
			case false:
				rmTagErr := immichAsset.RemoveTag(tag)
				if rmTagErr != nil {
					log.Error(requestID+" error removing tag from asset", "assetID", assetID, "error", rmTagErr)
					eg = errors.Join(eg, rmTagErr)
				}
			}
		}

		if slices.Contains(requestConfig.HideButtonAction, kiosk.HideButtonActionArchive) {
			archivedErr := immichAsset.ArchiveStatus(requestData.DeviceID, hideAsset)
			if archivedErr != nil {
				log.Error(requestID+" error archiving asset", "assetID", assetID, "error", archivedErr)
				eg = errors.Join(eg, archivedErr)
			}
		}

		if eg != nil {
			return Render(c, http.StatusOK, partials.HideButton(assetID, !hideAsset, true, com.Secret()))
		}

		return Render(c, http.StatusOK, partials.HideButton(assetID, hideAsset, true, com.Secret()))
	}
}
