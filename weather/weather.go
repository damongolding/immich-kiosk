package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
)

var weatherDataStore sync.Map

type WeatherLocation struct {
	Name string
	Lat  string
	Lon  string
	API  string
	Unit string
	Lang string
	Weather
}

type Weather struct {
	Coord      Coord         `json:"coord"`
	Data       []WeatherData `json:"weather"`
	Base       string        `json:"base"`
	Main       Main          `json:"main"`
	Visibility int           `json:"visibility"`
	Wind       Wind          `json:"wind"`
	Clouds     Clouds        `json:"clouds"`
	Dt         int           `json:"dt"`
	Sys        Sys           `json:"sys"`
	Timezone   int           `json:"timezone"`
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	Cod        int           `json:"cod"`
}

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type WeatherData struct {
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

func AddWeatherLocation(ctx context.Context, location config.WeatherLocation) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()

	w := &WeatherLocation{
		Name: location.Name,
		Lat:  location.Lat,
		Lon:  location.Lon,
		API:  location.API,
		Unit: location.Unit,
		Lang: location.Lang,
	}

	weatherDataStore.Store(w.Name, *w)

	// Run once immediately
	log.Debug("Getting initial weather for", "name", w.Name)
	newWeather, err := w.updateWeather()
	if err != nil {
		log.Error("Failed to update initial weather", "name", w.Name, "error", err)
	} else {
		weatherDataStore.Store(w.Name, newWeather)
		log.Debug("Retrieved initial weather for", "name", w.Name)
	}

	for {
		select {
		case <-ticker.C:
			log.Debug("Getting weather for", "name", w.Name)
			newWeather, err := w.updateWeather()
			if err != nil {
				log.Error("Failed to update weather", "name", w.Name, "error", err)
				continue
			}
			weatherDataStore.Store(w.Name, newWeather)
			log.Debug("Retrieved weather for", "name", w.Name)

		case <-ctx.Done():
			log.Debug("Stopping weather updates for", "name", w.Name)
			return
		}
	}
}

func Current(name string) WeatherLocation {
	value, ok := weatherDataStore.Load(name)
	if !ok {
		return WeatherLocation{}
	}
	return value.(WeatherLocation)
}

func (w *WeatherLocation) updateWeather() (WeatherLocation, error) {

	apiUrl := url.URL{
		Scheme:   "https",
		Host:     "api.openweathermap.org",
		Path:     "data/2.5/weather",
		RawQuery: fmt.Sprintf("appid=%s&lat=%s&lon=%s&units=%s&lang=%s", w.API, w.Lat, w.Lon, w.Unit, w.Lang),
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", apiUrl.String(), nil)
	if err != nil {
		log.Error(err)
		return *w, err
	}

	req.Header.Add("Accept", "application/json")

	var res *http.Response
	for attempts := 0; attempts < 3; attempts++ {
		res, err = client.Do(req)
		if err == nil {
			break
		}
		log.Error("Request failed, retrying", "attempt", attempts, "URL", apiUrl, "err", err)
		time.Sleep(time.Duration(attempts) * time.Second)
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
		log.Error("reading response body", "url", apiUrl, "err", err)
		return *w, err
	}

	var newWeather Weather

	if err := json.Unmarshal(responseBody, &newWeather); err != nil {
		log.Error(err)
	}

	w.Weather = newWeather

	return *w, nil
}
