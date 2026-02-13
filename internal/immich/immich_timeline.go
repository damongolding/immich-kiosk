package immich

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
)

func (a *Asset) TimelineBuckets(requestID, deviceID string) (TimelineBuckets, string, error) {
	var timelineBuckets TimelineBuckets

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return immichAPIFail(timelineBuckets, err, nil, "")
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "timeline", "buckets"),
	}

	queryParams := url.Values{}

	queryParams.Set("withStacked", "true")
	queryParams.Set("visibility", "timeline")
	queryParams.Set("withPartners", "true")

	apiURL.RawQuery = queryParams.Encode()

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, timelineBuckets)
	body, _, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return immichAPIFail(timelineBuckets, err, body, apiURL.String())
	}

	err = json.Unmarshal(body, &timelineBuckets)
	if err != nil {
		return immichAPIFail(timelineBuckets, err, body, apiURL.String())
	}

	return timelineBuckets, apiURL.String(), nil
}
