package config

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"

	"github.com/labstack/echo/v4"
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

	if c.ImmichUrl != originalUrl {
		t.Errorf("ImmichUrl field was allowed to be changed: %s", c.ImmichUrl)
	}

	if c.ImmichApiKey != originalApi {
		t.Errorf("ImmichApiKey field was allowed to be changed: %s", c.ImmichUrl)
	}
}

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

	c.ConfigWithOverrides(echoContenx)

	if len(c.Person) != 2 {
		t.Errorf("People were not added: %s", c.Person)
	}
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
			if err != nil {
				t.Error("Config load err", err)
			}

			if c.ImmichUrl != test.Want {
				t.Error("did not format url correctly", c.ImmichUrl)
			}

		})
	}
}

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

	configWithBase.ConfigWithOverrides(echoContenx)

	t.Log("album", configWithBase.Album)

	if slices.Contains(configWithBase.Album, "BASE_ALBUM") {
		t.Errorf("BASE_ALBUM is present: %s", configWithBase.Album)
	}

	if len(configWithBase.Album) != 2 {
		t.Errorf("Albums were not added: %s", configWithBase.Album)
	}

	// configWithBase
	configWithoutBase := New()

	q = make(url.Values)
	q.Add("album", "ALBUM_1")
	q.Add("album", "ALBUM_2")

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec = httptest.NewRecorder()

	echoContenx = e.NewContext(req, rec)

	t.Log("Trying to add:", echoContenx.Request().URL.Query())

	configWithoutBase.ConfigWithOverrides(echoContenx)

	t.Log("album", configWithoutBase.Album)

	if len(configWithoutBase.Album) != 2 {
		t.Errorf("Albums were not added: %s", configWithoutBase.Album)
	}

	// configWithBaseOnly
	configWithBaseOnly := New()
	configWithBaseOnly.Album = []string{"BASE_ALUMB_1", "BASE_ALUMB_2"}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()

	echoContenx = e.NewContext(req, rec)

	t.Log("album", configWithBaseOnly.Album)

	if len(configWithBaseOnly.Album) != 2 {
		t.Errorf("Base albums did not persist: %s", configWithBaseOnly.Album)
	}
}
