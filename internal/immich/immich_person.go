package immich

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"path"
	"slices"
	"sync/atomic"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/google/go-querystring/query"
	"golang.org/x/sync/errgroup"
)

func (a *Asset) people(requestID, deviceID string, knowPeopleOnly bool, bypassCache bool) ([]Person, error) {
	var people []Person
	page := 1

	for {

		if page > MaxPages {
			log.Warn(requestID + " Reached maximum page count when fetching people")
			break
		}

		var allPeople AllPeopleResponse

		u, err := url.Parse(a.requestConfig.ImmichURL)
		if err != nil {
			_, _, err = immichAPIFail(allPeople, err, nil, "")
			return people, err
		}

		apiURL := url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     path.Join("api", "people"),
			RawQuery: fmt.Sprintf("page=%d", page),
		}

		var body []byte

		if bypassCache {
			body, _, err = a.immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
			if err != nil {
				_, _, err = immichAPIFail(allPeople, err, body, apiURL.String())
				return people, err
			}
		} else {
			immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, allPeople)
			body, _, err = immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
			if err != nil {
				_, _, err = immichAPIFail(allPeople, err, body, apiURL.String())
				return people, err
			}
		}

		err = json.Unmarshal(body, &allPeople)
		if err != nil {
			_, _, err = immichAPIFail(allPeople, err, body, apiURL.String())
			return people, err
		}

		people = append(people, allPeople.People...)

		if !allPeople.HasNextPage {
			break
		}

		page++

	}

	if knowPeopleOnly {
		namedPeople := make([]Person, 0, len(people))
		for _, person := range people {
			if person.Name != "" {
				namedPeople = append(namedPeople, person)
			}
		}
		people = namedPeople
	}

	return people, nil
}

func (a *Asset) AllNamedPeople(requestID, deviceID string) ([]Person, error) {
	return a.people(requestID, deviceID, true, false)
}

// allPeopleAssetCount returns the total count of assets across all named people in the system.
// It performs concurrent queries for each person's asset count using a limited number of goroutines.
//
// Parameters:
//   - requestID: The ID of the API request for tracking purposes
//   - deviceID: The ID of the device making the request
//
// Returns:
//   - int: The total number of assets across all named people
//   - error: nil if successful, error if the people query fails or if any individual count fails
func (a *Asset) allPeopleAssetCount(requestID, deviceID string) (int, error) {
	allPeople, allPeopleErr := a.people(requestID, deviceID, true, false)
	if allPeopleErr != nil {
		return 0, allPeopleErr
	}

	var counts atomic.Int64
	var errGroup errgroup.Group
	errGroup.SetLimit(20)

	for _, person := range allPeople {
		p := person
		errGroup.Go(func() error {
			count, err := a.PersonAssetCount(p.ID, requestID, deviceID)
			if err != nil {
				log.Error(requestID+" Failed to count assets for person", "personID", p.ID, "error", err)
				return err
			}
			counts.Add(int64(count))
			return nil
		})

	}

	errGroupErr := errGroup.Wait()
	if errGroupErr != nil {
		return 0, errGroupErr
	}

	return int(counts.Load()), nil
}

// PersonAssetCount returns the number of assets associated with a specific person in Immich.
func (a *Asset) PersonAssetCount(personID, requestID, deviceID string) (int, error) {

	if personID == kiosk.PersonKeywordAll {
		return a.allPeopleAssetCount(requestID, deviceID)
	}

	var personStatistics PersonStatistics

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(personStatistics, err, nil, "")
		return 0, err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "people", personID, "statistics"),
	}

	immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, personStatistics)
	body, _, err := immichAPICall(a.ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		_, _, err = immichAPIFail(personStatistics, err, body, apiURL.String())
		return 0, err
	}

	err = json.Unmarshal(body, &personStatistics)
	if err != nil {
		_, _, err = immichAPIFail(personStatistics, err, body, apiURL.String())
		return 0, err
	}

	return personStatistics.Assets, nil
}

