package config

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"charm.land/log/v2"

	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/labstack/echo/v5"
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
	if err != nil {
		t.Errorf("ConfigWithOverrides should not return an error: %v", err)
	}

	if c.ImmichURL != originalURL {
		t.Errorf("ImmichURL field was allowed to be changed: got %q, want %q", c.ImmichURL, originalURL)
	}
	if c.ImmichAPIKey != originalAPI {
		t.Errorf("ImmichAPIKey field was allowed to be changed: got %q, want %q", c.ImmichAPIKey, originalAPI)
	}
	if !reflect.DeepEqual(c.ImmichUsersAPIKeys, originalUsersAPIKeys) {
		t.Errorf("ImmichUsersAPIKeys field was allowed to be changed: got %v, want %v", c.ImmichUsersAPIKeys, originalUsersAPIKeys)
	}
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
	if err != nil {
		t.Errorf("ConfigWithOverrides should not return an error: %v", err)
	}

	if got := len(c.People); got != 2 {
		t.Errorf("Expected 2 people to be added: got %d, want 2 (%v)", got, c.People)
	}
	if !slices.Contains(c.People, "bea") {
		t.Errorf("Expected 'bea' to be added to Person slice: got %v", c.People)
	}
	if !slices.Contains(c.People, "laura") {
		t.Errorf("Expected 'laura' to be added to Person slice: got %v", c.People)
	}
}

// TestMalformedURLs testing urls without scheme or ports
func TestMalformedURLs(t *testing.T) {
	tests := []struct {
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
			if err != nil {
				t.Errorf("Config load should not return an error: %v", err)
			}

			if c.ImmichURL != test.Want {
				t.Errorf("ImmichURL = %q, want %q", c.ImmichURL, test.Want)
			}
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
	if err != nil {
		t.Errorf("ConfigWithOverrides should not return an error: %v", err)
	}

	t.Log("album", configWithBase.Albums)

	if slices.Contains(configWithBase.Albums, "BASE_ALBUM") {
		t.Errorf("BASE_ALBUM should not be present: got %v", configWithBase.Albums)
	}
	if got := len(configWithBase.Albums); got != 2 {
		t.Errorf("Expected 2 albums to be added: got %d, want 2 (%v)", got, configWithBase.Albums)
	}
	if !slices.Contains(configWithBase.Albums, "ALBUM_1") {
		t.Errorf("ALBUM_1 should be present: got %v", configWithBase.Albums)
	}
	if !slices.Contains(configWithBase.Albums, "ALBUM_2") {
		t.Errorf("ALBUM_2 should be present: got %v", configWithBase.Albums)
	}

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
	if err != nil {
		t.Errorf("ConfigWithOverrides should not return an error: %v", err)
	}

	t.Log("album", configWithoutBase.Albums)

	if got := len(configWithoutBase.Albums); got != 2 {
		t.Errorf("Expected 2 albums to be added: got %d, want 2 (%v)", got, configWithoutBase.Albums)
	}
	if !slices.Contains(configWithoutBase.Albums, "ALBUM_1") {
		t.Errorf("ALBUM_1 should be present: got %v", configWithoutBase.Albums)
	}
	if !slices.Contains(configWithoutBase.Albums, "ALBUM_2") {
		t.Errorf("ALBUM_2 should be present: got %v", configWithoutBase.Albums)
	}

	// configWithBaseOnly
	configWithBaseOnly := New()
	configWithBaseOnly.Albums = []string{"BASE_ALBUM_1", "BASE_ALBUM_2"}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()

	echoContenx = e.NewContext(req, rec)

	err = configWithBaseOnly.ConfigWithOverrides(echoContenx.QueryParams(), echoContenx)
	if err != nil {
		t.Errorf("ConfigWithOverrides should not return an error: %v", err)
	}

	t.Log("album", configWithBaseOnly.Albums)

	if got := len(configWithBaseOnly.Albums); got != 2 {
		t.Errorf("Base albums should persist: got %d, want 2 (%v)", got, configWithBaseOnly.Albums)
	}
	if !slices.Contains(configWithBaseOnly.Albums, "BASE_ALBUM_1") {
		t.Errorf("BASE_ALBUM_1 should be present: got %v", configWithBaseOnly.Albums)
	}
	if !slices.Contains(configWithBaseOnly.Albums, "BASE_ALBUM_2") {
		t.Errorf("BASE_ALBUM_2 should be present: got %v", configWithBaseOnly.Albums)
	}
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

			if !reflect.DeepEqual(c.Albums, tc.expectedAlbum) {
				t.Errorf("Albums = %#v, want %#v", c.Albums, tc.expectedAlbum)
			}
			if !reflect.DeepEqual(c.People, tc.expectedPerson) {
				t.Errorf("People = %#v, want %#v", c.People, tc.expectedPerson)
			}
		})
	}
}

