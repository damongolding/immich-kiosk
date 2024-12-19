package immich

import (
	"net/url"
	"path"

	"github.com/charmbracelet/log"
)

func (i *ImmichAsset) VideoPlayback() []byte {

	var bytes []byte

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Error(err)
		return bytes
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", i.ID, "video", "playback"),
	}

	headers := map[string]string{"Accept": "application/octet-stream"}

	bytes, err = i.immichApiCall("GET", apiUrl.String(), nil, headers)
	if err != nil {
		log.Error(err)
	}

	return bytes

	// ext := path.Ext(i.OriginalPath)

	// fileName := fmt.Sprintf("%s.%s", i.ID, ext)
	// err = os.WriteFile(fileName, bytes, 0644)
	// if err != nil {
	// 	log.Error("Failed to write video file:", err)
	// 	return
	// }

}
