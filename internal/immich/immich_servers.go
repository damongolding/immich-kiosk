package immich

import (
	"math/rand/v2"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/config"
)

// ApplyServer selects an Immich server for this asset request.
// When pickRandom is true and no SelectedServer is set, a server is chosen at random
// from URLParamServers (if any) or from all configured ImmichServers.
// When pickRandom is false, only an already-selected server name is applied
// (e.g. from history or form actions).
func (a *Asset) ApplyServer(pickRandom bool) {
	if len(a.requestConfig.ImmichServers) == 0 {
		return
	}

	if a.requestConfig.SelectedServer != "" {
		if !a.applyNamedServer(a.requestConfig.SelectedServer) {
			log.Warn("Selected server not found in config", "server", a.requestConfig.SelectedServer)
		}
		return
	}

	if !pickRandom {
		return
	}

	candidates := a.serverCandidates()
	if len(candidates) == 0 {
		log.Warn("No valid Immich servers found for selection")
		return
	}

	name := candidates[rand.IntN(len(candidates))]
	if !a.applyNamedServer(name) {
		log.Warn("Server not found in config", "server", name)
	}
}

// serverCandidates returns Immich server names available for random selection.
// When URLParamServers is set, only known names from that list are returned.
func (a *Asset) serverCandidates() []string {
	if len(a.requestConfig.URLParamServers) > 0 {
		candidates := make([]string, 0, len(a.requestConfig.URLParamServers))
		for _, name := range a.requestConfig.URLParamServers {
			if _, ok := a.requestConfig.ImmichServers[name]; ok {
				candidates = append(candidates, name)
				continue
			}
			log.Warn("Server from URL query parameter not found in config", "server", name)
		}
		return candidates
	}

	candidates := make([]string, 0, len(a.requestConfig.ImmichServers))
	for name := range a.requestConfig.ImmichServers {
		candidates = append(candidates, name)
	}
	return candidates
}

// applyNamedServer applies the Immich URL, API key, and external URL
// for the named server on this asset's request config.
// Returns false if the server name is not present in ImmichServers.
func (a *Asset) applyNamedServer(name string) bool {
	server, ok := a.requestConfig.ImmichServers[name]
	if !ok {
		return false
	}

	a.requestConfig.SelectedServer = name
	a.requestConfig.ImmichURL = server.URL
	a.requestConfig.ImmichAPIKey = server.APIKey
	a.requestConfig.ImmichExternalURL = server.ExternalURL

	if a.requestConfig.ImmichUsersAPIKeys == nil {
		a.requestConfig.ImmichUsersAPIKeys = make(map[string]string)
	}
	a.requestConfig.ImmichUsersAPIKeys["default"] = server.APIKey

	return true
}

// SelectedServer returns the Immich server name selected for this asset request.
func (a *Asset) SelectedServer() string {
	return a.requestConfig.SelectedServer
}

// ImmichURL returns the Immich base URL currently applied to this asset request.
func (a *Asset) ImmichURL() string {
	return a.requestConfig.ImmichURL
}

// ImmichExternalURL returns the external Immich URL currently applied to this asset request.
func (a *Asset) ImmichExternalURL() string {
	return a.requestConfig.ImmichExternalURL
}

// RequestConfig returns the request configuration currently applied to this asset.
func (a *Asset) RequestConfig() config.Config {
	return a.requestConfig
}
