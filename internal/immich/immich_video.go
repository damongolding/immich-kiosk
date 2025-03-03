package immich

import (
	"net/url"
	"path"
)

// Video retrieves the video asset from Immich server.
// Returns the video data as a byte slice, the API URL used for the request, and any error encountered.
// The video is returned in octet-stream format.
func (i *ImmichAsset) Video() ([]byte, string, error) {

	var responseBody []byte

	u, err := url.Parse(i.requestConfig.ImmichUrl)
	if err != nil {
		return responseBody, "", err
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", i.ID, "video", "playback"),
	}

	octetStreamHeader := map[string]string{"Accept": "application/octet-stream"}

	responseBody, err = i.immichApiCall("GET", apiUrl.String(), nil, octetStreamHeader)
	if err != nil {
		return responseBody, apiUrl.String(), err
	}

	return responseBody, apiUrl.String(), nil

}
