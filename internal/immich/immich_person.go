package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/google/go-querystring/query"
)

// PersonImageCount returns the number of images associated with a specific person in Immich.
func (i *ImmichAsset) PersonImageCount(personID, requestID, deviceID string) (int, error) {

	var personStatistics ImmichPersonStatistics

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal(err)
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "people", personID, "statistics"),
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, deviceID, personStatistics)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		_, _, err = immichApiFail(personStatistics, err, body, apiUrl.String())
		return 0, err
	}

	err = json.Unmarshal(body, &personStatistics)
	if err != nil {
		_, _, err = immichApiFail(personStatistics, err, body, apiUrl.String())
		return 0, err
	}

	return personStatistics.Assets, err
}

// RandomImageOfPerson retrieve random image of person from Immich
func (i *ImmichAsset) RandomImageOfPerson(personID, requestID, deviceID string, isPrefetch bool) error {

	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		log.Fatal("parsing url", err)
	}

	requestBody := ImmichSearchRandomBody{
		PersonIds:  []string{personID},
		Type:       string(ImageType),
		WithExif:   true,
		WithPeople: true,
		Size:       requestConfig.Kiosk.FetchedAssetsSize,
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
		RawQuery: fmt.Sprintf("kiosk=%x", sha256.Sum256([]byte(queries.Encode()))),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal("marshaling request body", err)
	}

	immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, deviceID, immichAssets)
	apiBody, err := immichApiCall("POST", apiUrl.String(), jsonBody)
	if err != nil {
		_, _, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
		return err
	}

	err = json.Unmarshal(apiBody, &immichAssets)
	if err != nil {
		_, _, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
		return err
	}

	apiCacheKey := cache.ApiCacheKey(apiUrl.String(), deviceID)

	if len(immichAssets) == 0 {
		log.Debug(requestID + " No images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
		return i.RandomImageOfPerson(personID, requestID, deviceID, isPrefetch)
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

			// Replace cache with remaining images after removing used image(s)
			err = cache.Replace(apiCacheKey, jsonBytes)
			if err != nil {
				log.Debug("cache not found!")
			}
		}

		*i = img

		i.PersonName(personID)

		return nil
	}

	log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
	cache.Delete(apiCacheKey)
	return i.RandomImageOfPerson(personID, requestID, deviceID, isPrefetch)
}

func (i *ImmichAsset) PersonName(personID string) {
	for _, person := range i.People {
		if strings.EqualFold(person.ID, personID) {
			i.KioskSourceName = person.Name
		}
	}
}
