package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/charmbracelet/log"
)

var weatherDataStore []WeatherLocation

type WeatherLocation struct {
	Name    string
	Lat     string
	Lon     string
	API     string
	Weather Weather
}

type Weather struct {
	Coord      Coord         `json:"coord"`
	Weather    []WeatherData `json:"weather"`
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

func (w *WeatherLocation) Current(name string) {

}

func (w *WeatherLocation) updateWeather() {

	apiUrl := url.URL{
		Scheme:   "https",
		Host:     "api.openweathermap.org",
		Path:     "data/2.5/weather",
		RawQuery: fmt.Sprintf("appid=%s&lat=%s&lon=%s", w.API, w.Lat, w.Lon),
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", apiUrl.String(), nil)
	if err != nil {
		log.Error(err)
		return
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
		return
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		err = fmt.Errorf("unexpected status code: %d", res.StatusCode)
		log.Error(err)
		return
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error("reading response body", "url", apiUrl, "err", err)
		return
	}

	var newWeather Weather

	if err := json.Unmarshal(responseBody, &newWeather); err != nil {
		log.Error(err)
	}

	w.Weather = newWeather

}
