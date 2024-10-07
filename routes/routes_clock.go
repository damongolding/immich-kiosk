package routes

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
)

// Clock clock endpoint
func Clock(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		kioskVersionHeader := c.Request().Header.Get("kiosk-version")
		requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		// If kiosk version on client and server do not match refresh client.
		if kioskVersionHeader != "" && KioskVersion != kioskVersionHeader {
			c.Response().Header().Set("HX-Refresh", "true")
			return c.NoContent(http.StatusOK)
		}

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"ShowTime", requestConfig.ShowTime,
			"TimeFormat", requestConfig.TimeFormat,
			"ShowDate", requestConfig.ShowDate,
			"DateFormat", requestConfig.DateFormat,
		)

		clockTimeFormat := "15:04"
		if requestConfig.TimeFormat == "12" {
			clockTimeFormat = time.Kitchen
		}

		clockDateFormat := utils.DateToLayout(requestConfig.DateFormat)
		if clockDateFormat == "" {
			clockDateFormat = config.DefaultDateLayout
		}

		var data views.ClockData

		t := time.Now()

		switch {
		case (requestConfig.ShowTime && requestConfig.ShowDate):
			data.ClockTime = t.Format(clockTimeFormat)
			data.ClockDate = t.Format(clockDateFormat)
		case requestConfig.ShowTime:
			data.ClockTime = t.Format(clockTimeFormat)
		case requestConfig.ShowDate:
			data.ClockDate = t.Format(clockDateFormat)
		}

		return Render(c, http.StatusOK, views.Clock(data))
	}
}
