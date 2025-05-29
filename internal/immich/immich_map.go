package immich

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
)

type MapReverseGeocodeResponse struct {
	City    string `json:"city"`
	Country string `json:"country"`
	State   string `json:"state"`
}

// LocationFromLatLong performs a reverse geocoding lookup using the provided latitude and longitude.
// It returns the city, state, and country corresponding to the coordinates.
// If the lookup fails or no location is found, empty strings are returned for each value.
func (a *Asset) LocationFromLatLong(lat, long float64) (string, string, string) {

	var city, state, country string

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		log.Error(err)
		return city, state, country
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/map/reverse-geocode",
	}

	apiURL.Query().Add("lat", fmt.Sprintf("%f", lat))
	apiURL.Query().Add("lon", fmt.Sprintf("%f", long))

	apiBody, err := a.immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		log.Error(err)
		return city, state, country
	}

	mapReverseGeocode := make([]MapReverseGeocodeResponse, 0)

	err = json.Unmarshal(apiBody, &mapReverseGeocode)
	if err != nil {
		log.Error(err)
		return city, state, country
	}

	if len(mapReverseGeocode) == 0 {
		log.Error("no location found for", "latitude", lat, "longitude", long)
		return city, state, country
	}

	return mapReverseGeocode[0].City, mapReverseGeocode[0].State, mapReverseGeocode[0].Country

}
