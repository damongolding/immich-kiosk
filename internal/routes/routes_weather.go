package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/weather"
	"github.com/labstack/echo/v4"
)

func Weather(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
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

		var weatherData weather.WeatherLocation

		for attempts := 0; attempts < maxWeatherRetries; attempts++ {
			weatherData = weather.CurrentWeather(weatherLocation)
			if !strings.EqualFold(weatherData.Name, weatherLocation) || len(weatherData.Data) == 0 {
				log.Warn("weather data fetch attempt failed",
					"attempt", attempts+1,
					"location", weatherLocation)
				time.Sleep(time.Duration(1<<attempts) * time.Second)
				continue
			}
			return Render(c, http.StatusOK, partials.WeatherLocation(weatherData))
		}

		log.Error("failed to fetch weather data after all attempts",
			"location", weatherLocation,
			"received_name", weatherData.Name)
		return c.NoContent(http.StatusNoContent)
	}
}
