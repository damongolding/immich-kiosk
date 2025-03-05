package routes

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	imageComponent "github.com/damongolding/immich-kiosk/internal/templates/components/image"
	videoComponent "github.com/damongolding/immich-kiosk/internal/templates/components/video"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
)

// PreviousAsset handles requests to show previously viewed assets in the navigation history.
// It processes both images and videos in parallel, retrieving asset info and generating
// regular and blurred preview images.
//
// For each previous asset:
// - Fetches asset info and album details if configured
// - Generates regular and blurred preview images
// - Returns video component for videos when ShowTime is enabled, image component otherwise
//
// Returns 204 No Content if:
// - Sleep mode is active and not disabled
// - Navigation history has fewer than 2 entries
//
// Triggers webhook on successful render.
func PreviousAsset(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
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
			"deviceID", requestData.DeviceID,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		historyLen := len(requestConfig.History)

		if (!requestConfig.DisableSleep && isSleepMode(requestConfig)) || historyLen < 2 {
			return c.NoContent(http.StatusNoContent)
		}

		lastHistoryEntry := requestConfig.History[historyLen-2]
		prevAssets := strings.Split(lastHistoryEntry, ",")
		requestConfig.History = requestConfig.History[:historyLen-2]

		viewData := common.ViewData{
			KioskVersion: KioskVersion,
			DeviceID:     deviceID,
			Assets:       make([]common.ViewImageData, len(prevAssets)),
			Queries:      c.QueryParams(),
			Config:       requestConfig,
		}

		g, _ := errgroup.WithContext(c.Request().Context())

		for i, assetID := range prevAssets {

			parts := strings.Split(assetID, ":")
			if len(parts) != 2 {
				return fmt.Errorf("invalid history entry format: %s", assetID)
			}

			prevAssetsID, currentAssetID, selectedUser := i, parts[0], parts[1]

			g.Go(func(prevAssetsID int, currentAssetID string) func() error {
				return func() error {
					requestConfig.SelectedUser = selectedUser

					asset := immich.New(com.Context(), requestConfig)
					asset.ID = currentAssetID

					var wg sync.WaitGroup
					wg.Add(1)

					go func(asset *immich.Asset, requestID string, wg *sync.WaitGroup) {
						defer wg.Done()
						var processingErr error

						if assetInfoErr := asset.AssetInfo(requestID, deviceID); assetInfoErr != nil {
							processingErr = fmt.Errorf("failed to get asset info: %w", assetInfoErr)
							log.Error(processingErr)
						}

						if requestConfig.ShowAlbumName {
							asset.AlbumsThatContainAsset(requestID, deviceID)
						}

					}(&asset, requestID, &wg)

					imgBytes, previewErr := asset.ImagePreview()
					if previewErr != nil {
						return fmt.Errorf("retrieving image: %w", previewErr)
					}

					img, byteErr := utils.BytesToImage(imgBytes)
					if byteErr != nil {
						return byteErr
					}

					imgString, base64Err := imageToBase64(img, requestConfig, requestID, deviceID, "Converted", false)
					if base64Err != nil {
						return fmt.Errorf("converting image to base64: %w", base64Err)
					}

					imgBlurString, blurErr := processBlurredImage(img, asset.Type, requestConfig, requestID, deviceID, false)
					if blurErr != nil {
						return fmt.Errorf("converting blurred image to base64: %w", blurErr)
					}

					wg.Wait()

					viewData.Assets[prevAssetsID] = common.ViewImageData{
						ImmichAsset:   asset,
						ImageData:     imgString,
						ImageBlurData: imgBlurString,
						User:          selectedUser,
					}
					return nil
				}
			}(prevAssetsID, currentAssetID))
		}

		// Wait for all goroutines to complete and check for errors
		errGroupWait := g.Wait()
		if errGroupWait != nil {
			return RenderError(c, errGroupWait, "processing images")
		}

		go webhooks.Trigger(com.Context(), requestData, KioskVersion, webhooks.PreviousAsset, viewData)

		if len(viewData.Assets) > 0 && requestConfig.ShowTime && viewData.Assets[0].ImmichAsset.Type == immich.VideoType {
			return Render(c, http.StatusOK, videoComponent.Video(viewData, com.Secret()))
		}

		return Render(c, http.StatusOK, imageComponent.Image(viewData, com.Secret()))
	}
}
