package routes

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/labstack/echo/v4"
)

// Home home endpoint
func Home(c echo.Context) error {

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

	log.Debug(requestId, "path", c.Request().URL.String(), "instanceConfig", instanceConfig)

	pageData := views.PageData{
		KioskVersion: KioskVersion,
		Queries:      queries,
		Config:       instanceConfig,
	}

	return Render(c, http.StatusOK, views.Home(pageData))
}
