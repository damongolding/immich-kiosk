package config

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/charmbracelet/log"
)

// validateConfigFile checks if the given file path is valid and not a directory.
// It returns an error if the file is a directory, and nil if the file doesn't exist.
func validateConfigFile(path string) error {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("Config file is a directory: %s", path)
	}
	return nil
}

// checkUrlScheme checks given url has correct scheme and adds http:// if none is found.
// The function checks for http:// and https:// prefixes in a case-insensitive way.
// If neither prefix is found, it prepends the default scheme (http://).
func (c *Config) checkUrlScheme() {
	// check for correct scheme
	switch {
	case strings.HasPrefix(strings.ToLower(c.ImmichUrl), "http://"):
		break
	case strings.HasPrefix(strings.ToLower(c.ImmichUrl), "https://"):
		break
	default:
		c.ImmichUrl = defaultScheme + c.ImmichUrl
	}
}

// checkLowercaseTaggedFields processes struct fields tagged with `lowercase:"true"`.
// It uses reflection to identify string fields with this tag and converts their
// values to lowercase. This ensures consistent casing for configuration values
// that should be case-insensitive.
func (c *Config) checkLowercaseTaggedFields() {
	val := reflect.ValueOf(c).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check if the field has the `lowercase` tag set to "true"
		if fieldType.Tag.Get("lowercase") == "true" && field.Kind() == reflect.String && field.CanSet() {
			field.SetString(strings.ToLower(field.String()))
		}
	}
}

// checkRequiredFields verifies that all required configuration fields are set.
// Currently checks for:
// - ImmichUrl: The base URL for the Immich server
// - ImmichApiKey: The API key for authentication
// If any required field is missing, the function logs a fatal error and exits.
func (c *Config) checkRequiredFields() {
	switch {
	case c.ImmichUrl == "":
		log.Fatal("Immich Url is missing")
	case c.ImmichApiKey == "":
		log.Fatal("Immich API is missing")
	}
}

// checkDebuging enables the debug flag if verbose debugging is enabled.
// This ensures that verbose debugging also triggers regular debugging output.
func (c *Config) checkDebuging() {
	if c.Kiosk.DebugVerbose {
		c.Kiosk.Debug = true
	}
}

// checkAlbumAndPerson validates and cleans up the Album and Person slices in the Config.
// It removes any empty strings or placeholder values ("ALBUM_ID" or "PERSON_ID"),
// and trims whitespace from the remaining values.
func (c *Config) checkAlbumAndPerson() {
	newAlbum := []string{}
	for _, album := range c.Album {
		if album != "" && album != "ALBUM_ID" {
			newAlbum = append(newAlbum, strings.TrimSpace(album))
		}
	}
	c.Album = newAlbum

	newExcludedAlbums := []string{}
	for _, album := range c.ExcludedAlbums {
		if album != "" && album != "ALBUM_ID" {
			newExcludedAlbums = append(newExcludedAlbums, strings.TrimSpace(album))
		}
	}
	c.ExcludedAlbums = newExcludedAlbums

	newPerson := []string{}
	for _, person := range c.Person {
		if person != "" && person != "PERSON_ID" {
			newPerson = append(newPerson, strings.TrimSpace(person))
		}
	}
	c.Person = newPerson
}

// checkExcludedAlbums filters out any albums from c.Album that are present in
// c.ExcludedAlbums. It uses a map for O(1) lookups of excluded album IDs and
// filters in-place to avoid unnecessary allocations. If the resulting slice's
// capacity is significantly larger than its length, a new slice is allocated
// to prevent memory leaks.
func (c *Config) checkExcludedAlbums() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.ExcludedAlbums) == 0 || len(c.Album) == 0 {
		return
	}

	excludeMap := make(map[string]struct{}, len(c.ExcludedAlbums))
	for _, id := range c.ExcludedAlbums {
		excludeMap[id] = struct{}{}
	}

	filtered := c.Album[:0]
	for _, album := range c.Album {
		if _, excluded := excludeMap[album]; !excluded {
			filtered = append(filtered, album)
		}
	}

	c.Album = filtered

	if excess := cap(c.Album) - len(c.Album); excess > len(c.Album) {
		c.Album = append(make([]string, 0, len(c.Album)), c.Album...)
	}
}

