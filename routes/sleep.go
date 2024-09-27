package routes

import (
	"fmt"
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
		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		kioskVersionHeader := c.Request().Header.Get("kiosk-version")
		requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		// If kiosk version on client and server do not match refresh client.
		if kioskVersionHeader != "" && KioskVersion != kioskVersionHeader {
			c.Response().Header().Set("HX-Refresh", "true")
			return c.String(http.StatusTemporaryRedirect, "")
		}

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestId,
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
