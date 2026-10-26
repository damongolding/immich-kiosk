package immich

import (
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
)

func checkRawFilter(f string) (*SearchFilter, error) {
	filter := NewSearchFilterBuilder()

	b := []byte(f)

	err := json.Unmarshal(b, filter)
	if err != nil {
		if syn, ok := errors.AsType[*json.SyntaxError](err); ok {
			start := max(0, int(syn.Offset)-30)
			end := min(len(b), int(syn.Offset)+10)
			log.Error("RawFilter", "err", err, "offset", syn.Offset, "near", string(b[start:end]))
		} else {
			log.Error("RawFilter", "err", err)
		}
		return filter, err
	}

	return filter, nil
}

func (a *Asset) RawFilterCount(requestID, deviceID string, index int) (int, error) {
	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(0, err, nil, "")
		return 0, err
	}

	rawFiler, err := checkRawFilter(a.requestConfig.RawFilters[index].Filter)
	if err != nil {
		return 0, err
	}

	a.requestConfig.SelectedUser = a.requestConfig.RawFilters[index].User

	requestBody := SearchRandomBody{
		Filter:     rawFiler.Build(),
		WithPeople: false,
		WithExif:   false,
		Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
	}

	res, assetsErr := a.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)
	if assetsErr != nil {
		return 0, assetsErr
	}

	log.Debug(requestID+" RawFilterCount", "name", a.requestConfig.RawFilters[index].Name, "user", a.requestConfig.RawFilters[index].User, "assets matched", len(res.Assets))

	return len(res.Assets), nil
}

func indexFromID(s string) (int, error) {
	_, a, found := strings.Cut(s, "#")
	if !found {
		return 0, errors.New("malformed ID")
	}

	i, err := strconv.ParseInt(a, 0, 0)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

func (a *Asset) RandomAssetFromRawFilter(requestID, deviceID, id string) error {
	var rawFiler *SearchFilter

	i, err := indexFromID(id)
	if err != nil {
		return err
	}

	rawFiler, err = checkRawFilter(a.requestConfig.RawFilters[i].Filter)
	if err != nil {
		return err
	}

	a.requestConfig.SelectedUser = a.requestConfig.RawFilters[i].User

	for range MaxRetries {

		requestBody := SearchRandomBody{
			Filter:     rawFiler.Build(),
			WithPeople: true,
			WithExif:   true,
			Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
		}

		immichAssets, apiURL, fecthErr := a.fetchAssets(requestID, deviceID, requestBody)
		if fecthErr != nil {
			return fecthErr
		}

		apiCacheKey := cache.APICacheKey(apiURL.String(), deviceID, a.requestConfig.SelectedUser)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No assets left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		for immichAssetIndex, asset := range immichAssets {

			asset.Bucket = kiosk.SourceRawFilter
			asset.requestConfig = a.requestConfig
			asset.ctx = a.ctx

			if !asset.isValidAsset(requestID, deviceID, AllAssetTypes, a.RatioWanted) {
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
				cache.Set(apiCacheKey, jsonBytes, a.requestConfig.Duration, a.requestConfig.CacheDuration)
			}

			asset.BucketID = "RAW_FILTER"

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable assets left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}
	return errors.New("no assets found using your raw filter. Max retries reached")
}
