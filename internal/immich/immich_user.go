package immich

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/log"
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
	body, _, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return user, err
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (a *Asset) UserOwnsAsset(requestID, deviceID string) bool {

	me, meErr := a.Me(requestID, deviceID)
	if meErr != nil {
		log.Error("Error getting user", "error", meErr)
		return false
	}

	return strings.EqualFold(me.ID, a.OwnerID)
}
