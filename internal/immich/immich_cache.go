// manages the assets cache
package immich

import (
	"net/url"
	"path"

	"github.com/damongolding/immich-kiosk/internal/cache"
)

// RemoveAssetCache deletes the cached data for a specific asset
// identified by its ID for a given device and user combination.
//
// Parameters:
//   - deviceID: string - The unique identifier of the device
//
// Returns:
//   - error - Returns an error if URL parsing fails, nil otherwise
func (a *Asset) RemoveAssetCache(deviceID string) error {
	u, err := url.Parse(a.requestConfig.ImmichURL)
	if err != nil {
		return err
	}

	apiURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   path.Join("api", "assets", a.ID),
	}

	cacheKey := cache.APICacheKey(apiURL.String(), deviceID, a.requestConfig.SelectedUser)
	cache.Delete(cacheKey)

	return nil
}
