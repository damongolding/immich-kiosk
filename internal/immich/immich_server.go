package immich

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"charm.land/log/v2"
)

type ServerAboutResponse struct {
	Build                      *string `json:"build,omitempty"`
	BuildImage                 *string `json:"buildImage,omitempty"`
	BuildImageURL              *string `json:"buildImageUrl,omitempty"`
	BuildURL                   *string `json:"buildUrl,omitempty"`
	Exiftool                   *string `json:"exiftool,omitempty"`
	Ffmpeg                     *string `json:"ffmpeg,omitempty"`
	Imagemagick                *string `json:"imagemagick,omitempty"`
	Libvips                    *string `json:"libvips,omitempty"`
	Nodejs                     *string `json:"nodejs,omitempty"`
	Repository                 *string `json:"repository,omitempty"`
	RepositoryURL              *string `json:"repositoryUrl,omitempty"`
	SourceCommit               *string `json:"sourceCommit,omitempty"`
	SourceRef                  *string `json:"sourceRef,omitempty"`
	SourceURL                  *string `json:"sourceUrl,omitempty"`
	ThirdPartyBugFeatureURL    *string `json:"thirdPartyBugFeatureUrl,omitempty"`
	ThirdPartyDocumentationURL *string `json:"thirdPartyDocumentationUrl,omitempty"`
	ThirdPartySourceURL        *string `json:"thirdPartySourceUrl,omitempty"`
	ThirdPartySupportURL       *string `json:"thirdPartySupportUrl,omitempty"`
	Version                    string  `json:"version"`
	VersionURL                 string  `json:"versionUrl"`
	Licensed                   bool    `json:"licensed"`
}

type ServerVersionResponse struct {
	// Major Major version number
	Major int `json:"major"`

	// Minor Minor version number
	Minor int `json:"minor"`

	// Patch Patch version number
	Patch int `json:"patch"`

	// Prerelease Pre-release version number
	Prerelease *int `json:"prerelease"`
}

func (a *Asset) AboutInfo() (ServerAboutResponse, error) {
	var serverAboutResponse ServerAboutResponse

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return serverAboutResponse, err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/server/about",
	}

	apiBody, _, _, err := a.immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return serverAboutResponse, err
	}

	err = json.Unmarshal(apiBody, &serverAboutResponse)
	if err != nil {
		return serverAboutResponse, err
	}

	return serverAboutResponse, nil
}

type ServerPingResponse struct {
	Res string `json:"res"`
}

func IsOnline(ctx context.Context, immichURL string) bool {
	var pong ServerPingResponse

	u, err := url.Parse(immichURL)
	if err != nil {
		return false
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/server/ping",
	}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if reqErr != nil {
		return false
	}

	req.Header.Set("Accept", "application/json")

	res, resErr := HTTPClient.Do(req)
	if resErr != nil {
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Error("Immich server Ping", "status", res.StatusCode)
		return false
	}

	responseBody, responseBodyErr := io.ReadAll(res.Body)
	if responseBodyErr != nil {
		log.Error("reading response body", "method", "GET", "url", apiURL.String(), "err", responseBodyErr)
		return false
	}

	err = json.Unmarshal(responseBody, &pong)
	if err != nil {
		log.Error("Immich server Ping", "err", err)
		return false
	}

	return pong.Res == "pong"
}

func Version(ctx context.Context, immichURL string) (ServerVersionResponse, error) {
	var serverVersionResponse ServerVersionResponse

	u, err := url.Parse(immichURL)
	if err != nil {
		return serverVersionResponse, err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/server/version",
	}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if reqErr != nil {
		return serverVersionResponse, reqErr
	}

	req.Header.Set("Accept", "application/json")

	res, resErr := HTTPClient.Do(req)
	if resErr != nil {
		return serverVersionResponse, resErr
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return serverVersionResponse, fmt.Errorf("status code: %d", res.StatusCode)
	}

	responseBody, responseBodyErr := io.ReadAll(res.Body)
	if responseBodyErr != nil {
		return serverVersionResponse, responseBodyErr
	}

	err = json.Unmarshal(responseBody, &serverVersionResponse)
	if err != nil {
		return serverVersionResponse, err
	}

	return serverVersionResponse, nil
}
