package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"path"

	"github.com/charmbracelet/log"
	"github.com/google/go-querystring/query"
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

	// Parse HH:MM:SS format
	var hours, minutes, seconds int
	_, err := fmt.Sscanf(a.Duration, "%d:%d:%d", &hours, &minutes, &seconds)
	if err != nil {
		log.Error("Failed to parse video duration", "ID", a.ID, "duration", a.Duration)
		return false
	}
	totalSeconds := hours*3600 + minutes*60 + seconds

	return totalSeconds >= 1
}

func (a *Asset) AddVideos(requestID, deviceID string, assets *[]Asset, apiURL url.URL, requestBody SearchRandomBody) error {

	if len(*assets) == 0 {
		return nil
	}

	for _, a := range *assets {
		if a.Type == VideoType {
			return nil
		}
	}

	var videoAssets []Asset

	videoRequestBody := requestBody
	videoRequestBody.Type = string(VideoType)

	queries, err := query.Values(videoRequestBody)
	if err != nil {
		return err
	}
	apiURL.RawQuery = fmt.Sprintf("kiosk=%x", sha256.Sum256([]byte(queries.Encode())))

	jsonBody, err := json.Marshal(videoRequestBody)
	if err != nil {
		return err
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, videoAssets)
	apiBody, _, err := immichAPICall(a.ctx, http.MethodPost, apiURL.String(), jsonBody)
	if err != nil {
		return err
	}

	err = json.Unmarshal(apiBody, &videoAssets)
	if err != nil {
		return err
	}

	if len(videoAssets) == 0 {
		return nil
	}

	mergeVideoAssetsRandomly(assets, videoAssets)

	return nil
}

func mergeVideoAssetsRandomly(imageAssets *[]Asset, videoAssets []Asset) {
	*imageAssets = append(*imageAssets, videoAssets...)

	// Shuffle the combined slice
	rand.Shuffle(len(*imageAssets), func(i, j int) {
		(*imageAssets)[i], (*imageAssets)[j] = (*imageAssets)[j], (*imageAssets)[i]
	})
}
