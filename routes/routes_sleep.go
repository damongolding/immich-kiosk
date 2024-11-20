package routes

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/utils"
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

		if sleepTime, _ := utils.IsSleepTime(requestConfig.SleepStart, requestConfig.SleepEnd, time.Now()); sleepTime {
			return c.HTML(http.StatusOK, `
			<script>
    			document.body.classList.add('sleep');
       		</script>
			`)
		}

		return c.HTML(http.StatusOK, `
		<script>
    		document.body.classList.remove('sleep');
       	</script>
		`)
	}
}
