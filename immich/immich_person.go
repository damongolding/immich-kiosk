package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/google/go-querystring/query"
	"github.com/patrickmn/go-cache"
)

// DEPRECIATED
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
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &albums)
	if err != nil {
		return immichApiFail(albums, err, body, apiUrl.String())
	}

	return albums, nil
}

// DEPRECIATED
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
	body, err := immichApiCal("GET", apiUrl.String(), nil)
	if err != nil {
		return immichApiFail(images, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &images)
	if err != nil {
		return immichApiFail(images, err, body, apiUrl.String())
	}

	return images, nil
}

// DEPRECIATED
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
	body, err := immichApiCall("GET", apiUrl.String(), nil)
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

// DEPRECIATED
// RandomImageOfPerson retrieve random image of person from Immich
func (i *ImmichAsset) OLDRandomImageOfPerson(personID, requestID, kioskDeviceID string, isPrefetch bool) error {

	images, err := i.personAssets(personID, requestID)
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
					log.Debug(requestID, "PREFETCH", kioskDeviceID, "Got image of person", per.Name)
				} else {
					log.Debug(requestID, "Got image of person", per.Name)
				}

				break
			}
		}
	}

	return nil
}

// RandomImageOfPerson retrieve random image of person from Immich
func (i *ImmichAsset) RandomImageOfPerson(personID, requestID, kioskDeviceID string, isPrefetch bool) error {

	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal("parsing url", err)
	}

	requestBody := ImmichSearchBody{
		PersonIds: []string{personID},
		Type:      string(ImageType),
		WithExif:  true,
		Size:      1000,
	}

	if requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	// convert body to queries so url is unique and can be cached
	queries, _ := query.Values(requestBody)

	apiUrl := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     "api/search/random",
		RawQuery: queries.Encode(),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal("marshaling request body", err)
	}

	requestBodyReader := bytes.NewReader(jsonBody)

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, immichAssets)
	apiBody, err := immichApiCall("POST", apiUrl.String(), requestBodyReader)
	if err != nil {
		_, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
		return err
	}

	err = json.Unmarshal(apiBody, &immichAssets)
	if err != nil {
		_, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
		return err
	}

	if len(immichAssets) == 0 {
		log.Debug(requestID + " No images left in cache. Refreshing and trying again")
		apiCache.Delete(apiUrl.String())
		return i.RandomImageOfPerson(personID, requestID, kioskDeviceID, isPrefetch)
	}

	for immichAssetIndex, img := range immichAssets {
		// We only want images and that are not trashed or archived (unless wanted by user)
		if img.Type != ImageType || img.IsTrashed || (img.IsArchived && !requestConfig.ShowArchived) || !i.ratioCheck(&img) {
			continue
		}

		if requestConfig.Kiosk.Cache {
			// Remove the current image from the slice
			immichAssetsToCache := append(immichAssets[:immichAssetIndex], immichAssets[immichAssetIndex+1:]...)
			jsonBytes, err := json.Marshal(immichAssetsToCache)
			if err != nil {
				log.Error("Failed to marshal immichAssetsToCache", "error", err)
				return err
			}

			// replace cwith cache minus used image
			err = apiCache.Replace(apiUrl.String(), jsonBytes, cache.DefaultExpiration)
			if err != nil {
				log.Debug("cache not found!")
			}
		}

		*i = img
		return nil
	}

	log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
	apiCache.Delete(apiUrl.String())
	return i.RandomImageOfPerson(personID, requestID, kioskDeviceID, isPrefetch)
}
