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

	if a.requestConfig.SelectedUser != "" {
		apiURL.RawQuery += "&user=" + a.requestConfig.SelectedUser
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

func (a *Asset) SwitchUserFromID(assetID string) string {

	if !strings.Contains(assetID, "@") {
		log.Info("Switching user to", "user", "default")
		if defaultAPI, ok := a.requestConfig.ImmichUsersAPIKeys["default"]; ok {
			log.Info("Switched user to", "user", "default")
			a.requestConfig.SelectedUser = ""
			a.requestConfig.ImmichAPIKey = defaultAPI
		}
		return assetID
	}

	parts := strings.Split(assetID, "@")
	if len(parts) != 2 {
		log.Error("Invalid user format", "user", assetID)
		return assetID
	}

	user := parts[0]
	assetID = parts[1]

	log.Info("Switching user to", "user", user)

	if userAPI, ok := a.requestConfig.ImmichUsersAPIKeys[user]; ok {
		log.Info("Switched user to", "user", user)
		a.requestConfig.SelectedUser = user
		a.requestConfig.ImmichAPIKey = userAPI
	}

	return assetID
}

func (a *Asset) SelectedUser() string {
	return a.requestConfig.SelectedUser
}
