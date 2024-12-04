package routes

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

// Sleep sleep mode endpoint
func Sleep(baseConfig *config.Config) echo.HandlerFunc {
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
			"Sleep start", requestConfig.SleepStart,
			"Sleep end", requestConfig.SleepEnd,
		)

		sleepTime, _ := utils.IsSleepTime(requestConfig.SleepStart, requestConfig.SleepEnd, time.Now())

		return Render(c, http.StatusOK, partials.SleepController(sleepTime, requestData.RequestConfig.SleepIcon))

	}
}
