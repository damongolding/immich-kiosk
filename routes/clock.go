package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/labstack/echo/v4"
)

// Clock clock endpoint
func Clock(baseConfig config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this instance
		instanceConfig := baseConfig

		queries := c.Request().URL.Query()

		if len(queries) > 0 {
			instanceConfig = instanceConfig.ConfigWithOverrides(queries)
		}

		log.Debug(requestId, "path", c.Request().URL.String(), "config", instanceConfig.String())

		t := time.Now()

		clockTimeFormat := "15:04"
		if instanceConfig.TimeFormat == "12" {
			clockTimeFormat = time.Kitchen
		}

		clockDateFormat := utils.DateToLayout(instanceConfig.DateFormat)
		if clockDateFormat == "" {
			clockDateFormat = defaultDateLayout
		}

		var data views.ClockData

		switch {
		case (instanceConfig.ShowTime && instanceConfig.ShowDate):
			data.ClockTime = t.Format(clockTimeFormat)
			data.ClockDate = t.Format(clockDateFormat)
			break
		case instanceConfig.ShowTime:
			data.ClockTime = t.Format(clockTimeFormat)
			break
		case instanceConfig.ShowDate:
			data.ClockDate = t.Format(clockDateFormat)
			break
		}

		return Render(c, http.StatusOK, views.Clock(data))
	}
}
