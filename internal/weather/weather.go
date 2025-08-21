package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
)

const (
	MetricSystem   = "metric"
	ImperialSystem = "imperial"
	APINameKeyword = "-api"
)

var (
	weatherDataStore  sync.Map
	defaultLocationMu sync.RWMutex
	defaultLocation   string
)

type Location struct {
	Name string
	Lat  string
	Lon  string
	API  string
	Unit string
	Lang string
	Weather
	Forecast []DailySummary
}

type Weather struct {
	Coord      Coord  `json:"coord"`
	Data       []Data `json:"weather"`
	Base       string `json:"base"`
	Main       Main   `json:"main"`
	Visibility int    `json:"visibility"`
	Wind       Wind   `json:"wind"`
	Clouds     Clouds `json:"clouds"`
	DT         int64  `json:"dt"`
	Sys        Sys    `json:"sys"`
	Timezone   int    `json:"timezone"`
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Cod        int    `json:"cod"`
}

type Forecast struct {
	List []Weather `json:"list"`
}

type DailySummary struct {
	Date        time.Time
	DateStr     string
	MaxTemp     float64
	WeatherIcon int
}

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type Data struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	Humidity  int     `json:"humidity"`
	SeaLevel  int     `json:"sea_level"`
	GrndLevel int     `json:"grnd_level"`
}

type Wind struct {
	Speed float64 `json:"speed"`
	Deg   int     `json:"deg"`
	Gust  float64 `json:"gust"`
}

type Clouds struct {
	All int `json:"all"`
}

type Sys struct {
	Type    int    `json:"type"`
	ID      int    `json:"id"`
	Country string `json:"country"`
	Sunrise int    `json:"sunrise"`
	Sunset  int    `json:"sunset"`
}

func DefaultLocation() string {
	defaultLocationMu.RLock()
	defer defaultLocationMu.RUnlock()
	return defaultLocation
}

func SetDefaultLocation(location string) {
	defaultLocationMu.Lock()
	defer defaultLocationMu.Unlock()
	defaultLocation = location
}

// addWeatherLocation is the internal worker that manages a single location.
// It fetches initial data, then updates weather every 10 minutes and, if enabled, forecast on its own ticker.
func addWeatherLocation(ctx context.Context, location config.WeatherLocation, withForecast bool) {
	if location.Default && DefaultLocation() == "" {
		SetDefaultLocation(location.Name)
		log.Info("Set default weather location", "name", location.Name)
	}

	weatherTicker := time.NewTicker(time.Minute * 10)
	defer weatherTicker.Stop()

	var forecastTicker *time.Ticker
	if withForecast {
		forecastTicker = time.NewTicker(time.Hour * 3)
		defer forecastTicker.Stop()
	}

	w := &Location{
		Name: location.Name,
		Lat:  location.Lat,
		Lon:  location.Lon,
		API:  location.API,
		Unit: location.Unit,
		Lang: location.Lang,
	}

	weatherDataStore.Store(strings.ToLower(w.Name), *w)

	// Run once immediately
	log.Debug("Getting initial weather for", "name", w.Name)
	newWeatherInit, newWeatherInitErr := w.updateWeather(ctx)
	if newWeatherInitErr != nil {
		log.Error("Failed to update initial weather", "name", w.Name, "error", newWeatherInitErr)
	} else {
		weatherDataStore.Store(strings.ToLower(w.Name), newWeatherInit)
		log.Debug("Retrieved initial weather for", "name", w.Name)
	}

	if withForecast {
		// Run once immediately
		log.Debug("Getting initial forecast for", "name", w.Name)
		newForecastInit, newForecastInitErr := w.updateForecast(ctx)
		if newForecastInitErr != nil {
			log.Error("Failed to update initial forecast", "name", w.Name, "error", newForecastInitErr)
		} else {
			weatherDataStore.Store(strings.ToLower(w.Name), newForecastInit)
			log.Debug("Retrieved initial forecast for", "name", w.Name)
		}
	}

	var forecastCh <-chan time.Time
	if withForecast && forecastTicker != nil {
		forecastCh = forecastTicker.C
	}
	for {
		select {
		case <-ctx.Done():
			log.Debug("Stopping weather updates for", "name", w.Name)
			return
		case <-weatherTicker.C:
			log.Debug("Getting weather for", "name", w.Name)
			if newWeather, err := w.updateWeather(ctx); err != nil {
				log.Error("Failed to update weather", "name", w.Name, "error", err)
			} else {
				weatherDataStore.Store(strings.ToLower(w.Name), newWeather)
				log.Debug("Retrieved weather for", "name", w.Name)
			}
		case <-forecastCh:
			log.Debug("Getting forecast for", "name", w.Name)
			if newForecast, err := w.updateForecast(ctx); err != nil {
				log.Error("Failed to update forecast", "name", w.Name, "error", err)
			} else {
				weatherDataStore.Store(strings.ToLower(w.Name), newForecast)
				log.Debug("Retrieved forecast for", "name", w.Name)
			}
		}
	}
}

// AddWeatherLocation adds a new weather-only location (no forecast).
func AddWeatherLocation(ctx context.Context, location config.WeatherLocation) {
	addWeatherLocation(ctx, location, false)
}

