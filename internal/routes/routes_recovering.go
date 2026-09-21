package routes

import (
	"embed"
	"net/http"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/views"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/labstack/echo/v5"
)

func Recovering(baseConfig *config.Config, public *embed.FS) echo.HandlerFunc {
	return func(c *echo.Context) error {
		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID
		deviceID := requestData.DeviceID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		var customCSS []byte

		customCSS, err = utils.LoadCustomCSS()
		if err != nil {
			log.Error("loading custom css", "err", err)
		}

		viewData := common.ViewData{
			KioskVersion: KioskVersion,
			RequestID:    requestID,
			DeviceID:     deviceID,
			CustomCSS:    customCSS,
			Config:       requestConfig,
		}

		css, err := public.ReadFile("frontend/public/assets/css/kiosk.css")
		if err != nil {
			return err
		}

		return Render(c, http.StatusOK, views.Recovering(viewData.SystemLang, KioskVersion, css, viewData.CustomCSS, baseConfig.CustomCSS))
	}
}
