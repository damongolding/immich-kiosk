package immich

import (
	"context"
	"testing"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyServerNoServers(t *testing.T) {
	cfg := config.New()
	cfg.ImmichURL = "http://primary:2283"
	cfg.ImmichAPIKey = "primary-key"

	asset := New(context.Background(), *cfg)
	asset.ApplyServer(true)

	assert.Equal(t, "http://primary:2283", asset.ImmichURL())
	assert.Equal(t, "", asset.SelectedServer())
}

func TestApplyServerSelected(t *testing.T) {
	cfg := config.New()
	cfg.ImmichURL = "http://primary:2283"
	cfg.ImmichAPIKey = "primary-key"
	cfg.ImmichServers = map[string]config.ImmichServer{
		"home":  {URL: "http://home:2283", APIKey: "home-key", ExternalURL: "https://photos.home"},
		"cabin": {URL: "http://cabin:2283", APIKey: "cabin-key"},
	}
	cfg.SelectedServer = "cabin"

	asset := New(context.Background(), *cfg)
	asset.ApplyServer(false)

	assert.Equal(t, "cabin", asset.SelectedServer())
	assert.Equal(t, "http://cabin:2283", asset.ImmichURL())
	assert.Equal(t, "cabin-key", asset.requestConfig.ImmichAPIKey)
	assert.Equal(t, "cabin-key", asset.requestConfig.ImmichUsersAPIKeys["default"])
}

func TestApplyServerURLParams(t *testing.T) {
	cfg := config.New()
	cfg.ImmichServers = map[string]config.ImmichServer{
		"home":  {URL: "http://home:2283", APIKey: "home-key", ExternalURL: "https://photos.home"},
		"cabin": {URL: "http://cabin:2283", APIKey: "cabin-key"},
	}
	cfg.URLParamServers = []string{"home"}

	asset := New(context.Background(), *cfg)
	asset.ApplyServer(true)

	assert.Equal(t, "home", asset.SelectedServer())
	assert.Equal(t, "http://home:2283", asset.ImmichURL())
	assert.Equal(t, "https://photos.home", asset.ImmichExternalURL())
}

func TestApplyServerRandomFromAll(t *testing.T) {
	cfg := config.New()
	cfg.ImmichServers = map[string]config.ImmichServer{
		"home":  {URL: "http://home:2283", APIKey: "home-key"},
		"cabin": {URL: "http://cabin:2283", APIKey: "cabin-key"},
	}

	seen := map[string]bool{}
	for range 50 {
		asset := New(context.Background(), *cfg)
		asset.ApplyServer(true)
		require.NotEmpty(t, asset.SelectedServer())
		seen[asset.SelectedServer()] = true
	}

	assert.True(t, seen["home"] || seen["cabin"])
	assert.Len(t, seen, 2)
}

func TestApplyServerExternalURL(t *testing.T) {
	cfg := config.New()
	cfg.ImmichExternalURL = "https://fallback.example"
	cfg.ImmichServers = map[string]config.ImmichServer{
		"home": {URL: "http://home:2283", APIKey: "home-key", ExternalURL: "https://photos.home"},
	}
	cfg.SelectedServer = "home"

	asset := New(context.Background(), *cfg)
	asset.ApplyServer(false)

	assert.Equal(t, "https://photos.home", asset.ImmichExternalURL())
}
