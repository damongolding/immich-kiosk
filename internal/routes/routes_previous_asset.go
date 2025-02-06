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
func PreviousAsset(baseConfig *config.Config) echo.HandlerFunc {
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

		ViewData := common.ViewData{
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

			g.Go(func(id int, currentAssetID string) func() error {
				return func() error {
					requestConfig.SelectedUser = selectedUser

					asset := immich.NewAsset(requestConfig)
					asset.ID = currentAssetID

					var wg sync.WaitGroup
					wg.Add(1)

					go func(asset *immich.ImmichAsset, requestID string, wg *sync.WaitGroup) {
						defer wg.Done()
						var processingErr error

						if err := asset.AssetInfo(requestID, deviceID); err != nil {
							processingErr = fmt.Errorf("failed to get asset info: %w", err)
							log.Error(processingErr)
						}

						if requestConfig.ShowAlbumName {
							asset.AlbumsThatContainAsset(requestID, deviceID)
						}

					}(&asset, requestID, &wg)

					imgBytes, err := asset.ImagePreview()
					if err != nil {
						return fmt.Errorf("retrieving image: %w", err)
					}

					img, err := utils.BytesToImage(imgBytes)
					if err != nil {
						return err
					}

					imgString, err := imageToBase64(img, requestConfig, requestID, deviceID, "Converted", false)
					if err != nil {
						return fmt.Errorf("converting image to base64: %w", err)
					}

					imgBlurString, err := processBlurredImage(img, asset.Type, requestConfig, requestID, deviceID, false)
					if err != nil {
						return fmt.Errorf("converting blurred image to base64: %w", err)
					}

					wg.Wait()

					ViewData.Assets[prevAssetsID] = common.ViewImageData{
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
		if err := g.Wait(); err != nil {
			return RenderError(c, err, "processing images")
		}

		go webhooks.Trigger(requestData, KioskVersion, webhooks.PreviousAsset, ViewData)

		if len(ViewData.Assets) > 0 && requestConfig.ShowTime && ViewData.Assets[0].ImmichAsset.Type == immich.VideoType {
			return Render(c, http.StatusOK, videoComponent.Video(ViewData))
		}

		return Render(c, http.StatusOK, imageComponent.Image(ViewData))
	}
}
