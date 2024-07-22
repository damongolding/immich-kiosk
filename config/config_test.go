package config

import (
	"net/url"
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

// TestMalformedURLs testing urls without scheme or ports
func TestMalformedURLs(t *testing.T) {

	var tests = []struct {
		KIOSK_IMMICH_URL string
		Want             string
	}{
		{KIOSK_IMMICH_URL: "nope", Want: defaultScheme + "nope:" + defaultImmichPort},
		{KIOSK_IMMICH_URL: "nope:", Want: defaultScheme + "nope:" + defaultImmichPort},
		{KIOSK_IMMICH_URL: "nope::", Want: defaultScheme + "nope:" + defaultImmichPort},
		{KIOSK_IMMICH_URL: "nope:32", Want: defaultScheme + "nope:32"},
		{KIOSK_IMMICH_URL: "nope.com", Want: defaultScheme + "nope.com:" + defaultImmichPort},
		{KIOSK_IMMICH_URL: "123.123.123.123", Want: defaultScheme + "123.123.123.123:" + defaultImmichPort},
		{KIOSK_IMMICH_URL: "http://123.123.123.123", Want: "http://123.123.123.123:" + defaultImmichPort},
		{KIOSK_IMMICH_URL: "http://123.123.123.123:" + defaultImmichPort, Want: "http://123.123.123.123:" + defaultImmichPort},
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
