package routes

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/damongolding/immich-kiosk/webhooks"
)

// PreviousImage returns an echo.HandlerFunc that handles requests for previous images.
// It retrieves the previous images from the history and renders them.
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
		historyLen := len(requestConfig.History)

		if isSleepMode(requestConfig) || historyLen < 2 {
			return c.NoContent(http.StatusNoContent)
		}

		lastHistoryEntry := requestConfig.History[historyLen-2]
		prevImages := strings.Split(lastHistoryEntry, ",")
		requestConfig.History = requestConfig.History[:historyLen-2]

		ViewData := views.ViewData{
			KioskVersion: kioskDeviceVersion,
			DeviceID:     kioskDeviceID,
			Images:       make([]views.ImageData, len(prevImages)),
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

				img, err := imageToBase64(imgBytes, requestConfig, requestID, kioskDeviceID, "Converted", false)
				if err != nil {
					return fmt.Errorf("converting image to base64: %w", err)
				}

				imgBlur, err := processBlurredImage(imgBytes, requestConfig, requestID, kioskDeviceID, false)
				if err != nil {
					return fmt.Errorf("converting blurred image to base64: %w", err)
				}

				wg.Wait()

				ViewData.Images[i] = views.ImageData{
					ImmichImage:   image,
					ImageData:     img,
					ImageBlurData: imgBlur,
				}
				return nil
			})
		}

		// Wait for all goroutines to complete and check for errors
		if err := g.Wait(); err != nil {
			return RenderError(c, err, "processing images")
		}

		go webhooks.Trigger(requestConfig, KioskVersion, webhooks.PreviousAsset, ViewData)
		return Render(c, http.StatusOK, views.Image(ViewData))
	}
}
