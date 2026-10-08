package routes

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/i18n"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/labstack/echo/v5"
)

func SSE(baseConfig *config.Config, maintenanceManager *MaintenanceState) echo.HandlerFunc {
	return func(c *echo.Context) error {
		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		t := i18n.T()
		dm := t("maintenance_message")

		stream := c.QueryParam("stream")

		if stream != "maintenance" && stream != "time" {
			log.Error("missing SSE stream")
			return c.NoContent(http.StatusBadRequest)
		}

		log.Debug("SSE client connected", "ip", c.RealIP(), "stream", stream)

		w := c.Response()
		h := w.Header()
		h.Set("Content-Type", "text/event-stream")
		h.Set("Cache-Control", "no-cache")
		h.Set("Connection", "keep-alive")

		rc := http.NewResponseController(w)
		ctx := c.Request().Context()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		heartbeat := time.NewTicker(time.Second * 20)
		defer heartbeat.Stop()

		var prev string

		for {
			select {
			case <-ctx.Done():
				log.Debug("SSE client disconnected", "ip", c.RealIP())
				return nil

			case <-ticker.C:
				var out string

				switch stream {
				case "maintenance":
					out = maintenanceHTML(maintenanceManager, dm)
				case "time":
					var tp bytes.Buffer
					if err = partials.Clock(requestData.RequestConfig).Render(ctx, &tp); err != nil {
						log.Warn("rendering view", "err", err)
						continue
					}
					out = tp.String()
				}

				if out == prev {
					continue
				}
				prev = out

				if err = writeSSEvent(w, stream, out); err != nil {
					return nil
				}
				if err = rc.Flush(); err != nil {
					return nil
				}

			case <-heartbeat.C:
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

func maintenanceHTML(m *MaintenanceState, dm string) string {
	active, msg := m.IsActive()
	if msg == "" {
		msg = dm
	}
	class := "offline-custom-message"
	if active {
		class += " active"
	}
	return fmt.Sprintf(`<span class="%s">%s</span>`, class, html.EscapeString(msg))
}
