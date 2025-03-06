package immich

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func (a *Asset) Me(requestID, deviceID string) (UserResponse, error) {
	var user UserResponse

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return user, err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/users/me",
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, user)
	body, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return user, err
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}
