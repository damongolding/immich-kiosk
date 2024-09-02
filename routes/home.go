package routes

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/labstack/echo/v4"
)

// Home home endpoint
func Home(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		if log.GetLevel() == log.DebugLevel {
			fmt.Println()
		}

		requestId := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))

		// create a copy of the global config to use with this request
		requestConfig := *baseConfig

		queries := c.Request().URL.Query()

		if len(queries) > 0 {
			requestConfig = requestConfig.ConfigWithOverrides(queries)
		}

		log.Debug(requestId, "path", c.Request().URL.String(), "requestConfig", requestConfig.String())

		pageData := views.PageData{
			KioskVersion: KioskVersion,
			Queries:      queries,
			Config:       requestConfig,
		}

		return Render(c, http.StatusOK, views.Home(pageData))
	}
}
