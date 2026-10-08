package routes

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/i18n"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/labstack/echo/v5"
)

func sseMaintenance(w http.ResponseWriter, maintenanceManager *MaintenanceState, dm string) {
	active, msg := maintenanceManager.IsActive()

	if msg == "" {
		msg = dm
	}

	if active {
		msg = fmt.Sprintf("<span class=\"offline-custom-message active\">%s</span>", msg)
	} else {
		msg = fmt.Sprintf("<span class=\"offline-custom-message\">%s</span>", msg)
	}

	if _, err := fmt.Fprintf(w, "event: maintenance\ndata: %s\n\n", msg); err != nil {
		log.Error("sseMaintenance", "err", err)
	}
}

func SSE(baseConfig *config.Config, maintenanceManager *MaintenanceState) echo.HandlerFunc {
	return func(c *echo.Context) error {
		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		t := i18n.T()
		dm := t("maintenance_message")

		stream := c.QueryParam("stream")

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

		for {
			select {
			case <-ctx.Done():
				log.Debug("SSE client disconnected", "ip", c.RealIP())
				return nil // client disconnected
			case <-ticker.C:
				switch stream {
				case "maintenance":
					sseMaintenance(w, maintenanceManager, dm)
				case "time":

					var tp bytes.Buffer

					if err := partials.Clock(requestData.RequestConfig).Render(c.Request().Context(), &tp); err != nil {
						log.Warn("rendering view", "err", err)
						return err
					}

					if _, err := fmt.Fprintf(w, "event: time\ndata: %s\n\n", tp.String()); err != nil {
						return nil
					}
				}

				if err := rc.Flush(); err != nil {
					return nil
				}
			}
		}
	}
}
