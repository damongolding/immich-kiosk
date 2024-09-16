package config

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestConfigWithOverrides testing whether ImmichUrl and ImmichApiKey are immutable
func TestImmichUrlImmichApiKeyImmutability(t *testing.T) {

	originalUrl := "https://my-server.com"
	originalApi := "123456"

	c := New()
	c.ImmichUrl = originalUrl
	c.ImmichApiKey = originalApi

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	req.URL.Query().Add("immich_url", "https://my-new-server.com")
	req.URL.Query().Add("immich_api_key", "9999")

	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	c.ConfigWithOverrides(echoContenx)

	assert.Equal(t, originalUrl, c.ImmichUrl, "ImmichUrl field was allowed to be changed")
	assert.Equal(t, originalApi, c.ImmichApiKey, "ImmichApiKey field was allowed to be changed")
}

// TestImmichUrlImmichMulitplePerson tests the addition of multiple persons to the config
func TestImmichUrlImmichMulitplePerson(t *testing.T) {
	c := New()

	e := echo.New()

	q := make(url.Values)
	q.Add("person", "bea")
	q.Add("person", "laura")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.Request().URL.Query())

	err := c.ConfigWithOverrides(echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	assert.Equal(t, 2, len(c.Person), "Expected 2 people to be added")
	assert.Contains(t, c.Person, "bea", "Expected 'bea' to be added to Person slice")
	assert.Contains(t, c.Person, "laura", "Expected 'laura' to be added to Person slice")
}

// TestMalformedURLs testing urls without scheme or ports
func TestMalformedURLs(t *testing.T) {

	var tests = []struct {
		KIOSK_IMMICH_URL string
		Want             string
	}{
		{KIOSK_IMMICH_URL: "nope", Want: defaultScheme + "nope"},
		{KIOSK_IMMICH_URL: "192.168.1.1", Want: defaultScheme + "192.168.1.1"},
		{KIOSK_IMMICH_URL: "192.168.1.1:1234", Want: defaultScheme + "192.168.1.1:1234"},
		{KIOSK_IMMICH_URL: "https://192.168.1.1:1234", Want: "https://192.168.1.1:1234"},
		{KIOSK_IMMICH_URL: "nope:32", Want: defaultScheme + "nope:32"},
	}

	for _, test := range tests {

		t.Run(test.KIOSK_IMMICH_URL, func(t *testing.T) {
			t.Setenv("KIOSK_IMMICH_URL", test.KIOSK_IMMICH_URL)
			t.Setenv("KIOSK_IMMICH_API_KEY", "12345")

			var c Config

			err := c.Load()
			assert.NoError(t, err, "Config load should not return an error")

			assert.Equal(t, test.Want, c.ImmichUrl, "ImmichUrl should be formatted correctly")
		})
	}
}

// TestImmichUrlImmichMulitpleAlbum tests the addition and overriding of multiple albums in the config
func TestImmichUrlImmichMulitpleAlbum(t *testing.T) {

	// configWithBase
	configWithBase := New()
	configWithBase.Album = []string{"BASE_ALBUM"}

	e := echo.New()

	q := make(url.Values)
	q.Add("album", "ALBUM_1")
	q.Add("album", "ALBUM_2")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	echoContenx := e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.Request().URL.Query())

	err := configWithBase.ConfigWithOverrides(echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	t.Log("album", configWithBase.Album)

	assert.NotContains(t, configWithBase.Album, "BASE_ALBUM", "BASE_ALBUM should not be present")
	assert.Equal(t, 2, len(configWithBase.Album), "Expected 2 albums to be added")
	assert.Contains(t, configWithBase.Album, "ALBUM_1", "ALBUM_1 should be present")
	assert.Contains(t, configWithBase.Album, "ALBUM_2", "ALBUM_2 should be present")

	// configWithoutBase
	configWithoutBase := New()

	q = make(url.Values)
	q.Add("album", "ALBUM_1")
	q.Add("album", "ALBUM_2")

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec = httptest.NewRecorder()

	echoContenx = e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.Request().URL.Query())

	err = configWithoutBase.ConfigWithOverrides(echoContenx)
	assert.NoError(t, err, "ConfigWithOverrides should not return an error")

	t.Log("album", configWithoutBase.Album)

	assert.Equal(t, 2, len(configWithoutBase.Album), "Expected 2 albums to be added")
	assert.Contains(t, configWithoutBase.Album, "ALBUM_1", "ALBUM_1 should be present")
	assert.Contains(t, configWithoutBase.Album, "ALBUM_2", "ALBUM_2 should be present")

	// configWithBaseOnly
	configWithBaseOnly := New()
	configWithBaseOnly.Album = []string{"BASE_ALBUM_1", "BASE_ALBUM_2"}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()

	echoContenx = e.NewContext(req, rec)

	configWithBaseOnly.ConfigWithOverrides(echoContenx)

	t.Log("album", configWithBaseOnly.Album)

	assert.Equal(t, 2, len(configWithBaseOnly.Album), "Base albums should persist")
	assert.Contains(t, configWithBaseOnly.Album, "BASE_ALBUM_1", "BASE_ALBUM_1 should be present")
	assert.Contains(t, configWithBaseOnly.Album, "BASE_ALBUM_2", "BASE_ALBUM_2 should be present")
}
