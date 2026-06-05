package weather

import (
	"net/url"
	"testing"
	"time"

	"github.com/damongolding/immich-kiosk/internal/config"
)

// makeWeather is a test helper for building Forecast.List entries.
func makeWeather(unixSec int64, tempMax, tempMin float64) Weather {
	return Weather{
		DT:   unixSec,
		Main: Main{TempMax: tempMax, TempMin: tempMin},
	}
}

func TestComputeNext24hTempRange(t *testing.T) {
	now := time.Now()

	hour := func(offset float64) int64 {
		return now.Add(time.Duration(offset * float64(time.Hour))).Unix()
	}

	tests := []struct {
		name     string
		forecast Forecast
		wantHigh float64
		wantLow  float64
	}{
		{
			name: "single interval within window",
			forecast: Forecast{List: []Weather{
				makeWeather(hour(1), 20.0, 10.0),
			}},
			wantHigh: 20.0,
			wantLow:  10.0,
		},
		{
			name: "multiple intervals – picks max high and min low",
			forecast: Forecast{List: []Weather{
				makeWeather(hour(1), 18.0, 12.0),
				makeWeather(hour(6), 25.0, 8.0),
				makeWeather(hour(12), 22.0, 10.0),
			}},
			wantHigh: 25.0,
			wantLow:  8.0,
		},
		{
			name: "intervals outside window are ignored",
			forecast: Forecast{List: []Weather{
				makeWeather(hour(-3), 99.0, -99.0), // 3h in the past
				makeWeather(hour(1), 20.0, 10.0),   // within window
				makeWeather(hour(25), 99.0, -99.0), // 25h in the future
			}},
			wantHigh: 20.0,
			wantLow:  10.0,
		},
		{
			name:     "empty forecast returns zero values",
			forecast: Forecast{List: []Weather{}},
			wantHigh: 0,
			wantLow:  0,
		},
		{
			name: "all intervals outside window returns zero values",
			forecast: Forecast{List: []Weather{
				makeWeather(hour(-5), 30.0, 5.0), // past
				makeWeather(hour(25), 28.0, 6.0), // beyond 24h
			}},
			wantHigh: 0,
			wantLow:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			high, low := computeNext24hTempRange(tc.forecast)
			if high != tc.wantHigh {
				t.Errorf("high = %v, want %v", high, tc.wantHigh)
			}
			if low != tc.wantLow {
				t.Errorf("low = %v, want %v", low, tc.wantLow)
			}
		})
	}
}

func TestWindCompassDirection(t *testing.T) {
	tests := []struct {
		name     string
		degree   int
		expected string
	}{
		{"North", 0, "N"},
		{"NorthEast", 45, "NE"},
		{"East", 90, "E"},
		{"SouthEast", 135, "SE"},
		{"South", 180, "S"},
		{"SouthWest", 225, "SW"},
		{"West", 270, "W"},
		{"NorthWest", 315, "NW"},
		{"North 360", 360, "N"},
		{"North wrap", 338, "N"},
		{"NW upper boundary", 337, "NW"},
		{"Negative degree", -45, VarCompassDirection},
		{"Overflow degree", 405, VarCompassDirection},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := Wind{Deg: tt.degree}
			result := w.CompassDirection()
			if result != tt.expected {
				t.Errorf("CompassDirection() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestApplyURLOverrides(t *testing.T) {
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

	location = ApplyURLOverrides(location, url.Values{
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

func TestApplyURLOverridesOnlyChangesProvidedValues(t *testing.T) {
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

	location = ApplyURLOverrides(location, url.Values{
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

func TestApplyURLOverridesCanDisableConfiguredValues(t *testing.T) {
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

	location = ApplyURLOverrides(location, url.Values{
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

func TestApplyURLOverridesIgnoresInvalidValues(t *testing.T) {
	location := Location{
		Show: config.WeatherLocationStatOptions{
			Humidity: true,
		},
	}

	location = ApplyURLOverrides(location, url.Values{
		WeatherShowHumidityParam: []string{"maybe"},
	})

	if !location.Show.Humidity {
		t.Fatal("expected invalid values to be ignored")
	}
}
