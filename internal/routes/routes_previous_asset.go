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

// PreviousAsset returns an echo.HandlerFunc that handles requests for the previously viewed assets.
// It retrieves the previous assets from the navigation history and renders them, handling both images
// and videos. The function processes assets in parallel, retrieving asset info and generating both
// regular and blurred preview images. If sleep mode is active or there is insufficient history,
// returns no content. For videos with ShowTime enabled, renders the video component, otherwise
// renders the image component.
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
		prevImages := strings.Split(lastHistoryEntry, ",")
		requestConfig.History = requestConfig.History[:historyLen-2]

		ViewData := common.ViewData{
			KioskVersion: KioskVersion,
			DeviceID:     deviceID,
			Assets:       make([]common.ViewImageData, len(prevImages)),
			Queries:      c.QueryParams(),
			Config:       requestConfig,
		}

		g, _ := errgroup.WithContext(c.Request().Context())

		for i, assetID := range prevImages {
			i, assetID := i, strings.Replace(assetID, ":video", "", 1)
			g.Go(func() error {
				asset := immich.NewImage(requestConfig)
				asset.ID = assetID

				var wg sync.WaitGroup
				wg.Add(1)

				go func(asset *immich.ImmichAsset, requestID string, wg *sync.WaitGroup) {
					defer wg.Done()

					err := asset.AssetInfo(requestID, deviceID)
					if err != nil {
						log.Error(err)
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

				ViewData.Assets[i] = common.ViewImageData{
					ImmichAsset:   asset,
					ImageData:     imgString,
					ImageBlurData: imgBlurString,
				}
				return nil
			})
		}

		// Wait for all goroutines to complete and check for errors
		if err := g.Wait(); err != nil {
			return RenderError(c, err, "processing images")
		}

		go webhooks.Trigger(requestData, KioskVersion, webhooks.PreviousAsset, ViewData)

		if requestConfig.ShowTime && ViewData.Assets[0].ImmichAsset.Type == immich.VideoType {
			return Render(c, http.StatusOK, videoComponent.Video(ViewData))
		}

		return Render(c, http.StatusOK, imageComponent.Image(ViewData))
	}
}
