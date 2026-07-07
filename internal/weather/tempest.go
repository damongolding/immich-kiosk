package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"charm.land/log/v2"
)

// TempestResponse mirrors the relevant parts of the Tempest (WeatherFlow) "better_forecast"
// API response: https://apidocs.tempestwx.com/reference/get_better-forecast
type TempestResponse struct {
	CurrentConditions TempestCurrentConditions `json:"current_conditions"`
	Forecast          TempestForecast          `json:"forecast"`
	Timezone          string                   `json:"timezone"`
}

type TempestCurrentConditions struct {
	Time             int64   `json:"time"`
	Conditions       string  `json:"conditions"`
	Icon             string  `json:"icon"`
	AirTemperature   float64 `json:"air_temperature"`
	RelativeHumidity int     `json:"relative_humidity"`
	StationPressure  float64 `json:"station_pressure"`
	WindAvg          float64 `json:"wind_avg"`
	WindGust         float64 `json:"wind_gust"`
	WindDirection    int     `json:"wind_direction"`
}

type TempestForecast struct {
	Daily  []TempestForecastDaily  `json:"daily"`
	Hourly []TempestForecastHourly `json:"hourly"`
}

type TempestForecastDaily struct {
	DayStartLocal int64   `json:"day_start_local"`
	Conditions    string  `json:"conditions"`
	Icon          string  `json:"icon"`
	AirTempHigh   float64 `json:"air_temp_high"`
	AirTempLow    float64 `json:"air_temp_low"`
}

type TempestForecastHourly struct {
	Time           int64   `json:"time"`
	AirTemperature float64 `json:"air_temperature"`
}

// tempestAPIScheme and tempestAPIHost point at the Tempest REST API. They are
// declared as vars (rather than consts) so tests can redirect requests to a
// local httptest.Server.
var (
	tempestAPIScheme = "https"
	tempestAPIHost   = "swd.weatherflow.com"
)

// fetchTempestData fetches current conditions and forecast data from the Tempest
// "better_forecast" API for this location's station.
func (w *Location) fetchTempestData(ctx context.Context) (*TempestResponse, error) {
	apiURL := url.URL{
		Scheme: tempestAPIScheme,
		Host:   tempestAPIHost,
		Path:   "swd/rest/better_forecast",
	}

	unitsTemp, unitsWind, unitsPressure, unitsPrecip, unitsDistance := "c", "mps", "mb", "mm", "km"
	if strings.EqualFold(w.Unit, ImperialSystem) {
		unitsTemp, unitsWind, unitsPressure, unitsPrecip, unitsDistance = "f", "mph", "inhg", "in", "mi"
	}

	q := url.Values{}
	q.Set("station_id", w.StationID)
	q.Set("token", w.API)
	q.Set("units_temp", unitsTemp)
	q.Set("units_wind", unitsWind)
	q.Set("units_pressure", unitsPressure)
	q.Set("units_precip", unitsPrecip)
	q.Set("units_distance", unitsDistance)

	apiURL.RawQuery = q.Encode()

	// Prepare a redacted URL for logging (avoid leaking the access token)
	apiURLForLog := apiURL
	qLog := apiURLForLog.Query()
	qLog.Set("token", "REDACTED")
	apiURLForLog.RawQuery = qLog.Encode()

	client := httpClient
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	req.Header.Add("Accept", "application/json")

	var res *http.Response
	for attempt := range 3 {
		res, err = client.Do(req)
		if err == nil {
			break
		}
		// Log attempts as 1-based for clarity
		log.Error("Request failed, retrying", "attempt", attempt+1, "url", apiURLForLog.String(), "err", err)

		backoff := time.Duration(1<<attempt) * time.Second
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}

	if err != nil {
		log.Error("Request failed after retries", "url", apiURLForLog.String(), "err", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		bodyPreview, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		err = fmt.Errorf("unexpected status code: %d, body: %s",
			res.StatusCode, strings.TrimSpace(string(bodyPreview)))
		log.Error("Tempest API error",
			"url", apiURLForLog.String(),
			"status", res.StatusCode,
			"body", string(bodyPreview))
		return nil, err
	}

	var result TempestResponse
	if decErr := json.NewDecoder(res.Body).Decode(&result); decErr != nil {
		log.Error("fetchTempestData", "err", decErr)
		return nil, decErr
	}

	return &result, nil
}

// tempestIconToOWMID translates a Tempest forecast icon string into the equivalent
// OpenWeatherMap numeric condition-code bucket used by weatherIcon in weather.templ.
func tempestIconToOWMID(icon string) int {
	switch icon {
	case "cloudy", "partly-cloudy-day", "partly-cloudy-night":
		return 804
	case "foggy":
		return 741
	case "possibly-rainy-day", "possibly-rainy-night", "rainy":
		return 500
	case "possibly-sleet-day", "possibly-sleet-night", "sleet":
		return 611
	case "possibly-snow-day", "possibly-snow-night", "snow":
		return 600
	case "possibly-thunderstorm-day", "possibly-thunderstorm-night", "thunderstorm":
		return 200
	case "windy":
		return 771
	case "clear-day", "clear-night":
		fallthrough
	default:
		return 800
	}
}

// mapTempestCurrent converts a Tempest API response's current conditions into the
// shared Weather struct consumed by the weather templates.
func mapTempestCurrent(tr *TempestResponse) Weather {
	cc := tr.CurrentConditions
	return Weather{
		Data: []Data{
			{
				Description: cc.Conditions,
				ID:          tempestIconToOWMID(cc.Icon),
			},
		},
		Main: Main{
			Temp:     cc.AirTemperature,
			Humidity: cc.RelativeHumidity,
			Pressure: int(cc.StationPressure),
		},
		Wind: Wind{
			Speed: cc.WindAvg,
			Deg:   cc.WindDirection,
			Gust:  cc.WindGust,
		},
		DT: cc.Time,
	}
}

// mapTempestForecast converts a Tempest API response's forecast into the shared
// ForecastData struct consumed by the weather templates.
func mapTempestForecast(tr *TempestResponse) ForecastData {
	n := min(3, len(tr.Forecast.Daily))
	daily := make([]DailySummary, 0, n)
	for _, d := range tr.Forecast.Daily[:n] {
		date := time.Unix(d.DayStartLocal, 0)
		daily = append(daily, DailySummary{
			Date:        date,
			DateStr:     date.Format("2006-01-02"),
			MaxTemp:     d.AirTempHigh,
			WeatherIcon: tempestIconToOWMID(d.Icon),
		})
	}

	high, low := computeTempestNext24hTempRange(tr.Forecast.Hourly)

	return ForecastData{
		Daily:       daily,
		Next24hHigh: high,
		Next24hLow:  low,
	}
}

// computeTempestNext24hTempRange scans the next 24 hours of hourly forecast entries
// and returns the highest and lowest air temperature found.
func computeTempestNext24hTempRange(hourly []TempestForecastHourly) (float64, float64) {
	now := time.Now()
	cutoff := now.Add(24 * time.Hour)

	var high, low float64
	initialized := false
	for _, item := range hourly {
		itemTime := time.Unix(item.Time, 0)
		if itemTime.Before(now) || itemTime.After(cutoff) {
			continue
		}
		if !initialized {
			high = item.AirTemperature
			low = item.AirTemperature
			initialized = true
		} else {
			if item.AirTemperature > high {
				high = item.AirTemperature
			}
			if item.AirTemperature < low {
				low = item.AirTemperature
			}
		}
	}
	return high, low
}
