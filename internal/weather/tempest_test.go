package weather

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// withTempestTestServer points the package's Tempest API host/scheme at a local
// httptest.Server for the duration of fn, restoring the originals afterwards.
func withTempestTestServer(t *testing.T, srv *httptest.Server, fn func()) {
	t.Helper()

	parsed, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}

	origScheme, origHost := tempestAPIScheme, tempestAPIHost
	tempestAPIScheme, tempestAPIHost = parsed.Scheme, parsed.Host
	defer func() { tempestAPIScheme, tempestAPIHost = origScheme, origHost }()

	fn()
}

func TestTempestIconToOWMID(t *testing.T) {
	tests := []struct {
		icon string
		want int
	}{
		{"clear-day", 800},
		{"clear-night", 800},
		{"cloudy", 804},
		{"partly-cloudy-day", 804},
		{"partly-cloudy-night", 804},
		{"foggy", 741},
		{"possibly-rainy-day", 500},
		{"possibly-rainy-night", 500},
		{"rainy", 500},
		{"possibly-sleet-day", 611},
		{"possibly-sleet-night", 611},
		{"sleet", 611},
		{"possibly-snow-day", 600},
		{"possibly-snow-night", 600},
		{"snow", 600},
		{"possibly-thunderstorm-day", 200},
		{"possibly-thunderstorm-night", 200},
		{"thunderstorm", 200},
		{"windy", 771},
		{"some-unknown-icon", 800},
		{"", 800},
	}

	for _, tc := range tests {
		t.Run(tc.icon, func(t *testing.T) {
			if got := tempestIconToOWMID(tc.icon); got != tc.want {
				t.Errorf("tempestIconToOWMID(%q) = %d, want %d", tc.icon, got, tc.want)
			}
		})
	}
}

func TestMapTempestCurrent(t *testing.T) {
	tr := &TempestResponse{
		CurrentConditions: TempestCurrentConditions{
			Time:             1700000000,
			Conditions:       "Partly Cloudy",
			Icon:             "partly-cloudy-day",
			AirTemperature:   21.5,
			RelativeHumidity: 55,
			StationPressure:  1013.2,
			WindAvg:          3.4,
			WindGust:         5.1,
			WindDirection:    270,
		},
	}

	got := mapTempestCurrent(tr)

	if len(got.Data) != 1 {
		t.Fatalf("expected 1 Data entry, got %d", len(got.Data))
	}
	if got.Data[0].Description != "Partly Cloudy" {
		t.Errorf("Description = %q, want %q", got.Data[0].Description, "Partly Cloudy")
	}
	if got.Data[0].ID != 804 {
		t.Errorf("ID = %d, want %d", got.Data[0].ID, 804)
	}
	if got.Main.Temp != 21.5 {
		t.Errorf("Temp = %v, want %v", got.Main.Temp, 21.5)
	}
	if got.Main.Humidity != 55 {
		t.Errorf("Humidity = %d, want %d", got.Main.Humidity, 55)
	}
	if got.Wind.Speed != 3.4 {
		t.Errorf("Wind.Speed = %v, want %v", got.Wind.Speed, 3.4)
	}
	if got.Wind.Deg != 270 {
		t.Errorf("Wind.Deg = %d, want %d", got.Wind.Deg, 270)
	}
	if got.DT != 1700000000 {
		t.Errorf("DT = %d, want %d", got.DT, 1700000000)
	}
}

func TestMapTempestForecast(t *testing.T) {
	now := time.Now()
	hour := func(offset float64) int64 {
		return now.Add(time.Duration(offset * float64(time.Hour))).Unix()
	}

	tr := &TempestResponse{
		Forecast: TempestForecast{
			Daily: []TempestForecastDaily{
				{DayStartLocal: hour(0), Conditions: "Clear", Icon: "clear-day", AirTempHigh: 25, AirTempLow: 12},
				{DayStartLocal: hour(24), Conditions: "Rain", Icon: "rainy", AirTempHigh: 18, AirTempLow: 10},
				{DayStartLocal: hour(48), Conditions: "Snow", Icon: "snow", AirTempHigh: 2, AirTempLow: -4},
				{DayStartLocal: hour(72), Conditions: "Windy", Icon: "windy", AirTempHigh: 15, AirTempLow: 5},
			},
			Hourly: []TempestForecastHourly{
				{Time: hour(-3), AirTemperature: 99}, // in the past, ignored
				{Time: hour(1), AirTemperature: 20},
				{Time: hour(6), AirTemperature: 25},
				{Time: hour(12), AirTemperature: 8},
				{Time: hour(25), AirTemperature: -99}, // beyond 24h, ignored
			},
		},
	}

	got := mapTempestForecast(tr)

	if len(got.Daily) != 3 {
		t.Fatalf("expected 3 daily summaries (capped), got %d", len(got.Daily))
	}
	if got.Daily[0].MaxTemp != 25 || got.Daily[0].WeatherIcon != 800 {
		t.Errorf("Daily[0] = %+v, want MaxTemp=25 WeatherIcon=800", got.Daily[0])
	}
	if got.Daily[1].MaxTemp != 18 || got.Daily[1].WeatherIcon != 500 {
		t.Errorf("Daily[1] = %+v, want MaxTemp=18 WeatherIcon=500", got.Daily[1])
	}
	if got.Daily[2].MaxTemp != 2 || got.Daily[2].WeatherIcon != 600 {
		t.Errorf("Daily[2] = %+v, want MaxTemp=2 WeatherIcon=600", got.Daily[2])
	}
	if got.Next24hHigh != 25 {
		t.Errorf("Next24hHigh = %v, want %v", got.Next24hHigh, 25)
	}
	if got.Next24hLow != 8 {
		t.Errorf("Next24hLow = %v, want %v", got.Next24hLow, 8)
	}
}

func TestFetchTempestData(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("station_id") != "1234" {
				t.Errorf("station_id = %q, want %q", r.URL.Query().Get("station_id"), "1234")
			}
			if r.URL.Query().Get("token") != "secret-token" {
				t.Errorf("token = %q, want %q", r.URL.Query().Get("token"), "secret-token")
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(TempestResponse{
				CurrentConditions: TempestCurrentConditions{Conditions: "Clear", Icon: "clear-day", AirTemperature: 18},
			})
		}))
		defer srv.Close()

		withTempestTestServer(t, srv, func() {
			loc := &Location{StationID: "1234", API: "secret-token", Unit: MetricSystem}

			got, err := loc.fetchTempestData(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.CurrentConditions.Conditions != "Clear" {
				t.Errorf("Conditions = %q, want %q", got.CurrentConditions.Conditions, "Clear")
			}
		})
	})

	t.Run("non-200 status returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"status":{"status_code":401}}`))
		}))
		defer srv.Close()

		withTempestTestServer(t, srv, func() {
			loc := &Location{StationID: "1234", API: "bad-token", Unit: MetricSystem}

			if _, err := loc.fetchTempestData(context.Background()); err == nil {
				t.Fatal("expected error for non-200 status, got nil")
			}
		})
	})
}
