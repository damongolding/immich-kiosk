package immich

import (
	"net/http"
	"net/url"
	"path"
)

// Video retrieves the video asset from Immich server.
// Returns the video data as a byte slice, the API URL used for the request, and any error encountered.
// The video is returned in octet-stream format.
func (i *Asset) Video() ([]byte, string, error) {

	var responseBody []byte

	u, err := url.Parse(i.requestConfig.ImmichURL)
	if err != nil {
		return responseBody, "", err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", i.ID, "video", "playback"),
	}

	octetStreamHeader := map[string]string{"Accept": "application/octet-stream"}

	responseBody, err = i.immichAPICall(http.MethodGet, apiURL.String(), nil, octetStreamHeader)
	if err != nil {
		return responseBody, apiURL.String(), err
	}

	return responseBody, apiURL.String(), nil

}
