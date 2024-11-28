package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/labstack/echo/v4"
)

func PreLoad(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

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
			"deviceID", deviceID,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		if isSleepMode(requestConfig) {
			return c.NoContent(http.StatusNoContent)
		}

		spoof := "/image"

		if cachedViewData := fromCacheWithURL(spoof, deviceID); cachedViewData != nil {
			// go webhooks.Trigger(requestData, KioskVersion, webhooks.PrefetchAsset, cachedViewData[0])

			log.Info("preload", "ImagePath", cachedViewData[len(cachedViewData)-1].Images[0].ImagePath)

			s := strings.Builder{}

			for _, i := range cachedViewData[len(cachedViewData)-1].Images {
				s.WriteString(fmt.Sprintf("<link rel='preload' as='image' href='/image/%s'>", i.ImagePath))
			}

			return c.HTML(http.StatusOK, fmt.Sprintf(`
				<head>
					%s
				</head>
			`, s.String()))
		}

		log.Info(requestID, "deviceID", deviceID, "cache miss for prefetched image")

		return c.NoContent(http.StatusNoContent)
	}
}
