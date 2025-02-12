package immich

import (
	"encoding/json"
	"net/url"
	"path"
)

func (i *ImmichAsset) AllTags(requestID, deviceID string) ([]Tag, string, error) {
	var tags []Tag

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		return immichApiFail(tags, err, nil, "")
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "tags"),
	}

	immichApiCall := withImmichApiCache(i.immichApiCall, requestID, deviceID, tags)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		return immichApiFail(tags, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &tags)
	if err != nil {
		return immichApiFail(tags, err, body, apiUrl.String())
	}

	return tags, apiUrl.String(), nil
}
