package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/components/partials"
	"github.com/damongolding/immich-kiosk/config"
)

// Clock clock endpoint
func Clock(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"ShowTime", requestConfig.ShowTime,
			"TimeFormat", requestConfig.TimeFormat,
			"ShowDate", requestConfig.ShowDate,
			"DateFormat", requestConfig.DateFormat,
		)

		return Render(c, http.StatusOK, partials.Clock(requestConfig))
	}
}