// TestDisableURLQueriesStillBindsClientData tests that client data is bound when URL queries are disabled
func TestDisableURLQueriesStillBindsClientData(t *testing.T) {
	c := New()
	c.Duration = 600
	c.Kiosk.DisableURLQueries = true

	e := echo.New()

	form := make(url.Values)
	form.Add("client_width", "1280")
	form.Add("client_height", "800")
	form.Add("duration", "5")

	req := httptest.NewRequest(http.MethodPost, "/asset/new", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	formValues, err := echoContenx.FormValues()
	if err != nil {
		t.Errorf("FormValues should not return an error: %v", err)
	}

	err = c.ConfigWithOverrides(formValues, echoContenx)
	if err != nil {
		t.Errorf("ConfigWithOverrides should not return an error: %v", err)
	}

	if c.ClientData.Width != 1280 {
		t.Errorf("ClientData.Width = %d, want 1280 (client width should be bound)", c.ClientData.Width)
	}
	if c.ClientData.Height != 800 {
		t.Errorf("ClientData.Height = %d, want 800 (client height should be bound)", c.ClientData.Height)
	}
	if c.Duration != 600 {
		t.Errorf("Duration = %d, want 600 (should not be overridden)", c.Duration)
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
				Weather: WeatherConfig{
					Locations: []WeatherLocation{
						{Name: "City", Lat: "123", Lon: "456", API: "abc123"},
					},
				},
			},
			expected: "",
		},
		{
			name: "Missing name",
			config: &Config{
				Weather: WeatherConfig{
					Locations: []WeatherLocation{
						{Lat: "123", Lon: "456", API: "abc123"},
					},
				},
			},
			expected: "Weather location is missing required fields: name",
		},
		{
			name: "Missing latitude",
			config: &Config{
				Weather: WeatherConfig{
					Locations: []WeatherLocation{
						{Name: "City", Lon: "456", API: "abc123"},
					},
				},
			},
			expected: "Weather location is missing required fields: latitude",
		},
		{
			name: "Missing longitude",
			config: &Config{
				Weather: WeatherConfig{
					Locations: []WeatherLocation{{Name: "City", Lat: "123", API: "abc123"}},
				},
			},
			expected: "Weather location is missing required fields: longitude",
		},
		{
			name: "Missing API key",
			config: &Config{
				Weather: WeatherConfig{
					Locations: []WeatherLocation{{Name: "City", Lat: "123", Lon: "456"}},
				},
			},
			expected: "Weather location is missing required fields: API key",
		},
		{
			name: "Multiple missing fields",
			config: &Config{
				Weather: WeatherConfig{
					Locations: []WeatherLocation{{Name: "City"}},
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
			if tt.expected == "" && output != "" {
				t.Errorf("expected no log output, got %q", output)
			}
			if tt.expected != "" && output == "" {
				t.Error("expected log output, got none")
			}
		})
	}
}

