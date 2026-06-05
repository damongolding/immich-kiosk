package common

import (
	"testing"

	"github.com/google/go-querystring/query"
)

func TestURLBuilderRequestEncodesWeatherFields(t *testing.T) {
	req := URLBuilderRequest{
		Weather:                     ptr("london"),
		WeatherRotationInterval:     ptr(uint64(30)),
		WeatherShowForecast:         ptr(true),
		WeatherShowHumidity:         ptr(false),
		WeatherShowWind:             ptr(true),
		WeatherShowWindDirection:    ptr(false),
		WeatherShowVisibility:       ptr(true),
		WeatherShowTemperatureRange: ptr(false),
		WeatherRoundTemperature:     ptr(true),
	}

	values, err := query.Values(req)
	if err != nil {
		t.Fatalf("encoding URL builder request: %v", err)
	}

	assertQueryValue(t, values.Get("weather"), "london")
	assertQueryValue(t, values.Get("rotation_interval"), "30")
	assertQueryValue(t, values.Get("weather_show_forecast"), "true")
	assertQueryValue(t, values.Get("weather_show_humidity"), "false")
	assertQueryValue(t, values.Get("weather_show_wind"), "true")
	assertQueryValue(t, values.Get("weather_show_wind_direction"), "false")
	assertQueryValue(t, values.Get("weather_show_visibility"), "true")
	assertQueryValue(t, values.Get("weather_show_temperature_range"), "false")
	assertQueryValue(t, values.Get("weather_round_temperature"), "true")
}

func assertQueryValue(t *testing.T, actual string, expected string) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func ptr[T any](value T) *T {
	return &value
}
