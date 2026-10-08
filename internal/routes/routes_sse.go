package routes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/i18n"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/labstack/echo/v5"
)

const (
	sseTickInterval      = time.Second
	sseHeartbeatInterval = 20 * time.Second
	sseWriteTimeout      = 10 * time.Second
)

// SSE streams server-sent events for a single stream.
// Events are only written when the payload changes, with a periodic comment
// heartbeat to keep proxies from closing idle connections.
func SSE(ctx context.Context, baseConfig *config.Config, maintenanceManager *MaintenanceState) echo.HandlerFunc {
	return func(c *echo.Context) error {
		stream := c.QueryParam("stream")

		// Validate before doing any work
		if stream != kiosk.SSEMainenance && stream != kiosk.SSEClock {
			log.Debug("unknown SSE stream", "stream", stream)
			return c.NoContent(http.StatusBadRequest)
		}

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		log.Debug("SSE client connected", "ip", c.RealIP(), "stream", stream)

		t := i18n.T()
		dm := t("maintenance_message")

		w := c.Response()
		h := w.Header()
		h.Set("Content-Type", "text/event-stream")
		h.Set("Cache-Control", "no-cache")
		h.Set("X-Accel-Buffering", "no") // stop nginx buffering events

		rc := http.NewResponseController(w)
		reqCtx := c.Request().Context()

		ticker := time.NewTicker(sseTickInterval)
		defer ticker.Stop()

		heartbeat := time.NewTicker(sseHeartbeatInterval)
		defer heartbeat.Stop()

		var prev string

		// send builds the payload for this stream and writes it if it changed.
		// A non-nil error means the connection is dead and the handler should return.
		send := func() error {
			var payload string

			switch stream {
			case kiosk.SSEMainenance:
				if maintenanceManager == nil {
					return errors.New("maintenance mode not enabled")
				}
				payload = maintenanceHTML(maintenanceManager, dm)
			case kiosk.SSEClock:
				var tp bytes.Buffer
				if err = partials.Clock(requestData.RequestConfig).Render(reqCtx, &tp); err != nil {
					return err
				}
				payload = tp.String()
			}

			if payload == prev {
				return nil
			}

			prev = payload

			// Drop stalled clients instead of letting them linger
			_ = rc.SetWriteDeadline(time.Now().Add(sseWriteTimeout))

			if err = writeSSEvent(w, stream, payload); err != nil {
				return err
			}

			return rc.Flush()
		}

		// Send initial state straight away so the client doesn't wait for the first tick
		if err = send(); err != nil {
			return nil
		}

		for {
			select {
			case <-reqCtx.Done():
				log.Debug("SSE client disconnected", "ip", c.RealIP(), "stream", stream)
				return nil

			case <-ctx.Done():
				return nil

			case <-ticker.C:
				if err = send(); err != nil {
					return nil
				}

			case <-heartbeat.C:
				_ = rc.SetWriteDeadline(time.Now().Add(sseWriteTimeout))

				if _, err = io.WriteString(w, ": keepalive\n\n"); err != nil {
					return nil
				}
				if err = rc.Flush(); err != nil {
					return nil
				}
			}
		}
	}
}

func writeSSEvent(w io.Writer, event, data string) error {
	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		return err
	}
	for line := range strings.SplitSeq(strings.ReplaceAll(data, "\r\n", "\n"), "\n") {
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\n")
	return err
}

// maintenanceHTML renders the maintenance message. Only the messages read from
// the MAINTENANCE file is escaped; the i18n default is trusted.
func maintenanceHTML(m *MaintenanceState, dm string) string {
	active, msg := m.IsActive()
	if msg == "" {
		msg = dm
	} else {
		msg = html.EscapeString(msg)
	}

	class := "offline-custom-message"
	if active {
		class += " active"
	}

	return fmt.Sprintf(`<span class="%s">%s</span>`, class, msg)
}
