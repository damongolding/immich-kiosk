package immich

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"

	"github.com/charmbracelet/log"
)

// personAssets retrieves all assets associated with a specific person from Immich.
func (i *ImmichAsset) personAssets(personId, requestId string) ([]ImmichAsset, error) {

	var images []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personId + "/assets",
	}

	immichApiCal := immichApiCallDecorator(i.immichApiCall, requestId, images)
	body, err := immichApiCal(apiUrl.String())
	if err != nil {
		return immichApiFail(images, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &images)
	if err != nil {
		return immichApiFail(images, err, body, apiUrl.String())
	}

	for _, image := range images {
		image.AddRatio()
	}

	return images, nil
}

// RandomImageOfPerson retrieve random image of person from Immich
func (i *ImmichAsset) RandomImageOfPerson(personId, requestId, kioskDeviceID string, isPrefetch bool) error {

	images, err := i.personAssets(personId, requestId)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	rand.Shuffle(len(images), func(i, j int) {
		images[i], images[j] = images[j], images[i]
	})

	for _, pick := range images {
		// We only want images and that are not trashed or archived (unless wanted by user)
		if pick.Type != "IMAGE" || pick.IsTrashed || (pick.IsArchived && !requestConfig.ShowArchived) {
			continue
		}

		// is a specific ratio wanted?
		if i.RatioWanted == "" && i.RatioWanted != i.ExifInfo.Ratio {
			continue
		}

		*i = pick
		break
	}

	if i.ID == "" {
		log.Error("no images found")
		return fmt.Errorf("no images found")
	}

	if log.GetLevel() == log.DebugLevel {
		for _, per := range i.People {
			if per.ID == personId {

				if isPrefetch {
					log.Debug(requestId, "PREFETCH", kioskDeviceID, "Got image of person", per.Name)
				} else {
					log.Debug(requestId, "Got image of person", per.Name)
				}

				break
			}
		}
	}

	return nil
}
