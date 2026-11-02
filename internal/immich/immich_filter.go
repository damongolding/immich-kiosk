package immich

import (
	"encoding/json"
	"errors"
	"net/url"
	"slices"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
)

// ValidateRawFilter checks whether f can be decoded as a search filter.
// It returns JSON syntax or decoding errors; it does not check the filter with Immich.
func ValidateRawFilter(f string) error {
	_, err := validateRawFilter(f)
	return err
}

// validateRawFilter decodes f into a search filter with default image and timeline
// visibility settings. On a decoding error, it returns the partially decoded filter
// and the error.
func validateRawFilter(f string) (*SearchFilter, error) {
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

// RawFilterCount returns the number of assets fetched for the configured filter ID,
// up to the metadata pagination limit. It selects the filter's user for subsequent
// requests and returns URL, filter lookup, decoding, or metadata request errors.
func (a *Asset) RawFilterCount(requestID, deviceID string, id string) (int, error) {
	var rawFilter config.RawFilter
	var filter *SearchFilter

	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		_, _, err = immichAPIFail(0, err, nil, "")
		return 0, err
	}

	rawFilter, err = a.requestConfig.RawFilters.GetByID(id)
	if err != nil {
		_, _, err = immichAPIFail(0, err, nil, "")
		return 0, err
	}

	filter, err = validateRawFilter(rawFilter.Filter)
	if err != nil {
		return 0, err
	}

	a.requestConfig.SelectedUser = rawFilter.User

	requestBody := SearchRandomBody{
		Filter:     filter.Build(),
		WithPeople: false,
		WithExif:   false,
		Size:       a.requestConfig.Kiosk.FetchedAssetsSize,
	}

	res, assetsErr := a.fetchPaginatedMetadata(u, requestBody, requestID, deviceID)
	if assetsErr != nil {
		return 0, assetsErr
	}

	if !res.FromCache {
		log.Debug(requestID+" RawFilterCount", "id", rawFilter.ID, "name", rawFilter.Name, "user", rawFilter.User, "assets matched", len(res.Assets))
	}

	return len(res.Assets), nil
}

// RandomAssetFromRawFilter selects a viable asset for the configured filter ID
// and replaces a with it. It selects the filter's user and may update or clear
// cached assets. Lookup, decoding, fetch, and cache serialization errors are
// returned, as is an error when no viable asset is found within MaxRetries.
func (a *Asset) RandomAssetFromRawFilter(requestID, deviceID, id string) error {
	var rawFiler *SearchFilter

	f, err := a.requestConfig.RawFilters.GetByID(id)
	if err != nil {
		return err
	}

	rawFiler, err = validateRawFilter(f.Filter)
	if err != nil {
		return err
	}

	a.requestConfig.SelectedUser = f.User

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

			asset.BucketID = id

			*a = asset

			return nil
		}

		log.Debug(requestID + " No viable assets left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}
	return errors.New("no assets found using your raw filter. Max retries reached")
}
