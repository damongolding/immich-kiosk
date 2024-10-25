package routes

import (
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/utils"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/damongolding/immich-kiosk/weather"
	"github.com/labstack/echo/v4"
)

func Weather(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestID := utils.ColorizeRequestId(c.Response().Header().Get(echo.HeaderXRequestID))
		weatherLocation := c.QueryParam("weather")

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"location", weatherLocation,
		)

		if weatherLocation == "" {
			log.Error("missing weather location name url param")
			return c.NoContent(http.StatusOK)
		}

		weatherData := weather.CurrentWeather(weatherLocation)
		if !strings.EqualFold(weatherData.Name, weatherLocation) || len(weatherData.Data) == 0 {
			log.Error("missing weather location data", "location", weatherData.Name)
			return c.NoContent(http.StatusOK)
		}

		return Render(c, http.StatusOK, views.Weather(weatherData))

	}
}
