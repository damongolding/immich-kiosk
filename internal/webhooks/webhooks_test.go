package webhooks

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
)

func TestWebhookTimeout(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    time.Duration
	}{
		{"zero falls back to default", 0, defaultWebhookTimeout},
		{"negative falls back to default", -5, defaultWebhookTimeout},
		{"positive is honoured", 7, 7 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := webhookTimeout(tt.seconds); got != tt.want {
				t.Errorf("webhookTimeout(%d) = %v, want %v", tt.seconds, got, tt.want)
			}
		})
	}
}

// TestTrigger_HangingReceiverIsBounded checks that with http_timeout: 0 a
// receiver that never responds can't keep Trigger (and its goroutine) alive.
func TestTrigger_HangingReceiverIsBounded(t *testing.T) {
	orig := defaultWebhookTimeout
	defaultWebhookTimeout = 200 * time.Millisecond
	t.Cleanup(func() { defaultWebhookTimeout = orig })

	var received atomic.Int32
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		received.Add(1)
		select { // hang until the test ends
		case <-release:
		case <-r.Context().Done():
		}
	}))
	// Cleanups run last-in first-out: release the handler before Close waits on it.
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) })

	cfg := config.New()
	cfg.Kiosk.HTTPTimeout = 0
	cfg.Webhooks = config.Webhooks{{URL: srv.URL, Event: NewAsset.String()}}
	requestData := &common.RouteRequestData{DeviceID: "device", RequestConfig: *cfg}

	done := make(chan struct{})
	go func() {
		defer close(done)
		Trigger(t.Context(), requestData, "test", NewAsset, common.ViewData{})
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Trigger did not return: webhook request has no bounded lifetime when http_timeout is 0")
	}

	if got := received.Load(); got != 1 {
		t.Errorf("receiver got %d requests, want 1 (the timeout should end a sent request, not prevent it)", got)
	}
}
