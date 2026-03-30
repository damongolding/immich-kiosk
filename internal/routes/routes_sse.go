package routes

import (
	"fmt"
	"net/http"
	"sync"

	"charm.land/log/v2"
	"github.com/labstack/echo/v5"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/mqtt"
)

// sseHub manages all active SSE client connections grouped by client name.
var sseHub = &SSEHub{
	clients:      make(map[string]map[chan string]struct{}),
	knownClients: make(map[string]struct{}),
}

// SSEHub keeps track of connected SSE clients grouped by client name and
// broadcasts events to them.
type SSEHub struct {
	mu           sync.RWMutex
	clients      map[string]map[chan string]struct{} // clientName -> set of channels
	knownClients map[string]struct{}                 // clients that have had HA discovery published
	onNewClient  func(clientName string)             // called the first time a named client connects
}

// SetNewClientHandler registers a function to call when a new named client
// connects for the first time (used to publish HA MQTT discovery).
func (h *SSEHub) SetNewClientHandler(fn func(string)) {
	h.mu.Lock()
	h.onNewClient = fn
	h.mu.Unlock()
}

// subscribe registers a new SSE channel for the given client name.
func (h *SSEHub) subscribe(clientName string) chan string {
	ch := make(chan string, 4)
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[clientName] == nil {
		h.clients[clientName] = make(map[chan string]struct{})
	}
	h.clients[clientName][ch] = struct{}{}

	// Notify on first ever connection for this name
	if _, known := h.knownClients[clientName]; !known && h.onNewClient != nil {
		h.knownClients[clientName] = struct{}{}
		go h.onNewClient(clientName)
	}

	return ch
}

// unsubscribe removes a channel from the hub.
func (h *SSEHub) unsubscribe(clientName string, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients[clientName], ch)
	if len(h.clients[clientName]) == 0 {
		delete(h.clients, clientName)
		delete(h.knownClients, clientName)
	}
	close(ch)
}

// broadcast sends an event to clients. If target is empty, all clients receive
// the event. Otherwise only the named client group receives it.
func (h *SSEHub) broadcast(target, event string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	send := func(ch chan string) {
		select {
		case ch <- event:
		default:
			// client too slow, skip
		}
	}

	if target == "" {
		for _, channels := range h.clients {
			for ch := range channels {
				send(ch)
			}
		}
		return
	}

	for ch := range h.clients[target] {
		send(ch)
	}
}

// SetNewClientHandler wires up a callback for when a new named client connects.
func SetNewClientHandler(fn func(string)) {
	sseHub.SetNewClientHandler(fn)
}

// MQTTCommandHandler returns an mqtt.Handler that broadcasts navigation commands
// as SSE events to the correct client group (or all if target is empty).
func MQTTCommandHandler() mqtt.Handler {
	return func(cmd mqtt.Command, target string) {
		log.Debug("MQTT command dispatched via SSE", "command", cmd, "target", target)
		sseHub.broadcast(target, string(cmd))
	}
}

// SSEEvents is an Echo handler that keeps an SSE connection open and streams
// navigation commands to the browser. Pass ?client=<name> in the URL to
// register the connection under a specific client name.
func SSEEvents(_ *config.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		w := c.Response()
		r := c.Request()

		flusher, ok := w.(http.Flusher)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "streaming not supported")
		}

		clientName := c.QueryParam("client")
		if clientName == "" {
			clientName = "_global"
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)

		ch := sseHub.subscribe(clientName)
		defer sseHub.unsubscribe(clientName, ch)

		fmt.Fprintf(w, ": connected client=%s\n\n", clientName)
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return nil
			case event, ok := <-ch:
				if !ok {
					return nil
				}
				fmt.Fprintf(w, "event: kiosk-command\ndata: %s\n\n", event)
				flusher.Flush()
			}
		}
	}
}