func TestConfig_checkIDs(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		ids           []string
		allowSuffix   bool
		allowKeywords bool
		wantWarn      bool
		wantValues    []string // ids expected to appear in the warning output
	}{
		{
			name:          "valid uuid",
			key:           "album_ids",
			ids:           []string{"550e8400-e29b-41d4-a716-446655440000"},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      false,
		},
		{
			name:          "valid uuid with suffix",
			key:           "album_ids",
			ids:           []string{"550e8400-e29b-41d4-a716-446655440000@user"},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      false,
		},
		{
			name:          "valid uuid with invalid suffix",
			key:           "album_ids",
			ids:           []string{"550e8400-e29b-41d4-a716-446655440000@"},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      true,
		},
		{
			name:          "keyword all",
			key:           "album_ids",
			ids:           []string{kiosk.AlbumKeywordAll},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      false,
		},
		{
			name: "keyword owned/shared/favourites/favorites",
			key:  "album_ids",
			ids: []string{
				kiosk.AlbumKeywordOwned,
				kiosk.AlbumKeywordShared,
				kiosk.AlbumKeywordFavourites,
				kiosk.AlbumKeywordFavorites,
			},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      false,
		},
		{
			name:          "invalid id",
			key:           "album_ids",
			ids:           []string{"not-a-uuid"},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      true,
			wantValues:    []string{"not-a-uuid"},
		},
		{
			name:          "mixed valid and invalid",
			key:           "person_ids",
			ids:           []string{"550e8400-e29b-41d4-a716-446655440000", "garbage", kiosk.AlbumKeywordAll},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      true,
			wantValues:    []string{"garbage"},
		},
		{
			name: "keywords",
			key:  "albums",
			ids: []string{
				kiosk.AlbumKeywordAll,
				kiosk.AlbumKeywordFavorites,
				kiosk.AlbumKeywordFavourites,
				kiosk.AlbumKeywordOwned,
				kiosk.AlbumKeywordShared,
				kiosk.PersonKeywordAll,
			},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      false,
		},
		{
			name: "keywords with suffix",
			key:  "albums",
			ids: []string{
				kiosk.AlbumKeywordAll + "@user",
				kiosk.AlbumKeywordFavorites + "@user",
				kiosk.AlbumKeywordFavourites + "@user",
				kiosk.AlbumKeywordOwned + "@user",
				kiosk.AlbumKeywordShared + "@user",
				kiosk.PersonKeywordAll + "@user",
			},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      false,
		},
		{
			name: "keywords with suffix but no user",
			key:  "albums",
			ids: []string{
				kiosk.AlbumKeywordAll + "@",
				kiosk.AlbumKeywordFavorites + "@",
				kiosk.AlbumKeywordFavourites + "@",
				kiosk.AlbumKeywordOwned + "@",
				kiosk.AlbumKeywordShared + "@",
				kiosk.PersonKeywordAll + "@",
			},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      true,
		},
		{
			name: "keywords with suffix",
			key:  "albums",
			ids: []string{
				kiosk.AlbumKeywordAll + "@user",
				kiosk.AlbumKeywordFavorites + "@user",
				kiosk.AlbumKeywordFavourites + "@user",
				kiosk.AlbumKeywordOwned + "@user",
				kiosk.AlbumKeywordShared + "@user",
				kiosk.PersonKeywordAll + "@user",
			},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      false,
		},

		{
			name:          "empty slice",
			key:           "album_ids",
			ids:           []string{},
			allowSuffix:   true,
			allowKeywords: true,
			wantWarn:      false,
		},
		{
			name:          "excluded album with suffix",
			key:           "excluded_albums",
			ids:           []string{"550e8400-e29b-41d4-a716-446655440000@user"},
			allowSuffix:   false,
			allowKeywords: false,
			wantWarn:      true,
		},
		{
			name:          "keyword rejected when keywords disallowed",
			key:           "excluded_albums",
			ids:           []string{kiosk.AlbumKeywordAll},
			allowSuffix:   false,
			allowKeywords: false,
			wantWarn:      true,
			wantValues:    []string{kiosk.AlbumKeywordAll},
		},
		{
			name:          "person keyword rejected when keywords disallowed",
			key:           "excluded_people",
			ids:           []string{kiosk.PersonKeywordAll},
			allowSuffix:   false,
			allowKeywords: false,
			wantWarn:      true,
			wantValues:    []string{kiosk.PersonKeywordAll},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			c := &Config{}
			c.checkIDs(tt.key, tt.ids, tt.allowSuffix, tt.allowKeywords)

			out := buf.String()

			if tt.wantWarn && !strings.Contains(out, "WARN") {
				t.Errorf("expected warning log, got none. output: %q", out)
			}
			if !tt.wantWarn && strings.Contains(out, "WARN") {
				t.Errorf("expected no warning log, got: %q", out)
			}
			for _, v := range tt.wantValues {
				if !strings.Contains(out, v) {
					t.Errorf("expected log output to contain %q, got: %q", v, out)
				}
			}
		})
	}
}

func TestCheckTimelines(t *testing.T) {
	tests := []struct {
		name      string
		timelines []string
		apiKeys   map[string]string
		want      []string
	}{
		{
			name:      "nil timelines is a no-op",
			timelines: nil,
			apiKeys:   map[string]string{"alice": "key-a"},
			want:      nil,
		},
		{
			name:      "empty timelines is a no-op",
			timelines: []string{},
			apiKeys:   map[string]string{"alice": "key-a"},
			want:      []string{},
		},
		{
			name:      "all users valid",
			timelines: []string{"alice", "bob"},
			apiKeys:   map[string]string{"alice": "key-a", "bob": "key-b"},
			want:      []string{"alice", "bob"},
		},
		{
			name:      "unknown user is removed",
			timelines: []string{"alice", "ghost"},
			apiKeys:   map[string]string{"alice": "key-a"},
			want:      []string{"alice"},
		},
		{
			name:      "user with empty api key is removed",
			timelines: []string{"alice", "bob"},
			apiKeys:   map[string]string{"alice": "key-a", "bob": ""},
			want:      []string{"alice"},
		},
		{
			name:      "all users invalid leaves empty list",
			timelines: []string{"ghost", "phantom"},
			apiKeys:   map[string]string{"alice": "key-a"},
			want:      []string{},
		},
		{
			name:      "nil api key map removes everyone",
			timelines: []string{"alice"},
			apiKeys:   nil,
			want:      []string{},
		},
		{
			name:      "order is preserved",
			timelines: []string{"carol", "alice", "bob"},
			apiKeys:   map[string]string{"alice": "a", "bob": "b", "carol": "c"},
			want:      []string{"carol", "alice", "bob"},
		},
		{
			name:      "usernames are case sensitive",
			timelines: []string{"Alice"},
			apiKeys:   map[string]string{"alice": "key-a"},
			want:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Timelines:          tt.timelines,
				ImmichUsersAPIKeys: tt.apiKeys,
			}

			c.checkTimelines()

			if !slices.Equal(c.Timelines, tt.want) {
				t.Errorf("Timelines = %v, want %v", c.Timelines, tt.want)
			}
		})
	}
}
