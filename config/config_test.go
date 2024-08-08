package config

import (
	"net/url"
	"slices"
	"strings"
	"testing"
)

// TestTransition check transitions for being transformed
func TestTransition(t *testing.T) {

	originalUrl := "https://my-server.com"
	originalApi := "123456"

	c := Config{
		ImmichUrl:    originalUrl,
		ImmichApiKey: originalApi,
	}

	transitions := []string{"FadE", "NONE", "CroSs-FADE"}

	for _, transition := range transitions {

		q := url.Values{}

		q.Add("transition", transition)

		c.ConfigWithOverrides(q)

		if c.Transition != strings.ToLower(transition) {
			t.Errorf("Transition was not transformed to lowercase: %s", c.Transition)
		}
	}
}

// TestConfigWithOverrides testing whether ImmichUrl and ImmichApiKey are immutable
func TestImmichUrlImmichApiKeyImmutability(t *testing.T) {

	originalUrl := "https://my-server.com"
	originalApi := "123456"

	c := Config{
		ImmichUrl:    originalUrl,
		ImmichApiKey: originalApi,
	}

	q := url.Values{}

	q.Add("immich_url", "https://my-new-server.com")
	q.Add("immich_api_key", "9999")

	c.ConfigWithOverrides(q)

	if c.ImmichUrl != originalUrl {
		t.Errorf("ImmichUrl field was allowed to be changed: %s", c.ImmichUrl)
	}

	if c.ImmichApiKey != originalApi {
		t.Errorf("ImmichApiKey field was allowed to be changed: %s", c.ImmichUrl)
	}
}

func TestImmichUrlImmichMulitplePerson(t *testing.T) {

	c := Config{}

	q := url.Values{}

	q.Add("person", "bea")
	q.Add("person", "laura")

	t.Log("Trying to add:", q)

	c.ConfigWithOverrides(q)

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
				t.Error(err)
			}

			if c.ImmichUrl != test.Want {
				t.Error("did not format url correctly", c.ImmichUrl)
			}

		})
	}
}

func TestImmichUrlImmichMulitpleAlbum(t *testing.T) {

	// configWithBase
	configWithBase := Config{
		Album: []string{"BASE_ALBUM"},
	}

	q := url.Values{}

	q.Add("album", "ALBUM_1")
	q.Add("album", "ALBUM_2")

	t.Log("Trying to add:", q)

	configWithBase.ConfigWithOverrides(q)

	t.Log("album", configWithBase.Album)

	if slices.Contains(configWithBase.Album, "BASE_ALBUM") {
		t.Errorf("BASE_ALBUM is present: %s", configWithBase.Album)
	}

	if len(configWithBase.Album) != 2 {
		t.Errorf("Albums were not added: %s", configWithBase.Album)
	}

	// configWithBase
	configWithoutBase := Config{}

	q = url.Values{}

	q.Add("album", "ALBUM_1")
	q.Add("album", "ALBUM_2")

	t.Log("Trying to add:", q)

	configWithoutBase.ConfigWithOverrides(q)

	t.Log("album", configWithoutBase.Album)

	if len(configWithoutBase.Album) != 2 {
		t.Errorf("Albums were not added: %s", configWithoutBase.Album)
	}

	// configWithBaseOnly
	configWithBaseOnly := Config{
		Album: []string{"BASE_ALUMB_1", "BASE_ALUMB_2"},
	}

	q = url.Values{}

	configWithBaseOnly.ConfigWithOverrides(q)

	t.Log("album", configWithBaseOnly.Album)

	if len(configWithBaseOnly.Album) != 2 {
		t.Errorf("Base albums did not persist: %s", configWithBaseOnly.Album)
	}
}
