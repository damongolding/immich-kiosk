package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
)

// Clock clock endpoint
func Clock(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

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

		return Render(c, http.StatusOK, views.Clock(requestConfig))
	}
}
