package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
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

	// Do not add videos if any videos are already present
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

	// If we've already reached the fetch limit for image assets, assume the user
	// has more images than videos. In this case, limit the number of videos added
	// so they do not exceed 10% of the final asset mix.
	if len(*assets) == a.requestConfig.Kiosk.FetchedAssetsSize {
		videoAssets = videoLimiter(len(*assets), videoAssets, 0.10)
	}
	mergeVideoAssetsRandomly(assets, videoAssets)

	return nil
}

// videoLimiter returns a randomly shuffled slice of video assets,
// limited so that when merged into an existing asset list of size `assetLen`,
// the resulting list does not exceed `videoLimit` percent videos.
//
// The calculation uses rounding instead of truncation to produce
// more natural results for small sample sizes.
//
// Parameters:
//   - assetLen: number of existing (non-video) assets already present
//   - videoAssets: slice of available video assets to choose from
//   - videoLimit: fraction (0.0–1.0) representing the maximum allowed video ratio
//
// Returns:
//
//	A shuffled slice of video assets whose count satisfies the video limit.
func videoLimiter(assetLen int, videoAssets []Asset, videoLimit float64) []Asset {
	if videoLimit < 0 || videoLimit > 1 {
		log.Warn("videoLimiter: limit must be between 0 and 1", "videoLimit", videoLimit)
		return videoAssets
	}

	e := assetLen
	v := len(videoAssets)

	// Calculate the raw maximum using the ratio constraint:
	//   x / (E + x) <= videoLimit
	// Solves to:
	//   x <= (videoLimit * E) / (1 - videoLimit)
	rawMax := (videoLimit * float64(e)) / (1 - videoLimit)

	// Apply rounding instead of flooring for smoother behavior on small sets
	rounded := int(math.Round(rawMax))

	// Clamp to valid range [0, V]
	maxAllowed := min(max(0, rounded), v)

	// Nothing to add
	if maxAllowed == 0 {
		return videoAssets[:0]
	}

	// Shuffle in place
	rand.Shuffle(v, func(i, j int) {
		videoAssets[i], videoAssets[j] = videoAssets[j], videoAssets[i]
	})

	return videoAssets[:maxAllowed]
}

func mergeVideoAssetsRandomly(existingAssets *[]Asset, videoAssets []Asset) {

	*existingAssets = append(*existingAssets, videoAssets...)

	// Shuffle the combined slice
	rand.Shuffle(len(*existingAssets), func(i, j int) {
		(*existingAssets)[i], (*existingAssets)[j] = (*existingAssets)[j], (*existingAssets)[i]
	})
}
