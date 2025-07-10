package config

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestConfigWithOverrides testing whether ImmichURL and ImmichApiKey are immutable
func TestImmichURLImmichApiKeyImmutability(t *testing.T) {

	originalURL := "https://my-server.com"
	originalAPI := "123456"
	originalUsersAPIKeys := map[string]string{"default": "123456"}

	c := New()
	c.ImmichURL = originalURL
	c.ImmichAPIKey = originalAPI
	c.ImmichUsersAPIKeys = originalUsersAPIKeys

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	q := req.URL.Query()
	q.Add("immich_url", "https://my-new-server.com")
	q.Add("immich_api_key", "9999")
	q.Add("immich_users_api_keys", "{\"user1\": \"9999\"}")

	req.URL.RawQuery = q.Encode()

	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	err := c.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	assert.Equal(t, originalURL, c.ImmichURL, "ImmichURL field was allowed to be changed")
	assert.Equal(t, originalAPI, c.ImmichAPIKey, "ImmichAPIKey field was allowed to be changed")
	assert.Equal(t, originalUsersAPIKeys, c.ImmichUsersAPIKeys, "ImmichUsersAPIKeys field was allowed to be changed")
}

// TestImmichURLImmichMultiplePerson tests the addition of multiple persons to the config
func TestImmichURLImmichMultiplePerson(t *testing.T) {
	c := New()

	e := echo.New()

	q := make(url.Values)
	q.Add("person", "bea")
	q.Add("person", "laura")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.QueryParams())

	err := c.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	assert.Equal(t, 2, len(c.People), "Expected 2 people to be added")
	assert.Contains(t, c.People, "bea", "Expected 'bea' to be added to Person slice")
	assert.Contains(t, c.People, "laura", "Expected 'laura' to be added to Person slice")
}

// TestMalformedURLs testing urls without scheme or ports
func TestMalformedURLs(t *testing.T) {

	var tests = []struct {
		URL  string
		Want string
	}{
		{URL: "nope", Want: defaultScheme + "nope"},
		{URL: "192.168.1.1", Want: defaultScheme + "192.168.1.1"},
		{URL: "192.168.1.1:1234", Want: defaultScheme + "192.168.1.1:1234"},
		{URL: "https://192.168.1.1:1234", Want: "https://192.168.1.1:1234"},
		{URL: "nope:32", Want: defaultScheme + "nope:32"},
	}

	for _, test := range tests {

		t.Run(test.URL, func(t *testing.T) {
			t.Setenv("KIOSK_IMMICH_URL", test.URL)
			t.Setenv("KIOSK_IMMICH_API_KEY", "12345")

			c := New()

			err := c.Load()
			assert.NoError(t, err, "Config load should not return an error")

			assert.Equal(t, test.Want, c.ImmichURL, "ImmichURL should be formatted correctly")
		})
	}
}

// TestImmichURLImmichMultipleAlbum tests the addition and overriding of multiple albums in the config
func TestImmichURLImmichMultipleAlbum(t *testing.T) {

	// configWithBase
	configWithBase := New()
	configWithBase.Albums = []string{"BASE_ALBUM"}

	e := echo.New()

	q := make(url.Values)
	q.Add("album", "ALBUM_1")
	q.Add("album", "ALBUM_2")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.QueryParams())

	err := configWithBase.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	t.Log("album", configWithBase.Albums)

	assert.NotContains(t, configWithBase.Albums, "BASE_ALBUM", "BASE_ALBUM should not be present")
	assert.Equal(t, 2, len(configWithBase.Albums), "Expected 2 albums to be added")
	assert.Contains(t, configWithBase.Albums, "ALBUM_1", "ALBUM_1 should be present")
	assert.Contains(t, configWithBase.Albums, "ALBUM_2", "ALBUM_2 should be present")

	// configWithoutBase
	configWithoutBase := New()

	q = make(url.Values)
	q.Add("album", "ALBUM_1")
	q.Add("album", "ALBUM_2")

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec = httptest.NewRecorder()

	echoContenx = e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.QueryParams())

	err = configWithoutBase.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	t.Log("album", configWithoutBase.Albums)

	assert.Equal(t, 2, len(configWithoutBase.Albums), "Expected 2 albums to be added")
	assert.Contains(t, configWithoutBase.Albums, "ALBUM_1", "ALBUM_1 should be present")
	assert.Contains(t, configWithoutBase.Albums, "ALBUM_2", "ALBUM_2 should be present")

	// configWithBaseOnly
	configWithBaseOnly := New()
	configWithBaseOnly.Albums = []string{"BASE_ALBUM_1", "BASE_ALBUM_2"}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()

	echoContenx = e.NewContext(req, rec)

	err = configWithBaseOnly.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	t.Log("album", configWithBaseOnly.Albums)

	assert.Equal(t, 2, len(configWithBaseOnly.Albums), "Base albums should persist")
	assert.Contains(t, configWithBaseOnly.Albums, "BASE_ALBUM_1", "BASE_ALBUM_1 should be present")
	assert.Contains(t, configWithBaseOnly.Albums, "BASE_ALBUM_2", "BASE_ALBUM_2 should be present")
}

