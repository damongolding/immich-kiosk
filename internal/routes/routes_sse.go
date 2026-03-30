package routes

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"charm.land/log/v2"
	"github.com/labstack/echo/v5"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/mqtt"
)

const (
	// sseHeartbeatInterval how often a keepalive comment is sent to prevent
	// proxies and iOS from closing the idle SSE connection.
	sseHeartbeatInterval = 30 * time.Second

	// ssePendingTTL how long a queued command waits for a client to reconnect.
	ssePendingTTL = 10 * time.Second
)

// sseHub manages all active SSE client connections grouped by client name.
var sseHub = &SSEHub{
	clients:      make(map[string]map[chan string]struct{}),
	knownClients: make(map[string]struct{}),
	pending:      make(map[string]pendingCmd),
}

// pendingCmd holds a command that arrived when no client was connected.
type pendingCmd struct {
	event   string
	expires time.Time
}

// SSEHub keeps track of connected SSE clients grouped by client name and
// broadcasts events to them.
type SSEHub struct {
	mu           sync.RWMutex
	clients      map[string]map[chan string]struct{} // clientName -> set of channels
	knownClients map[string]struct{}                 // clients that have had HA discovery published
	pending      map[string]pendingCmd               // commands waiting for a client to reconnect
	onNewClient  func(clientName string)             // called the first time a named client connects
}

// SetNewClientHandler registers a function to call when a new named client
// connects for the first time (used to publish HA MQTT discovery).
func (h *SSEHub) SetNewClientHandler(fn func(string)) {
	h.mu.Lock()
	h.onNewClient = fn
	h.mu.Unlock()
}

// subscribe registers a new SSE channel for the given client name and
// delivers any queued command that arrived while the client was disconnected.
func (h *SSEHub) subscribe(clientName string) chan string {
	ch := make(chan string, 4)
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[clientName] == nil {
		h.clients[clientName] = make(map[chan string]struct{})
	}
	h.clients[clientName][ch] = struct{}{}

	// Deliver a pending command if it hasn't expired yet.
	// Check both a client-specific command and a global broadcast command.
	now := time.Now()
	for _, key := range []string{clientName, ""} {
		if p, ok := h.pending[key]; ok {
			if now.Before(p.expires) {
				select {
				case ch <- p.event:
				default:
				}
			}
			delete(h.pending, key)
			break
		}
	}

	// Notify on first ever connection for this name.
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
// the event. If no client is currently connected the command is queued briefly
// so it can be delivered when the client reconnects.
func (h *SSEHub) broadcast(target, event string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	send := func(ch chan string) {
		select {
		case ch <- event:
		default:
			// client too slow, skip
		}
	}

	delivered := false

	if target == "" {
		for _, channels := range h.clients {
			for ch := range channels {
				send(ch)
				delivered = true
			}
		}
		// Queue globally: store under empty-string key so any reconnecting
		// client (regardless of name) picks it up.
		if !delivered {
			h.pending[""] = pendingCmd{event: event, expires: time.Now().Add(ssePendingTTL)}
		}
		return
	}

	for ch := range h.clients[target] {
		send(ch)
		delivered = true
	}
	if !delivered {
		h.pending[target] = pendingCmd{event: event, expires: time.Now().Add(ssePendingTTL)}
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
//
// A heartbeat comment is sent every 30 s to prevent proxies and iOS from
// closing the idle connection.
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

		heartbeat := time.NewTicker(sseHeartbeatInterval)
		defer heartbeat.Stop()

		for {
			select {
			case <-r.Context().Done():
				return nil
			case <-heartbeat.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
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
