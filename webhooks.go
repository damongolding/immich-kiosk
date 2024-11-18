package webhooks

type WebhookEvent string

const (
    NewAsset      WebhookEvent = "newAsset"
    PreviousAsset WebhookEvent = "previousAsset"
)

type Payload struct {
	Event     string    `json:"event"`
	Timestamp time.Time `json:"timestamp"`
	Assets []immich.ImmichAsset `json:"assets"`
	Meta struct {
		Source  string `json:"source"`
		Version string `json:"version"`
	} `json:"meta"`
	Signature string `json:"signature"`
}

func Trigger(event WebhookEvent, assets) {}