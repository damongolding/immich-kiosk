package routes

import (
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/views"
	"github.com/damongolding/immich-kiosk/weather"
	"github.com/labstack/echo/v4"
)

func Weather(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		requestID := requestData.RequestID

		weatherLocation := c.QueryParam("weather")

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"location", weatherLocation,
		)

		if weatherLocation == "" {
			if !baseConfig.HasWeatherDefault {
				log.Warn("No weather location provided and no default set")
				return c.NoContent(http.StatusNoContent)
			}
			for _, loc := range baseConfig.WeatherLocations {
				if loc.Default {
					weatherLocation = loc.Name
					break
				}
			}
			log.Debug("Using default weather location", "location", weatherLocation)
		}

		weatherData := weather.CurrentWeather(weatherLocation)
		if !strings.EqualFold(weatherData.Name, weatherLocation) || len(weatherData.Data) == 0 {
			log.Error("missing weather location data", "location", weatherData.Name)
			return c.NoContent(http.StatusNoContent)
		}

		return Render(c, http.StatusOK, views.Weather(weatherData))

	}
}
