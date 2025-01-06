package immich

import (
	"encoding/json"
	"net/url"

	"github.com/charmbracelet/log"
)

func (i *ImmichAsset) CheckForFaces(requestID, deviceID string) {

	var faces []Face

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "api/faces",
		RawQuery: "id=" + i.ID,
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, deviceID, faces)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		_, err = immichApiFail(faces, err, body, apiUrl.String())
		log.Error("adding faces", "err", err)
		return
	}

	err = json.Unmarshal(body, &faces)
	if err != nil {
		_, err = immichApiFail(faces, err, body, apiUrl.String())
		log.Error("adding faces", "err", err)
		return
	}

	p := Person{
		Faces: faces,
	}

	i.People = append(i.People, p)

}