// checkWeatherLocations validates the WeatherLocations in the Config.
// It checks each WeatherLocation for required fields (name, latitude, longitude, and API key),
// and logs an error message if any required fields are missing.
func (c *Config) checkWeatherLocations() {
	for i := 0; i < len(c.WeatherLocations); i++ {
		w := c.WeatherLocations[i]
		missingFields := []string{}
		if w.Name == "" {
			missingFields = append(missingFields, "name")
		}
		if w.Lat == "" {
			missingFields = append(missingFields, "latitude")
		}
		if w.Lon == "" {
			missingFields = append(missingFields, "longitude")
		}
		if w.API == "" {
			missingFields = append(missingFields, "API key")
		}
		if w.Default {
			if c.HasWeatherDefault {
				log.Warn("Multiple default weather locations found. Using the first one.", "name", w.Name)
				w.Default = false
			} else {
				c.HasWeatherDefault = true
			}
		}
		if len(missingFields) > 0 {
			log.Warn("Weather location is missing required fields. Ignoring this location.", "missing fields", strings.Join(missingFields, ", "), "name", w.Name)
			c.WeatherLocations = append(c.WeatherLocations[:i], c.WeatherLocations[i+1:]...)
			i--
		}
	}
}

// checkHideCountries processes the list of countries to hide in location information
// by converting all country names to lowercase for case-insensitive matching.
// If the HideCountries slice is empty, the function returns early without making
// any modifications.
//
// This normalization ensures consistent matching of country names regardless of
// the casing used in the configuration or location data.
func (c *Config) checkHideCountries() {
	if len(c.HideCountries) == 0 {
		return
	}

	for i, country := range c.HideCountries {
		c.HideCountries[i] = strings.ToLower(country)
	}
}

// checkFetchedAssetsSize validates and adjusts the FetchedAssetsSize setting.
// It ensures the value stays within acceptable bounds:
// - Minimum: 1
// - Maximum: 1000
// If the value is outside these bounds, it is clamped to the nearest valid value
// and a warning is logged.
func (c *Config) checkFetchedAssetsSize() {
	if c.Kiosk.FetchedAssetsSize < 1 {
		log.Warn("FetchedAssetsSize too small, setting to minimum value", "value", 1)
		c.Kiosk.FetchedAssetsSize = 1
	} else if c.Kiosk.FetchedAssetsSize > 1000 {
		log.Warn("FetchedAssetsSize too large, setting to maximum value", "value", 1000)
		c.Kiosk.FetchedAssetsSize = 1000
	}
}

// checkRedirects validates and processes the configured redirects in the Config.
// It performs several checks and validations:
// - Skips redirects with empty names or URLs
// - Ensures redirect names are unique
// - Validates URLs are properly formatted
// - Handles relative redirects starting with "?"
// - Detects and removes circular redirects
// - Builds a map for O(1) redirect lookups
//
// The function updates the Config's RedirectsMap field with valid redirects.
// Invalid redirects are logged as warnings and excluded from the final map.
func (c *Config) checkRedirects() {
	redirects := make(map[string]Redirect)
	seen := make(map[string]bool)

	for _, r := range c.Kiosk.Redirects {
		if r.Name == "" {
			log.Warn("Skipping redirect with empty name", "url", r.URL)
			continue
		}
		if r.URL == "" {
			log.Warn("Skipping redirect with empty URL", "name", r.Name)
			continue
		}

		if seen[r.Name] {
			log.Warn("Duplicate redirect name found", "name", r.Name)
			continue
		}
		if _, err := url.Parse(r.URL); err != nil {
			log.Warn("Invalid redirect URL", "name", r.Name, "url", r.URL, "error", err)
			continue
		}
		seen[r.Name] = true

		if strings.HasPrefix(r.URL, "?") {
			r.URL = "/" + r.URL
		}
		redirects[r.Name] = Redirect{
			URL:  r.URL,
			Type: r.Type,
		}

		if c.Kiosk.Debug {
			log.Debug("Registered redirect", "name", r.Name, "url", r.URL)
		}

	}

	for name, targetURL := range redirects {
		visited := make(map[string]bool)
		current := name

		for {
			if visited[current] {
				log.Warn("Circular redirect detected",
					"starting_point", name,
					"current", current,
					"url", targetURL.URL)
				delete(redirects, name)
				break
			}

			visited[current] = true

			// Check if the URL points to another internal redirect
			if strings.HasPrefix(targetURL.URL, "/") {
				nextRedirect := strings.TrimPrefix(targetURL.URL, "/")
				nextURL, exists := redirects[nextRedirect]
				if !exists {
					break
				}
				current = nextRedirect
				targetURL = nextURL
			} else {
				break
			}
		}
	}

	c.Kiosk.RedirectsMap = redirects
}
