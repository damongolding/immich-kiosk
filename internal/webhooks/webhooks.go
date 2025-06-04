package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

type WebhookEvent string

const (
	NewAsset                      WebhookEvent = "asset.new"
	NextHistoryAsset              WebhookEvent = "asset.history.next"
	PreviousHistoryAsset          WebhookEvent = "asset.history.previous"
	PrefetchAsset                 WebhookEvent = "asset.prefetch"
	CacheFlush                    WebhookEvent = "cache.flush"
	UserInteractionClick          WebhookEvent = "user.interaction.click"
	UserWebhookTriggerInfoOverlay WebhookEvent = "user.webhook.trigger.info_overlay"
	UserLikeInfoOverlay           WebhookEvent = "user.like.info_overlay"
	UserUnlikeInfoOverlay         WebhookEvent = "user.unlike.info_overlay"
	UserHideInfoOverlay           WebhookEvent = "user.hide.info_overlay"
	UserUnhideInfoOverlay         WebhookEvent = "user.unhide.info_overlay"
)

type Meta struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

type Payload struct {
	Event      string         `json:"event"`
	Timestamp  string         `json:"timestamp"`
	DeviceID   string         `json:"deviceID"`
	ClientName string         `json:"clientName"`
	AssetCount int            `json:"assetCount"`
	Assets     []immich.Asset `json:"assets"`
	Config     config.Config  `json:"config"`
	Meta       Meta           `json:"meta"`
}

// newHTTPClient creates a new HTTP client with the specified timeout duration.
// It returns a pointer to an http.Client configured with the given timeout.
// This client is used for making webhook requests to external endpoints.
func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

// Trigger handles sending webhook payloads to configured webhooks endpoints for specified events.
// It packages up the current request context, images, and config into a JSON payload and sends it
// to any webhook URLs configured for the event type.
//
// requestData contains the current request context including device ID and client name.
// kioskVersion is the current version string of the kiosk application.
// event specifies which webhook event (NewAsset, PreviousAsset, etc) triggered this webhook.
// viewData contains the images and other view context for the current request.
func Trigger(ctx context.Context, requestData *common.RouteRequestData, kioskVersion string, event WebhookEvent, viewData common.ViewData) {

	if viewData.Kiosk.DemoMode {
		return
	}

	if requestData == nil {
		log.Error("invalid request data")
		return
	}

	requestConfig := requestData.RequestConfig

	httpClient := newHTTPClient(time.Second * time.Duration(requestConfig.Kiosk.HTTPTimeout))

	var wg sync.WaitGroup
	for _, userWebhook := range requestConfig.Webhooks {
		if userWebhook.Event != string(event) {
			continue
		}

		if _, err := url.Parse(userWebhook.URL); err != nil {
			log.Error("invalid webhook URL", "url", userWebhook.URL, "err", err)
			continue
		}

		images := make([]immich.Asset, len(viewData.Assets))

		for i, image := range viewData.Assets {
			images[i] = image.ImmichAsset
		}

		payload := Payload{
			Event:      string(event),
			Timestamp:  time.Now().Format(time.RFC3339),
			DeviceID:   requestData.DeviceID,
			ClientName: requestData.ClientName,
			AssetCount: len(images),
			Assets:     images,
			Config:     requestConfig,
			Meta: Meta{
				Source:  "immich-kiosk",
				Version: kioskVersion,
			},
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Error("webhook marshal", "err", err)
			continue
		}

		wg.Add(1)
		go func(webhook config.Webhook, payload []byte) {
			defer wg.Done()

			req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewBuffer(payload))
			if reqErr != nil {
				log.Error("webhook request creation", "err", reqErr)
				return
			}

			req.Header.Set("Content-Type", "application/json")

			if webhook.Secret != "" {
				signature := utils.CalculateSignature(webhook.Secret, string(payload))
				req.Header.Set("X-Kiosk-Signature-256", "sha256="+signature)
			}

			resp, respErr := httpClient.Do(req)
			if respErr != nil {
				log.Error("webhook post", "err", respErr)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				log.Error("webhook request failed",
					"url", webhook.URL,
					"status", resp.StatusCode)
				return
			}
		}(userWebhook, jsonPayload)
	}

	wg.Wait()
}
