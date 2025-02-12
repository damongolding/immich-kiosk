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
		if strings.EqualFold(tag.Value, tagValue) || strings.EqualFold(tag.ID, tagValue) {
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

func (i *ImmichAsset) AssetsWithTagCount(tagID string, requestID, deviceID string) (int, error) {

	var allAssetsCount int
	pageCount := 1

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		_, _, err = immichApiFail(allAssetsCount, err, nil, "")
		return allAssetsCount, err
	}

	requestBody := ImmichSearchRandomBody{
		Type:       string(ImageType),
		TagIDs:     []string{tagID},
		WithPeople: false,
		WithExif:   false,
		Size:       requestConfig.Kiosk.FetchedAssetsSize,
	}

	if requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	DateFilter(&requestBody, requestConfig.DateFilter)

	for {

		var taggedAssets ImmichSearchMetadataResponse

		requestBody.Page = pageCount

		// convert body to queries so url is unique and can be cached
		queries, _ := query.Values(requestBody)

		apiUrl := url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     "api/search/metadata",
			RawQuery: queries.Encode(),
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			_, _, err = immichApiFail(allAssetsCount, err, nil, apiUrl.String())
			return allAssetsCount, err
		}

		immichApiCall := withImmichApiCache(i.immichApiCall, requestID, deviceID, taggedAssets)
		apiBody, err := immichApiCall("POST", apiUrl.String(), jsonBody)
		if err != nil {
			_, _, err = immichApiFail(taggedAssets, err, apiBody, apiUrl.String())
			return allAssetsCount, err
		}

		err = json.Unmarshal(apiBody, &taggedAssets)
		if err != nil {
			_, _, err = immichApiFail(taggedAssets, err, apiBody, apiUrl.String())
			return allAssetsCount, err
		}

		allAssetsCount += taggedAssets.Assets.Total

		if taggedAssets.Assets.NextPage == "" {
			break
		}

		pageCount++
	}

	return allAssetsCount, nil
}

func (i *ImmichAsset) AssetsWithTag(tagID string, requestID, deviceID string) ([]ImmichAsset, string, error) {

	var immichAssets []ImmichAsset

	u, err := url.Parse(requestConfig.ImmichUrl)
	if err != nil {
		return immichApiFail(immichAssets, err, nil, "")
	}

	requestBody := ImmichSearchRandomBody{
		Type:       string(ImageType),
		TagIDs:     []string{tagID},
		WithExif:   true,
		WithPeople: true,
		Size:       requestConfig.Kiosk.FetchedAssetsSize,
	}

	if requestConfig.ShowArchived {
		requestBody.WithArchived = true
	}

	DateFilter(&requestBody, requestConfig.DateFilter)

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
		return immichApiFail(immichAssets, err, nil, apiUrl.String())
	}

	immichApiCall := withImmichApiCache(i.immichApiCall, requestID, deviceID, immichAssets)
	apiBody, err := immichApiCall("POST", apiUrl.String(), jsonBody)
	if err != nil {
		return immichApiFail(immichAssets, err, nil, apiUrl.String())
	}

	err = json.Unmarshal(apiBody, &immichAssets)
	if err != nil {
		return immichApiFail(immichAssets, err, nil, apiUrl.String())
	}

	return immichAssets, apiUrl.String(), nil
}

func (i *ImmichAsset) RandomAssetWithTag(tagID string, requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random image with tag", tagID)
	} else {
		log.Debug(requestID+" Getting Random image with tag", tagID)
	}

	for retries := 0; retries < MaxRetries; retries++ {

		immichAssets, apiUrl, err := i.AssetsWithTag(tagID, requestID, deviceID)
		if err != nil {
			return err
		}

		apiCacheKey := cache.ApiCacheKey(apiUrl, deviceID, requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No images left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)

			immichAssets, _, retryErr := i.AssetsWithTag(tagID, requestID, deviceID)
			if retryErr != nil || len(immichAssets) == 0 {
				return fmt.Errorf("no assets found with tag %s after refresh", tagID)
			}

			continue
		}

		for immichAssetIndex, asset := range immichAssets {

			if !asset.isValidAsset(ImageOnlyAssetTypes) {
				continue
			}

			log.Info("After check", "watnted", asset.RatioWanted, "l", asset.IsLandscape, "p", asset.IsPortrait)

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
					log.Debug("Failed to update device cache for tag", "tagID", tagID, "deviceID", deviceID)
				}
			}

			asset.Bucket = kiosk.SourceTag
			asset.BucketID = tagID

			*i = asset

			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}

	return fmt.Errorf("No images found for '%s'. Max retries reached.", tagID)
}
