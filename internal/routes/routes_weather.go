package routes

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/weather"
	"github.com/labstack/echo/v5"
)

func Weather(baseConfig *config.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
		}

		requestID := requestData.RequestID

		locationName := c.FormValue(weather.WeatherParam)
		nextWeatherRotation := 0

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"location", locationName,
		)

		// handle weather rotation
		if locationName == weather.WeatherRotation {
			currentWeatherRotation, cwrErr := strconv.Atoi(c.FormValue(weather.WeatherRotationParam))
			if cwrErr != nil {
				log.Error("Could not parse weather location position", "error", cwrErr)
				return c.NoContent(http.StatusNoContent)
			}
			nextWeatherRotation, locationName = weather.LocationRotator.Next(currentWeatherRotation)
			log.Info("Rotating weather location", "location", locationName, "position", nextWeatherRotation)
		}

		if locationName == "" {
			if !baseConfig.Weather.HasDefault {
				log.Warn("No weather location provided and no default set")
				return c.NoContent(http.StatusNoContent)
			}
			for _, loc := range baseConfig.Weather.Locations {
				if loc.Default {
					locationName = loc.Name
					break
				}
			}
			log.Debug("Using default weather location", "location", locationName)
		}

		var weatherLocation weather.Location

		for attempts := range maxWeatherRetries {
			weatherLocation = weather.CurrentWeather(locationName)
			if !strings.EqualFold(weatherLocation.Name, locationName) || len(weatherLocation.Data) == 0 {
				log.Warn("weather data fetch attempt failed",
					"attempt", attempts+1,
					"location", locationName)
				time.Sleep(time.Duration(1<<attempts) * time.Second)
				continue
			}
			return Render(c, http.StatusOK, partials.WeatherLocation(weatherLocation, nextWeatherRotation, baseConfig.SystemLang))

		}

		log.Error("failed to fetch weather data after all attempts",
			"location", locationName,
			"received_name", weatherLocation.Name)
		return c.NoContent(http.StatusNoContent)
	}
}