// RandomAssetOfPerson retrieves a random asset for a given person from the Immich API.
// It handles retries, caching, and filtering to find suitable assets. The function will make
// multiple attempts to find a valid asset that matches the criteria (not trashed, correct type, etc).
// If caching is enabled, it will maintain a cache of unused assets for future requests.
//
// Parameters:
//   - personID: The ID of the person whose assets to search for
//   - requestID: The ID of the API request for tracking purposes
//   - deviceID: The ID of the device making the request
//
// Returns:
//   - error: nil if successful, error otherwise. Returns specific error if no suitable
//     asset is found after MaxRetries attempts or if there are API/parsing failures
//
// The function mutates the receiver (i *ImmichAsset) to store the selected asset if successful.
func (a *Asset) RandomAssetOfPerson(personID, requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random asset of", personID)
	} else {
		log.Debug(requestID+" Getting Random asset of", personID)
	}

	for range MaxRetries {

		var immichAssets []Asset

		u, err := url.Parse(a.requestConfig.ImmichURL)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, nil, "")
			return err
		}

		requestBody := SearchRandomBody{
			PersonIDs:  []string{personID},
			Type:       string(ImageType),
			WithExif:   true,
			WithPeople: true,
			Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
		}

		// Include videos if show videos is enabled
		if a.requestConfig.ShowVideos {
			requestBody.Type = ""
		}

		if a.requestConfig.RequireAllPeople {
			requestBody.PersonIDs = make([]string, len(a.requestConfig.People))
			copy(requestBody.PersonIDs, a.requestConfig.People)
		}

		if a.requestConfig.ShowArchived {
			requestBody.WithArchived = true
		}

		DateFilter(&requestBody, a.requestConfig.DateFilter)

		// convert body to queries so url is unique and can be cached
		queries, _ := query.Values(requestBody)

		apiURL := url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     "api/search/random",
			RawQuery: fmt.Sprintf("kiosk=%x", sha256.Sum256([]byte(queries.Encode()))),
		}

		jsonBody, bodyMarshalErr := json.Marshal(requestBody)
		if bodyMarshalErr != nil {
			_, _, bodyMarshalErr = immichAPIFail(immichAssets, bodyMarshalErr, nil, apiURL.String())
			return bodyMarshalErr
		}

		immichAPICall := withImmichAPICache(a.immichAPICall, requestID, deviceID, a.requestConfig, immichAssets)
		apiBody, _, err := immichAPICall(a.ctx, http.MethodPost, apiURL.String(), jsonBody)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, apiBody, apiURL.String())
			return err
		}

		err = json.Unmarshal(apiBody, &immichAssets)
		if err != nil {
			_, _, err = immichAPIFail(immichAssets, err, apiBody, apiURL.String())
			return err
		}

		apiCacheKey := cache.APICacheKey(apiURL.String(), deviceID, a.requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No assets left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		wantedAssetType := ImageOnlyAssetTypes
		if a.requestConfig.ShowVideos {
			wantedAssetType = AllAssetTypes
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourcePerson
			asset.requestConfig = a.requestConfig
			asset.ctx = a.ctx

			if !asset.isValidAsset(requestID, deviceID, wantedAssetType, a.RatioWanted) {
				continue
			}

			if a.requestConfig.Kiosk.Cache {
				// Remove the current asset from the slice
				immichAssetsToCache := slices.Delete(immichAssets, immichAssetIndex, immichAssetIndex+1)
				jsonBytes, cacheMarshalErr := json.Marshal(immichAssetsToCache)
				if cacheMarshalErr != nil {
					log.Error("Failed to marshal immichAssetsToCache", "error", cacheMarshalErr)
					return cacheMarshalErr
				}

				// Replace cache with remaining assets after removing used asset(s)
				cache.Set(apiCacheKey, jsonBytes, a.requestConfig.Duration)
			}

			asset.BucketID = personID

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable assets left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}
	return fmt.Errorf("no assets found for person '%s'. Max retries reached", personID)
}

// RandomPersonFromAllPeople returns a random person ID from all people in the system.
// It can optionally filter to only include people who have been given names.
//
// Parameters:
//   - requestID: The ID of the API request for tracking purposes
//   - deviceID: The ID of the device making the request
//   - knowPeopleOnly: If true, only returns people who have been given names
//
// Returns:
//   - string: The ID of the randomly selected person
//   - error: nil if successful, error if no people are found or if the API call fails
func (a *Asset) RandomPersonFromAllPeople(requestID, deviceID string, knowPeopleOnly bool) (string, error) {

	people, err := a.people(requestID, deviceID, knowPeopleOnly, false)
	if err != nil {
		return "", fmt.Errorf("failed to get people: %w", err)
	}

	if len(people) == 0 {
		return "", errors.New("no valid people found with names")
	}

	if len(a.requestConfig.ExcludedPeople) > 0 {
		people = slices.DeleteFunc(people, func(person Person) bool {
			return slices.Contains(a.requestConfig.ExcludedPeople, person.ID)
		})
	}

	if len(people) == 0 {
		return "", errors.New("no valid people found after exclusions")
	}

	picked := people[rand.IntN(len(people))]

	return picked.ID, nil
}
