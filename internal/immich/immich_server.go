package immich

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type ServerAboutResponse struct {
	Build                      *string `json:"build,omitempty"`
	BuildImage                 *string `json:"buildImage,omitempty"`
	BuildImageUrl              *string `json:"buildImageUrl,omitempty"`
	BuildUrl                   *string `json:"buildUrl,omitempty"`
	Exiftool                   *string `json:"exiftool,omitempty"`
	Ffmpeg                     *string `json:"ffmpeg,omitempty"`
	Imagemagick                *string `json:"imagemagick,omitempty"`
	Libvips                    *string `json:"libvips,omitempty"`
	Licensed                   bool    `json:"licensed"`
	Nodejs                     *string `json:"nodejs,omitempty"`
	Repository                 *string `json:"repository,omitempty"`
	RepositoryUrl              *string `json:"repositoryUrl,omitempty"`
	SourceCommit               *string `json:"sourceCommit,omitempty"`
	SourceRef                  *string `json:"sourceRef,omitempty"`
	SourceUrl                  *string `json:"sourceUrl,omitempty"`
	ThirdPartyBugFeatureUrl    *string `json:"thirdPartyBugFeatureUrl,omitempty"`
	ThirdPartyDocumentationUrl *string `json:"thirdPartyDocumentationUrl,omitempty"`
	ThirdPartySourceUrl        *string `json:"thirdPartySourceUrl,omitempty"`
	ThirdPartySupportUrl       *string `json:"thirdPartySupportUrl,omitempty"`
	Version                    string  `json:"version"`
	VersionUrl                 string  `json:"versionUrl"`
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

	return serverAboutResponse, err
}
