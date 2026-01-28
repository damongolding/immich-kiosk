package routes

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

// Sleep returns an Echo HTTP handler that displays the sleep mode page, indicating whether the current time falls within the configured sleep period and applying the relevant sleep settings.
func Sleep(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"clock_source", requestConfig.ClockSource,
			"client_time", requestConfig.ClientData.ClientTime,
			"Sleep start", requestConfig.SleepStart,
			"Sleep end", requestConfig.SleepEnd,
		)

		now := time.Now()
		if requestConfig.ClockSource == kiosk.Client {
			clientTime, timeParseErr := time.Parse(time.RFC3339, requestConfig.ClientData.ClientTime)
			if timeParseErr != nil {
				log.Error("Failed to parse client time", timeParseErr)
			} else {
				now = clientTime
			}
		}

		sleepTime, _ := utils.IsSleepTime(requestConfig.SleepStart, requestConfig.SleepEnd, now)

		runningInImmichFrame := c.Request().Header.Get("X-Requested-With") == "com.immichframe.immichframe"

		return Render(c, http.StatusOK, partials.SleepController(sleepTime, requestData.RequestConfig.SleepIcon, requestData.RequestConfig.SleepDimScreen, runningInImmichFrame))

	}
}
