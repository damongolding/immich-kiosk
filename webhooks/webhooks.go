package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/common"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
	"github.com/damongolding/immich-kiosk/views"
)

type WebhookEvent string

const (
	NewAsset      WebhookEvent = "asset.new"
	PreviousAsset WebhookEvent = "asset.previous"
	PrefetchAsset WebhookEvent = "asset.prefetch"
	CacheFlush    WebhookEvent = "cache.flush"
)

var httpClient = &http.Client{
	Timeout: 20 * time.Second,
}

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

	config := requestData.RequestConfig

	httpClient.Timeout = time.Second * time.Duration(config.Kiosk.HTTPTimeout)

	for _, userWebhook := range config.Webhooks {
		if userWebhook.Event != string(event) {
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
			Config:     config,
			Meta: Meta{
				Source:  "immich-kiosk",
				Version: KioskVersion,
			},
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Error("webhook marshal", "err", err)
			return
		}

		resp, err := httpClient.Post(userWebhook.Url, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			log.Error("webhook post", "err", err)
			return
		}
		defer resp.Body.Close()
	}
}
