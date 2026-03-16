package immich

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
)

type StatisticsResponse struct {
	Images int `json:"images"`
	Videos int `json:"videos"`
	Total  int `json:"total"`
}

// TotalAssetCount returns the total number of assets from the (default user) timeline.
func (a *Asset) TotalAssetCount() int {
	s, err := a.assetsStatistics()
	if err != nil {
		log.Error("TotalAssetCount", "err", err)
		return 0
	}

	return s.Total
}

// assetsStatistics makes a request to the Immich API to retrieve statistics about timeline assets.
func (a *Asset) assetsStatistics() (StatisticsResponse, error) {
	var stats StatisticsResponse

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return stats, err
	}

	apiURL := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     path.Join("api", "assets", "statistics"),
		RawQuery: "visibility=timeline",
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, kiosk.DebugID, kiosk.GlobalCache, a.requestConfig, stats)
	body, _, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return stats, err
	}

	err = json.Unmarshal(body, &stats)
	if err != nil {
		return stats, err
	}

	return stats, nil
}