// AddWeatherLocationWithForecast adds a new location and enables periodic forecast updates.
func AddWeatherLocationWithForecast(ctx context.Context, location config.WeatherLocation) {
	addWeatherLocation(ctx, location, true)
}

// CurrentWeather retrieves the current weather data for a given location name.
// Returns a WeatherLocation struct containing the weather data, or an empty struct if not found.
func CurrentWeather(name string) Location {
	value, ok := weatherDataStore.Load(strings.ToLower(name))
	if !ok {
		return Location{}
	}
	loc, ok := value.(Location)
	if !ok {
		return Location{}
	}
	return loc
}

// fetchWeatherData is a generic function to fetch weather or forecast data from the OpenWeatherMap API.
// The 'endpoint' argument should be either "weather" or "forecast/daily".
// The 'result' argument should be a pointer to the struct to unmarshal into (Weather or Forecast).
func (w *Location) fetchWeatherData(ctx context.Context, endpoint string, result any) error {
	apiURL := url.URL{
		Scheme: "https",
		Host:   "api.openweathermap.org",
		Path:   fmt.Sprintf("data/2.5/%s", endpoint),
	}

	// Build query string
	q := url.Values{}
	q.Set("appid", w.API)
	q.Set("lat", w.Lat)
	q.Set("lon", w.Lon)
	q.Set("units", w.Unit)
	q.Set("lang", w.Lang)

	apiURL.RawQuery = q.Encode()

	// Prepare a redacted URL for logging (avoid leaking API key)
	apiURLForLog := apiURL
	qLog := apiURLForLog.Query()
	qLog.Set("appid", "REDACTED")
	apiURLForLog.RawQuery = qLog.Encode()

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		log.Error(err)
		return err
	}

	req.Header.Add("Accept", "application/json")

	var res *http.Response
	for attempts := range 3 {
		res, err = client.Do(req)
		if err == nil {
			break
		}
		log.Error("Request failed, retrying", "attempt", attempts, "url", apiURLForLog.String(), "err", err)

		time.Sleep(time.Duration(1<<attempts) * time.Second)
	}
	if err != nil {
		log.Error("Request failed after retries", "err", err)
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		err = fmt.Errorf("unexpected status code: %d", res.StatusCode)
		log.Error(err)
		return err
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error("reading response body", "url", apiURLForLog.String(), "err", err)
		return err
	}

	unmarshalErr := json.Unmarshal(responseBody, result)
	if unmarshalErr != nil {
		log.Error("fetchWeatherData", "err", unmarshalErr)
		return unmarshalErr
	}

	return nil
}

// updateWeather fetches new weather data from the OpenWeatherMap API for this location.
// Returns the updated Location and any error that occurred.
func (w *Location) updateWeather(ctx context.Context) (Location, error) {
	var newWeather Weather
	err := w.fetchWeatherData(ctx, "weather", &newWeather)
	if err != nil {
		return *w, err
	}
	w.Weather = newWeather
	return *w, nil
}

// updateForecast fetches new forecast data from the OpenWeatherMap API for this location.
// Returns the updated Location and any error that occurred.
func (w *Location) updateForecast(ctx context.Context) (Location, error) {
	var newForecast Forecast
	err := w.fetchWeatherData(ctx, "forecast", &newForecast)
	if err != nil {
		return *w, err
	}
	w.Forecast = processForecast(newForecast)
	return *w, nil
}

func processForecast(forecast Forecast) []DailySummary {
	// Today’s date at midnight UTC
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	daily := make(map[string]*DailySummary)
	weatherCounts := make(map[string]map[int]int)

	for _, item := range forecast.List {
		itemTime := time.Unix(item.DT, 0).UTC()
		itemDate := time.Date(itemTime.Year(), itemTime.Month(), itemTime.Day(), 0, 0, 0, 0, time.UTC)

		// Skip today and past
		if !itemDate.After(today) {
			continue
		}

		dateStr := itemDate.Format("2006-01-02")

		// Init if not exists
		if _, ok := daily[dateStr]; !ok {
			daily[dateStr] = &DailySummary{
				Date:        itemDate,
				DateStr:     dateStr,
				MaxTemp:     item.Main.TempMax,
				WeatherIcon: item.Data[0].ID,
			}
			weatherCounts[dateStr] = make(map[int]int)
		}

		// Update max temp
		if item.Main.TempMax > daily[dateStr].MaxTemp {
			daily[dateStr].MaxTemp = item.Main.TempMax
		}

		// Count weather.id
		weatherID := item.Data[0].ID
		weatherCounts[dateStr][weatherID]++

		// Update most common
		maxCount := 0
		mostCommon := daily[dateStr].WeatherIcon
		for w, c := range weatherCounts[dateStr] {
			if c > maxCount {
				maxCount = c
				mostCommon = w
			}
		}
		daily[dateStr].WeatherIcon = mostCommon
	}

	// Sort and print
	var summaries []DailySummary
	for _, v := range daily {
		summaries = append(summaries, *v)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].DateStr < summaries[j].DateStr
	})

	return summaries[:3]

}
