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
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
)

// PreviousImage returns an echo.HandlerFunc that handles requests for previous images.
// It retrieves the previous images from the history and renders them.
func PreviousImage(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID
		kioskDeviceID := requestData.DeviceID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"deviceID", requestData.DeviceID,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)
		historyLen := len(requestConfig.History)

		if isSleepMode(requestConfig) || historyLen < 2 {
			return c.NoContent(http.StatusNoContent)
		}

		lastHistoryEntry := requestConfig.History[historyLen-2]
		prevImages := strings.Split(lastHistoryEntry, ",")
		requestConfig.History = requestConfig.History[:historyLen-2]

		ViewData := common.ViewData{
			KioskVersion: KioskVersion,
			DeviceID:     kioskDeviceID,
			Images:       make([]common.ViewImageData, len(prevImages)),
			Queries:      c.QueryParams(),
			Config:       requestConfig,
		}

		g, _ := errgroup.WithContext(c.Request().Context())

		for i, imageID := range prevImages {
			i, imageID := i, imageID
			g.Go(func() error {
				image := immich.NewImage(requestConfig)
				image.ID = imageID

				var wg sync.WaitGroup
				wg.Add(1)

				go func(image *immich.ImmichAsset, requestID string, wg *sync.WaitGroup) {
					defer wg.Done()

					image.AssetInfo(requestID)

				}(&image, requestID, &wg)

				imgBytes, err := image.ImagePreview()
				if err != nil {
					return fmt.Errorf("retrieving image: %w", err)
				}

				img, err := utils.BytesToImage(imgBytes)
				if err != nil {
					return err
				}

				imgString, err := imageToBase64(img, requestConfig, requestID, kioskDeviceID, "Converted", false)
				if err != nil {
					return fmt.Errorf("converting image to base64: %w", err)
				}

				imgBlurString, err := processBlurredImage(img, requestConfig, requestID, kioskDeviceID, false)
				if err != nil {
					return fmt.Errorf("converting blurred image to base64: %w", err)
				}

				wg.Wait()

				ViewData.Images[i] = common.ViewImageData{
					ImmichImage:   image,
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
		return Render(c, http.StatusOK, imageComponent.Image(ViewData))
	}
}
