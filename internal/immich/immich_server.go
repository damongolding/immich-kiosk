package immich

import (
	"encoding/json"
	"net/http"
	"net/url"
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
	Licensed                   bool    `json:"licensed"`
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

	apiBody, err := a.immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return serverAboutResponse, err
	}

	err = json.Unmarshal(apiBody, &serverAboutResponse)
	if err != nil {
		return serverAboutResponse, err
	}

	return serverAboutResponse, nil
}
