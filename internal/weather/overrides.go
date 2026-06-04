package weather

import (
	"net/url"
	"strconv"
)

const (
	WeatherShowHumidityParam         = "weather_show_humidity"
	WeatherShowWindParam             = "weather_show_wind"
	WeatherShowWindDirectionParam    = "weather_show_wind_direction"
	WeatherShowVisibilityParam       = "weather_show_visibility"
	WeatherShowTemperatureRangeParam = "weather_show_temperature_range"
	WeatherShowForecastParam         = "weather_show_forecast"
	WeatherRoundTemperatureParam     = "weather_round_temperature"
)

// ApplyDisplayOverrides applies per-request weather display options without
// changing the stored weather data or the global configuration.
func ApplyDisplayOverrides(location Location, values url.Values) Location {
	applyBool := func(key string, field *bool, canEnable bool) {
		param := values.Get(key)
		if param == "" {
			return
		}

		value, err := strconv.ParseBool(param)
		if err != nil {
			return
		}
		if value && !canEnable {
			return
		}

		*field = value
	}

	forecastAvailable := len(location.Forecast.Daily) > 0
	applyBool(WeatherShowHumidityParam, &location.Show.Humidity, true)
	applyBool(WeatherShowWindParam, &location.Show.Wind, true)
	applyBool(WeatherShowWindDirectionParam, &location.Show.WindDirection, true)
	applyBool(WeatherShowVisibilityParam, &location.Show.Visibility, true)
	applyBool(WeatherShowTemperatureRangeParam, &location.Show.TemperatureRange, forecastAvailable)
	applyBool(WeatherShowForecastParam, &location.ShowForecast, forecastAvailable)
	applyBool(WeatherRoundTemperatureParam, &location.RoundTemp, true)

	return location
}
