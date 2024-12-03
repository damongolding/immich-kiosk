package routes

import (
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/views"
	"github.com/damongolding/immich-kiosk/internal/weather"
)

// Home home endpoint
func Home(baseConfig *config.Config) echo.HandlerFunc {
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
			"requestConfig", requestConfig.String(),
		)

		var customCss []byte

		if utils.FileExists("./custom.css") {
			customCss, err = os.ReadFile("./custom.css")
			if err != nil {
				log.Error("reading custom css", "err", err)
			}
		}

		queryParams := c.QueryParams()
		if !queryParams.Has("weather") && requestConfig.HasWeatherDefault {
			queryParams.Set("weather", weather.DefaultLocation())
		}

		viewData := common.ViewData{
			KioskVersion: KioskVersion,
			DeviceID:     utils.GenerateUUID(),
			Queries:      queryParams,
			CustomCss:    customCss,
			Config:       requestConfig,
		}

		return Render(c, http.StatusOK, views.Home(viewData))
	}
}
