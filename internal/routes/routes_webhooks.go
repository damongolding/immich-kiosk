package routes

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/webhooks"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

func Webhooks(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
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

		calculatedSignature := utils.CalculateSignature(com.Secret(), receivedTimestamp)

		// Compare the received signature with the calculated signature
		if !utils.IsValidSignature(receivedSignature, calculatedSignature) {
			return echo.NewHTTPError(http.StatusForbidden, "Invalid signature")
		}

		switch webhooks.WebhookEvent(kioskWebhookEvent) {
		case webhooks.UserWebhookTriggerInfoOverlay,
			webhooks.UserLikeInfoOverlay,
			webhooks.UserUnlikeInfoOverlay,
			webhooks.UserHideInfoOverlay,
			webhooks.UserUnhideInfoOverlay:

			historyLen := len(requestConfig.History)

			if historyLen == 0 {
				log.Error("webhook request missing history")
				return c.NoContent(http.StatusBadRequest)
			}

			lastHistoryEntry := requestConfig.History[historyLen-1]
			prevImages := strings.Split(lastHistoryEntry, ",")

			viewData := common.ViewData{
				KioskVersion: KioskVersion,
				RequestID:    requestID,
				DeviceID:     deviceID,
				Assets:       make([]common.ViewImageData, len(prevImages)),
				Config:       requestConfig,
			}

			g, _ := errgroup.WithContext(c.Request().Context())

			for i, imageID := range prevImages {

				parts := strings.Split(imageID, ":")
				if len(parts) != 2 {
					return fmt.Errorf("invalid history entry format: %s", imageID)
				}

				currentAssetID := parts[0]

				g.Go(func(currentAssetID string) func() error {
					return func() error {
						image := immich.New(com.Context(), requestConfig)
						image.ID = currentAssetID

						assetInfoErr := image.AssetInfo(requestID, deviceID)
						if assetInfoErr != nil {
							log.Error(assetInfoErr)
							return assetInfoErr
						}

						viewData.Assets[i] = common.ViewImageData{
							ImmichAsset: image,
						}
						return nil
					}
				}(currentAssetID))
			}

			// Wait for all goroutines to complete and check for errors
			errGroupWait := g.Wait()
			if errGroupWait != nil {
				return RenderError(c, errGroupWait, "retrieving image data")
			}

			go webhooks.Trigger(com.Context(), requestData, KioskVersion, webhooks.WebhookEvent(kioskWebhookEvent), viewData)

			return c.String(http.StatusOK, "Triggered")
		case webhooks.NewAsset, webhooks.NextHistoryAsset, webhooks.PreviousHistoryAsset, webhooks.PrefetchAsset, webhooks.CacheFlush:
			// to stop lint moaning
		}

		return c.NoContent(http.StatusNoContent)
	}
}
