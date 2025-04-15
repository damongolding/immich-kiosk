package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
}

type Weather struct {
	Coord      Coord  `json:"coord"`
	Data       []Data `json:"weather"`
	Base       string `json:"base"`
	Main       Main   `json:"main"`
	Visibility int    `json:"visibility"`
	Wind       Wind   `json:"wind"`
	Clouds     Clouds `json:"clouds"`
	Dt         int    `json:"dt"`
	Sys        Sys    `json:"sys"`
	Timezone   int    `json:"timezone"`
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Cod        int    `json:"cod"`
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

// AddWeatherLocation adds a new weather location to be monitored.
// It takes a context.Context for cancellation and a config.WeatherLocation struct to configure the monitoring.
// The weather data is fetched immediately and then updated every 10 minutes until the context is cancelled.
// If the location is marked as default and no default exists yet, it will be set as the default location.
func AddWeatherLocation(ctx context.Context, location config.WeatherLocation) {

	if location.Default && DefaultLocation() == "" {
		SetDefaultLocation(location.Name)
		log.Info("Set default weather location", "name", location.Name)
	}

	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()

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

	for {
		select {
		case <-ctx.Done():
			log.Debug("Stopping weather updates for", "name", w.Name)
			return
		case <-ticker.C:
			log.Debug("Getting weather for", "name", w.Name)
			newWeather, newWeatherErr := w.updateWeather(ctx)
			if newWeatherErr != nil {
				log.Error("Failed to update weather", "name", w.Name, "error", newWeatherErr)
				continue
			}
			weatherDataStore.Store(strings.ToLower(w.Name), newWeather)
			log.Debug("Retrieved weather for", "name", w.Name)
		}
	}
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

// updateWeather fetches new weather data from the OpenWeatherMap API for this location.
// Returns the updated WeatherLocation and any error that occurred.
func (w *Location) updateWeather(ctx context.Context) (Location, error) {

	apiURL := url.URL{
		Scheme:   "https",
		Host:     "api.openweathermap.org",
		Path:     "data/2.5/weather",
		RawQuery: fmt.Sprintf("appid=%s&lat=%s&lon=%s&units=%s&lang=%s", w.API, w.Lat, w.Lon, w.Unit, w.Lang),
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		log.Error(err)
		return *w, err
	}

	req.Header.Add("Accept", "application/json")

	var res *http.Response
	for attempts := range 3 {
		res, err = client.Do(req)
		if err == nil {
			break
		}
		log.Error("Request failed, retrying", "attempt", attempts, "URL", apiURL, "err", err)
		time.Sleep(time.Duration(1<<attempts) * time.Second)

	}
	if err != nil {
		log.Error("Request failed after retries", "err", err)
		return *w, err
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		err = fmt.Errorf("unexpected status code: %d", res.StatusCode)
		log.Error(err)
		return *w, err
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error("reading response body", "url", apiURL, "err", err)
		return *w, err
	}

	var newWeather Weather

	unmarshalErr := json.Unmarshal(responseBody, &newWeather)
	if unmarshalErr != nil {
		log.Error("updateWeather", "err", unmarshalErr)
	}

	w.Weather = newWeather

	return *w, nil
}
