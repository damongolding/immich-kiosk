package routes

import (
	"net/http"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/damongolding/immich-kiosk/webhooks"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

func Webhooks(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID
		kioskDeviceID := requestData.DeviceID

		kioskWebhookEvent := c.Request().Header.Get("kiosk-webhook-event")

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"webhook event", kioskWebhookEvent,
		)

		switch kioskWebhookEvent {
		case string(webhooks.UserWebhookTriggerInfoOverlay):

			historyLen := len(requestConfig.History)

			lastHistoryEntry := requestConfig.History[historyLen-1]
			prevImages := strings.Split(lastHistoryEntry, ",")

			ViewData := views.ViewData{
				KioskVersion: KioskVersion,
				DeviceID:     kioskDeviceID,
				Images:       make([]views.ImageData, len(prevImages)),
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

					wg.Wait()

					ViewData.Images[i] = views.ImageData{
						ImmichImage: image,
					}
					return nil
				})
			}

			// Wait for all goroutines to complete and check for errors
			if err := g.Wait(); err != nil {
				return RenderError(c, err, "retrieving image data")
			}

			go webhooks.Trigger(requestData, KioskVersion, webhooks.UserWebhookTriggerInfoOverlay, ViewData)
		}

		return c.NoContent(http.StatusNoContent)
	}
}
