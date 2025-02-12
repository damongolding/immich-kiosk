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
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/google/go-querystring/query"
)

type Tags []Tag

func (t Tags) Get(tagValue string) (Tag, error) {

	tagValue, err := url.PathUnescape(tagValue)
	if err != nil {
		return Tag{}, err
	}

	for _, tag := range t {
		if strings.EqualFold(tag.Value, tagValue) {
			return tag, nil
		}
	}

	return Tag{}, fmt.Errorf("tag not found. tag=%s", tagValue)
}

func (i *ImmichAsset) AllTags(requestID, deviceID string) (Tags, string, error) {
	var tags []Tag

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		return immichApiFail(tags, err, nil, "")
	}

	apiUrl := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "tags"),
	}

	immichApiCall := withImmichApiCache(i.immichApiCall, requestID, deviceID, tags)
	body, err := immichApiCall("GET", apiUrl.String(), nil)
	if err != nil {
		return immichApiFail(tags, err, body, apiUrl.String())
	}

	err = json.Unmarshal(body, &tags)
	if err != nil {
		return immichApiFail(tags, err, body, apiUrl.String())
	}

	return tags, apiUrl.String(), nil
}

func (i *ImmichAsset) tagAssets(tag Tag) {}

func (i *ImmichAsset) AssetFromTag(tagValue string, requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random image from tag", tagValue)
	} else {
		log.Debug(requestID+" Getting Random image from tag", tagValue)
	}

	for retries := 0; retries < MaxRetries; retries++ {

		tags, _, err := i.AllTags(requestID, deviceID)
		if err != nil {
			return err
		}

		tag, err := tags.Get(tagValue)
		if err != nil {
			log.Error("getting tag from tags", "err", err)
			continue
		}

		var immichAssets []ImmichAsset

		u, err := url.Parse(requestConfig.ImmichUrl)
		if err != nil {
			return fmt.Errorf("parsing url: %w", err)
		}

		requestBody := ImmichSearchRandomBody{
			Type: string(ImageType),
			// TakenAfter:  dateStart.Format(time.RFC3339),
			// TakenBefore: dateEnd.Format(time.RFC3339),
			TagIDs:     []string{tag.ID},
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
			return fmt.Errorf("marshaling request body: %w", err)
		}

		immichApiCall := withImmichApiCache(i.immichApiCall, requestID, deviceID, immichAssets)
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

		apiCacheKey := cache.ApiCacheKey(apiUrl.String(), deviceID, requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No images left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		for immichAssetIndex, asset := range immichAssets {

			if !asset.isValidAsset(ImageOnlyAssetTypes) {
				continue
			}

			err := asset.AssetInfo(requestID, deviceID)
			if err != nil {
				log.Error("Failed to get additional asset data", "error", err)
			}

			if asset.containsTag(kiosk.TagSkip) {
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

				// replace cache with used image(s) removed
				err = cache.Replace(apiCacheKey, jsonBytes)
				if err != nil {
					log.Debug("Failed to update cache", "error", err, "url", apiUrl.String())
				}
			}

			asset.Bucket = kiosk.SourceTag
			asset.BucketID = tag.ID

			*i = asset

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("No images found for '%s'. Max retries reached.", tagValue)

}
