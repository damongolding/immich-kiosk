package immich

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"

	"charm.land/log/v2"
)

func (a *Asset) Me(requestID, deviceID string) (UserResponse, error) {

	var user UserResponse

	u, err := url.Parse(a.RequestConfig.ImmichURL)
	if err != nil {
		return user, err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/users/me",
	}

	if a.RequestConfig.SelectedUser != "" {
		q := apiURL.Query()
		q.Set("user", a.RequestConfig.SelectedUser)
		apiURL.RawQuery = q.Encode()
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.RequestConfig, user)
	body, _, _, err := immichAPICall(a.Ctx, http.MethodGet, apiURL.String(), nil)
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

func (a *Asset) ApplyUserFromAssetID(assetID string) (string, string) {

	var userAPI string
	var userFound bool

	// assetID has @user
	id, user, ok := strings.Cut(assetID, "@")
	if ok {
		if userAPI, userFound = a.RequestConfig.ImmichUsersAPIKeys[user]; userFound {
			a.RequestConfig.SelectedUser = user
			a.RequestConfig.ImmichAPIKey = userAPI
			return id, user
		}
		log.Warn("User from assetID not found in API keys")
	}

	// User provided via URL query parameter
	if len(a.RequestConfig.URLParamUsers) > 0 {
		randomIndex := rand.IntN(len(a.RequestConfig.URLParamUsers))
		selectedUser := a.RequestConfig.URLParamUsers[randomIndex]
		if userAPI, userFound = a.RequestConfig.ImmichUsersAPIKeys[selectedUser]; userFound {
			a.RequestConfig.SelectedUser = selectedUser
			a.RequestConfig.ImmichAPIKey = userAPI
			return id, selectedUser
		}
		log.Warn("User from URL query parameter not found in API keys")
	}

	// use default
	a.ApplyDefaultUser()

	return assetID, ""
}

func (a *Asset) ApplyDefaultUser() {
	if defaultAPI, apiFound := a.RequestConfig.ImmichUsersAPIKeys["default"]; apiFound {
		a.RequestConfig.SelectedUser = ""
		a.RequestConfig.ImmichAPIKey = defaultAPI
	} else {
		log.Error("Default user not found in API keys")
	}
}

func (a *Asset) SelectedUser() string {
	return a.RequestConfig.SelectedUser
}
