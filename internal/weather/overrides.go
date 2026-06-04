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
	applyBool := func(key string, setter func(bool)) {
		raw, ok := firstValue(values, key)
		if !ok {
			return
		}

		value, err := strconv.ParseBool(raw)
		if err != nil {
			return
		}

		setter(value)
	}

	applyBool(WeatherShowHumidityParam, func(value bool) {
		location.Show.Humidity = value
	})
	applyBool(WeatherShowWindParam, func(value bool) {
		location.Show.Wind = value
	})
	applyBool(WeatherShowWindDirectionParam, func(value bool) {
		location.Show.WindDirection = value
	})
	applyBool(WeatherShowVisibilityParam, func(value bool) {
		location.Show.Visibility = value
	})
	applyBool(WeatherShowTemperatureRangeParam, func(value bool) {
		location.Show.TemperatureRange = value
	})
	applyBool(WeatherShowForecastParam, func(value bool) {
		location.ShowForecast = value
	})
	applyBool(WeatherRoundTemperatureParam, func(value bool) {
		location.RoundTemp = value
	})

	return location
}

func firstValue(values url.Values, key string) (string, bool) {
	if values == nil {
		return "", false
	}

	raw, ok := values[key]
	if !ok || len(raw) == 0 {
		return "", false
	}

	return raw[0], true
}
