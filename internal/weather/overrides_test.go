package weather

import (
	"net/url"
	"testing"

	"github.com/damongolding/immich-kiosk/internal/config"
)

func TestApplyDisplayOverrides(t *testing.T) {
	location := Location{
		Show: config.WeatherLocationStatOptions{
			Humidity:         false,
			Wind:             false,
			WindDirection:    false,
			Visibility:       false,
			TemperatureRange: false,
		},
		ShowForecast: false,
		RoundTemp:    false,
		Forecast: ForecastData{
			Daily: []DailySummary{{DateStr: "2026-06-04"}},
		},
	}

	location = ApplyDisplayOverrides(location, url.Values{
		WeatherShowHumidityParam:         []string{"true"},
		WeatherShowWindParam:             []string{"true"},
		WeatherShowWindDirectionParam:    []string{"true"},
		WeatherShowVisibilityParam:       []string{"true"},
		WeatherShowTemperatureRangeParam: []string{"true"},
		WeatherShowForecastParam:         []string{"true"},
		WeatherRoundTemperatureParam:     []string{"true"},
	})

	if !location.Show.Humidity {
		t.Fatal("expected humidity to be enabled")
	}
	if !location.Show.Wind {
		t.Fatal("expected wind to be enabled")
	}
	if !location.Show.WindDirection {
		t.Fatal("expected wind direction to be enabled")
	}
	if !location.Show.Visibility {
		t.Fatal("expected visibility to be enabled")
	}
	if !location.Show.TemperatureRange {
		t.Fatal("expected temperature range to be enabled")
	}
	if !location.ShowForecast {
		t.Fatal("expected forecast to be enabled")
	}
	if !location.RoundTemp {
		t.Fatal("expected round temperature to be enabled")
	}
}

func TestApplyDisplayOverridesOnlyChangesProvidedValues(t *testing.T) {
	location := Location{
		Show: config.WeatherLocationStatOptions{
			Humidity:         true,
			Wind:             true,
			WindDirection:    true,
			Visibility:       true,
			TemperatureRange: true,
		},
		ShowForecast: true,
		RoundTemp:    true,
	}

	location = ApplyDisplayOverrides(location, url.Values{
		WeatherShowWindParam: []string{"false"},
		"unknown":            []string{"false"},
	})

	if !location.Show.Humidity {
		t.Fatal("expected humidity to keep its configured value")
	}
	if location.Show.Wind {
		t.Fatal("expected wind to be disabled")
	}
	if !location.Show.WindDirection {
		t.Fatal("expected wind direction to keep its configured value")
	}
	if !location.Show.Visibility {
		t.Fatal("expected visibility to keep its configured value")
	}
	if !location.Show.TemperatureRange {
		t.Fatal("expected temperature range to keep its configured value")
	}
	if !location.ShowForecast {
		t.Fatal("expected forecast to keep its configured value")
	}
	if !location.RoundTemp {
		t.Fatal("expected round temperature to keep its configured value")
	}
}

func TestApplyDisplayOverridesCanDisableConfiguredValues(t *testing.T) {
	location := Location{
		Show: config.WeatherLocationStatOptions{
			Humidity:         true,
			Wind:             true,
			WindDirection:    true,
			Visibility:       true,
			TemperatureRange: true,
		},
		ShowForecast: true,
		RoundTemp:    true,
	}

	location = ApplyDisplayOverrides(location, url.Values{
		WeatherShowHumidityParam:         []string{"false"},
		WeatherShowWindParam:             []string{"false"},
		WeatherShowWindDirectionParam:    []string{"false"},
		WeatherShowVisibilityParam:       []string{"false"},
		WeatherShowTemperatureRangeParam: []string{"false"},
		WeatherShowForecastParam:         []string{"false"},
		WeatherRoundTemperatureParam:     []string{"false"},
	})

	if location.Show.Humidity {
		t.Fatal("expected humidity to be disabled")
	}
	if location.Show.Wind {
		t.Fatal("expected wind to be disabled")
	}
	if location.Show.WindDirection {
		t.Fatal("expected wind direction to be disabled")
	}
	if location.Show.Visibility {
		t.Fatal("expected visibility to be disabled")
	}
	if location.Show.TemperatureRange {
		t.Fatal("expected temperature range to be disabled")
	}
	if location.ShowForecast {
		t.Fatal("expected forecast to be disabled")
	}
	if location.RoundTemp {
		t.Fatal("expected round temperature to be disabled")
	}
}

func TestApplyDisplayOverridesDoesNotEnableUnavailableForecastData(t *testing.T) {
	location := Location{
		Show: config.WeatherLocationStatOptions{
			TemperatureRange: false,
		},
		ShowForecast: false,
	}

	location = ApplyDisplayOverrides(location, url.Values{
		WeatherShowTemperatureRangeParam: []string{"true"},
		WeatherShowForecastParam:         []string{"true"},
	})

	if location.Show.TemperatureRange {
		t.Fatal("expected temperature range to remain disabled without forecast data")
	}
	if location.ShowForecast {
		t.Fatal("expected forecast to remain disabled without forecast data")
	}
}

func TestApplyDisplayOverridesIgnoresInvalidValues(t *testing.T) {
	location := Location{
		Show: config.WeatherLocationStatOptions{
			Humidity: true,
		},
	}

	location = ApplyDisplayOverrides(location, url.Values{
		WeatherShowHumidityParam: []string{"maybe"},
	})

	if !location.Show.Humidity {
		t.Fatal("expected invalid values to be ignored")
	}
}
