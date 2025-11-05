package immich

import (
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/charmbracelet/log"
)

// Video retrieves the video asset from Immich server.
// Returns the video data as a byte slice, the contentType, and any error encountered.
// The video is returned in octet-stream format.
func (a *Asset) Video() ([]byte, string, error) {

	var responseBody []byte
	var contentType string

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return responseBody, "", err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", a.ID, "video", "playback"),
	}

	octetStreamHeader := map[string]string{"Accept": "application/octet-stream"}

	responseBody, contentType, err = a.immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil, octetStreamHeader)
	if err != nil {
		return responseBody, contentType, err
	}

	return responseBody, contentType, nil
}

// durationCheck verifies that the video duration string in the Asset is valid and represents
// a duration of at least one second.
//
// Returns true if the duration is valid and at least one second, false otherwise.
//
// The duration string is expected to be in the format "HH:MM:SS".
func (a *Asset) durationCheck() bool {

	d, err := time.Parse("15:04:05", a.Duration)
	if err != nil {
		log.Error("Failed to parse video duration", "ID", a.ID, "duration", a.Duration)
		return false
	}
	if d.Hour() == 0 && d.Minute() == 0 && d.Second() < 1 {
		return false
	}

	return true
}