func TestAlbumAndPerson(t *testing.T) {
	testCases := []struct {
		name           string
		inputAlbum     []string
		inputPerson    []string
		expectedAlbum  []string
		expectedPerson []string
	}{
		{
			name:           "No empty values",
			inputAlbum:     []string{"album1", "album2"},
			inputPerson:    []string{"person1", "person2"},
			expectedAlbum:  []string{"album1", "album2"},
			expectedPerson: []string{"person1", "person2"},
		},
		{
			name:           "Empty values in album",
			inputAlbum:     []string{"album1", "", "album2", ""},
			inputPerson:    []string{"person1", "person2"},
			expectedAlbum:  []string{"album1", "album2"},
			expectedPerson: []string{"person1", "person2"},
		},
		{
			name:           "Empty values in person",
			inputAlbum:     []string{"album1", "album2"},
			inputPerson:    []string{"", "person1", "", "person2"},
			expectedAlbum:  []string{"album1", "album2"},
			expectedPerson: []string{"person1", "person2"},
		},
		{
			name:           "Empty values in both",
			inputAlbum:     []string{"", "album1", "", "album2"},
			inputPerson:    []string{"person1", "", "", "person2"},
			expectedAlbum:  []string{"album1", "album2"},
			expectedPerson: []string{"person1", "person2"},
		},
		{
			name:           "All empty values",
			inputAlbum:     []string{"", "", ""},
			inputPerson:    []string{"", "", ""},
			expectedAlbum:  []string{},
			expectedPerson: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{
				Albums: tc.inputAlbum,
				People: tc.inputPerson,
			}

			c.checkAssetBuckets()

			assert.Equal(t, tc.expectedAlbum, c.Albums, "Album mismatch")
			assert.Equal(t, tc.expectedPerson, c.People, "Person mismatch")
		})
	}
}

func TestCheckWeatherLocations(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "All fields present",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City", Lat: "123", Lon: "456", API: "abc123"},
				},
			},
			expected: "",
		},
		{
			name: "Missing name",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Lat: "123", Lon: "456", API: "abc123"},
				},
			},
			expected: "Weather location is missing required fields: name",
		},
		{
			name: "Missing latitude",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City", Lon: "456", API: "abc123"},
				},
			},
			expected: "Weather location is missing required fields: latitude",
		},
		{
			name: "Missing longitude",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City", Lat: "123", API: "abc123"},
				},
			},
			expected: "Weather location is missing required fields: longitude",
		},
		{
			name: "Missing API key",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City", Lat: "123", Lon: "456"},
				},
			},
			expected: "Weather location is missing required fields: API key",
		},
		{
			name: "Multiple missing fields",
			config: &Config{
				WeatherLocations: []WeatherLocation{
					{Name: "City"},
				},
			},
			expected: "Weather location is missing required fields: latitude, longitude, API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			tt.config.checkWeatherLocations()

			output := strings.TrimSpace(buf.String())
			if tt.expected == "" {
				assert.Empty(t, output)
			} else {
				assert.NotEmpty(t, output)
			}
		})
	}
}
