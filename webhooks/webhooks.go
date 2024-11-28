package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/common"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/views"
)

type WebhookEvent string

const (
	NewAsset                      WebhookEvent = "asset.new"
	PreviousAsset                 WebhookEvent = "asset.previous"
	PrefetchAsset                 WebhookEvent = "asset.prefetch"
	CacheFlush                    WebhookEvent = "cache.flush"
	UserWebhookTriggerInfoOverlay WebhookEvent = "user.webhook.trigger.info_overlay"
)

type Meta struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

type Payload struct {
	Event      string               `json:"event"`
	Timestamp  string               `json:"timestamp"`
	DeviceID   string               `json:"deviceID"`
	ClientName string               `json:"clientName"`
	AssetCount int                  `json:"assetCount"`
	Assets     []immich.ImmichAsset `json:"assets"`
	Config     config.Config        `json:"config"`
	Meta       Meta                 `json:"meta"`
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
// KioskVersion is the current version string of the kiosk application.
// event specifies which webhook event (NewAsset, PreviousAsset, etc) triggered this webhook.
// viewData contains the images and other view context for the current request.
func Trigger(requestData *common.RouteRequestData, KioskVersion string, event WebhookEvent, viewData views.ViewData) {

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

		if _, err := url.Parse(userWebhook.Url); err != nil {
			log.Error("invalid webhook URL", "url", userWebhook.Url, "err", err)
			continue
		}

		images := make([]immich.ImmichAsset, len(viewData.Images))

		for i, image := range viewData.Images {
			images[i] = image.ImmichImage
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
				Version: KioskVersion,
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

			resp, err := httpClient.Post(userWebhook.Url, "application/json", bytes.NewBuffer(jsonPayload))
			if err != nil {
				log.Error("webhook post", "err", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				log.Error("webhook request failed",
					"url", webhook.Url,
					"status", resp.StatusCode)
				return
			}
		}(userWebhook, jsonPayload)
	}
	wg.Wait()
}
