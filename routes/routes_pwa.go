package routes

import (
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/labstack/echo/v4"
)

func paramsToQueries(c echo.Context) {

	config := c.Param("config")
	if config == "" {
		return
	}

	pairs := strings.Split(c.Param("config"), "&")

	for _, pair := range pairs {
		// Split each pair into key and value
		kv := strings.Split(pair, "=")

		if len(kv) == 2 {
			key := kv[0]
			value := kv[1]
			c.QueryParams().Add(key, value)
		}
	}
}

func PWA(baseConfig *config.Config) echo.HandlerFunc {

	return func(c echo.Context) error {

		requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		paramsToQueries(c)

		err := requestConfig.ConfigWithOverrides(c)
		if err != nil {
			log.Error("overriding config", "err", err)
		}

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		var customCss []byte

		if utils.FileExists("./custom.css") {
			customCss, err = os.ReadFile("./custom.css")
			if err != nil {
				log.Error("reading custom css", "err", err)
			}
		}

		viewData := views.ViewData{
			KioskVersion: KioskVersion,
			DeviceID:     utils.GenerateUUID(),
			Queries:      c.QueryParams(),
			CustomCss:    customCss,
			Config:       requestConfig,
		}

		return Render(c, http.StatusOK, views.Home(viewData))
	}
}
