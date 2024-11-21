package routes

import (
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/common"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/utils"
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

		receivedSignature := c.Request().Header.Get("X-Signature")
		receivedTimestamp := c.Request().Header.Get("X-Timestamp")
		kioskWebhookEvent := c.Request().Header.Get("kiosk-webhook-event")

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"webhook event", kioskWebhookEvent,
		)

		body := c.Request().Body
		defer body.Close()

		payload, err := io.ReadAll(body)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read request body")
		}

		// Expect payload to be empty
		if len(payload) != 0 {
			return c.NoContent(http.StatusBadRequest)
		}

		// 5-minute tolerance
		if !utils.IsValidTimestamp(receivedTimestamp, 300) {
			return c.NoContent(http.StatusBadRequest)
		}

		calculatedSignature := utils.CalculateSignature(common.SharedSecret, receivedTimestamp)

		// Compare the received signature with the calculated signature
		if !utils.IsValidSignature(receivedSignature, calculatedSignature) {
			return echo.NewHTTPError(http.StatusForbidden, "Invalid signature")
		}

		switch kioskWebhookEvent {
		case string(webhooks.UserWebhookTriggerInfoOverlay):

			historyLen := len(requestConfig.History)

			if historyLen == 0 {
				log.Error("webhook request missing history")
				return c.NoContent(http.StatusBadRequest)
			}

			lastHistoryEntry := requestConfig.History[historyLen-1]
			prevImages := strings.Split(lastHistoryEntry, ",")

			viewData := views.ViewData{
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

					image.AssetInfo(requestID)

					viewData.Images[i] = views.ImageData{
						ImmichImage: image,
					}
					return nil
				})
			}

			// Wait for all goroutines to complete and check for errors
			if err := g.Wait(); err != nil {
				return RenderError(c, err, "retrieving image data")
			}

			go webhooks.Trigger(requestData, KioskVersion, webhooks.UserWebhookTriggerInfoOverlay, viewData)

			return c.String(http.StatusOK, "Triggered")

		}

		return c.NoContent(http.StatusNoContent)
	}
}
