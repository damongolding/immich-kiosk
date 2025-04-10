package routes

import (
	"net/http"
	"os"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/views"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/weather"
)

// Home home endpoint
func Home(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		c.SetCookie(&http.Cookie{
			Name:   redirectCountHeader,
			MaxAge: -1,
		})

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
			"requestConfig", requestConfig.String(),
		)

		var customCSS []byte

		customCSS, err = loadCustomCSS()
		if err != nil {
			log.Error("loading custom css", "err", err)
		}

		queryParams := c.QueryParams()
		if !queryParams.Has("weather") && requestConfig.HasWeatherDefault {
			queryParams.Set("weather", weather.DefaultLocation())
		}

		viewData := common.ViewData{
			KioskVersion: KioskVersion,
			RequestID:    requestID,
			DeviceID:     utils.GenerateUUID(),
			Queries:      queryParams,
			CustomCSS:    customCSS,
			Config:       requestConfig,
		}

		return Render(c, http.StatusOK, views.Home(viewData))
	}
}

func loadCustomCSS() ([]byte, error) {
	if !utils.FileExists("./custom.css") {
		return nil, nil
	}
	return os.ReadFile("./custom.css")
}
