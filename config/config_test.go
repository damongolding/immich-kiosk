package config

import (
	"net/url"
	"testing"
)

// TestConfigWithOverrides testing whether ImmichUrl and ImmichApiKey are immutable
func TestConfigWithOverrides(t *testing.T) {

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
