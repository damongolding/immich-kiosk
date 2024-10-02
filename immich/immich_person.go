package immich

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"

	"github.com/charmbracelet/log"
)

func (i *ImmichAsset) people(requestID string, shared bool) (ImmichAlbums, error) {
	var albums ImmichAlbums

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people",
	}

	if shared {
		apiUrl.RawQuery = "shared=true"
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, albums)
	body, err := immichApiCall(apiUrl.String())
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &albums)
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	return albums, nil
}

// personAssets retrieves all assets associated with a specific person from Immich.
func (i *ImmichAsset) personAssets(personID, requestID string) ([]ImmichAsset, error) {

	var images []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personID + "/assets",
	}

	immichApiCal := immichApiCallDecorator(i.immichApiCall, requestID, images)
	body, err := immichApiCal(apiUrl.String())
	if err != nil {
		return immichApiFail(images, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &images)
	if err != nil {
		return immichApiFail(images, err, body, apiUrl.String())
	}

	return images, nil
}

// PersonImageCount returns the number of images associated with a specific person in Immich.
func (i *ImmichAsset) PersonImageCount(personID, requestID string) (int, error) {

	var personStatistics ImmichPersonStatistics

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "api/people/" + personID + "/statistics",
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, personStatistics)
	body, err := immichApiCall(apiUrl.String())
	if err != nil {
		_, err = immichApiFail(personStatistics, err, body, apiUrl.String())
		return 0, err
	}

	err = json.Unmarshal(body, &personStatistics)
	if err != nil {
		_, err = immichApiFail(personStatistics, err, body, apiUrl.String())
		return 0, err
	}

	return personStatistics.Assets, err
}

// RandomImageOfPerson retrieve random image of person from Immich
func (i *ImmichAsset) RandomImageOfPerson(personID, requestId, kioskDeviceID string, isPrefetch bool) error {

	images, err := i.personAssets(personID, requestId)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		log.Error("no images found", "for person", personID)
		return fmt.Errorf("no images found for person %s", personID)
	}

	rand.Shuffle(len(images), func(i, j int) {
		images[i], images[j] = images[j], images[i]
	})

	for _, pick := range images {
		// Filter out non-image assets, trashed, archived (unless configured), and incorrect ratio
		if pick.Type != ImageType || pick.IsTrashed || (pick.IsArchived && !requestConfig.ShowArchived) || !i.ratioCheck(&pick) {
			continue
		}

		*i = pick
		break
	}

	if i.ID == "" {
		log.Error("no images found", "for person", personID)
		return fmt.Errorf("no images found for person %s", personID)
	}

	if log.GetLevel() == log.DebugLevel {
		for _, per := range i.People {
			if per.ID == personID {

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
