package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/xeipuuv/gojsonschema"
)

var (
	SchemaJSON string
)

// IsSchemaLoaded returns true if the schema has been initialized
func IsSchemaLoaded() bool {
	return SchemaJSON != ""
}

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

// checkURLScheme checks given url has correct scheme and adds http:// if none is found.
// The function checks for http:// and https:// prefixes in a case-insensitive way.
// If neither prefix is found, it prepends the default scheme (http://).
func (c *Config) checkURLScheme() {
	// check for correct scheme
	switch {
	case strings.HasPrefix(strings.ToLower(c.ImmichURL), "http://"):
		break
	case strings.HasPrefix(strings.ToLower(c.ImmichURL), "https://"):
		break
	default:
		c.ImmichURL = defaultScheme + c.ImmichURL
	}
}

// checkLowercaseTaggedFields processes struct fields tagged with `lowercase:"true"`.
// It uses reflection to identify string fields with this tag and converts their
// values to lowercase. This ensures consistent casing for configuration values
// that should be case-insensitive.
func (c *Config) checkLowercaseTaggedFields() {
	val := reflect.ValueOf(c).Elem()
	typ := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check if the field has the `lowercase` tag set to "true"
		if fieldType.Tag.Get("lowercase") == "true" && field.CanSet() {
			switch field.Kind() {
			case reflect.String:
				field.SetString(strings.ToLower(field.String()))

			case reflect.Slice:
				if field.Type().Elem().Kind() == reflect.String {
					sliceLen := field.Len()
					for j := range sliceLen {
						field.Index(j).SetString(strings.ToLower(field.Index(j).String()))
					}
				}
			}
		}
	}
}

// loadSecretFromFile attempts to read and return a secret from the specified file
func loadSecretFromFile(filePath string) (string, bool) {
	data, readErr := os.ReadFile(filePath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			log.Warn("Secret file not found", "file", filePath)
		} else {
			log.Error("Failed to read secret file", "file", filePath, "error", readErr)
		}
		return "", false
	}

	value := strings.TrimSpace(string(data))
	if value == "" {
		log.Warn("Secret file is empty", "file", filePath)
		return "", false
	}

	return value, true
}

func (c *Config) checkSecrets() {

	apiKeyFile := os.Getenv(apiKeyFileEnv)
	if apiKeyFile != "" {
		apiKeyFile = filepath.Clean(apiKeyFile)
		if apiKey, ok := loadSecretFromFile(apiKeyFile); ok {
			log.Info("Loaded Immich API key", "source", "docker secret")
			c.ImmichAPIKey = apiKey
		}
	}

	passwordFile := os.Getenv(passwordFileEnv)
	if passwordFile != "" {
		passwordFile = filepath.Clean(passwordFile)
		if password, ok := loadSecretFromFile(passwordFile); ok {
			log.Info("Loaded password", "source", "docker secret")
			c.Kiosk.Password = password
		}
	}

	weatherAPIFile := os.Getenv(weatherAPIKeyFileEnv)
	if weatherAPIFile != "" {
		weatherAPIFile = filepath.Clean(weatherAPIFile)
		if weatherAPIKey, ok := loadSecretFromFile(weatherAPIFile); ok {
			log.Info("Loaded weather API key", "source", "docker secret")
			for i, location := range c.WeatherLocations {
				if location.API == "" {
					log.Info("Added weather API key to", "location", location.Name)
					c.WeatherLocations[i].API = weatherAPIKey
				}
			}
		}
	}

	credsDir := os.Getenv(systemdCredDirEnv)
	if credsDir == "" {
		// Not using systemD creds
		return
	}

	systemdAPIFile := filepath.Clean(filepath.Join(credsDir, systemdCredAPIKeyFileEnv))
	if apiKey, ok := loadSecretFromFile(systemdAPIFile); ok {
		log.Info("Loaded Immich API key", "source", "systemd credential")
		c.ImmichAPIKey = apiKey
	}

	systemdPasswordFile := filepath.Clean(filepath.Join(credsDir, systemdCredPasswordFileEnv))
	if password, ok := loadSecretFromFile(systemdPasswordFile); ok {
		log.Info("Loaded password", "source", "systemd credential")
		c.Kiosk.Password = password
	}

	systemdWeatherAPIFile := filepath.Clean(filepath.Join(credsDir, systemdCredWeatherAPIKeyFileEnv))
	if weatherAPIKey, ok := loadSecretFromFile(systemdWeatherAPIFile); ok {
		log.Info("Loaded weather API key", "source", "systemd credential")
		for i, location := range c.WeatherLocations {
			if location.API == "" {
				log.Info("Added weather API key to", "location", location.Name)
				c.WeatherLocations[i].API = weatherAPIKey
			}
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
	case c.ImmichURL == "":
		log.Fatal("Immich URL is missing")
	case c.ImmichAPIKey == "":
		log.Fatal("Immich API key is missing")
	}
}

// checkDebuging enables the debug flag if verbose debugging is enabled.
// This ensures that verbose debugging also triggers regular debugging output.
func (c *Config) checkDebuging() {
	if c.Kiosk.DebugVerbose {
		c.Kiosk.Debug = true
	}
}

// cleanupSlice removes empty strings and placeholder values from a slice,
// and trims whitespace from remaining values.
func (c *Config) cleanupSlice(slice []string, placeholders ...string) []string {
	cleaned := make([]string, 0, len(slice))
	for _, item := range slice {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		isPlaceholder := false
		for _, placeholder := range placeholders {
			if strings.EqualFold(trimmed, placeholder) {
				isPlaceholder = true
				break
			}
		}
		if !isPlaceholder {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

// checkAssetBuckets validates and cleans up various asset filter lists in the Config.
// It processes Album, ExcludedAlbums, Person, and Date slices by:
// - Removing empty strings and placeholder values like "ALBUM_ID", "PERSON_ID", etc.
// - Trimming whitespace from all remaining values
// - Filtering out invalid date range formats
// The cleaned lists are then stored back in their respective Config fields.
func (c *Config) checkAssetBuckets() {

	c.Albums = c.cleanupSlice(c.Albums, "ALBUM_ID")
	c.ExcludedAlbums = c.cleanupSlice(c.ExcludedAlbums, "ALBUM_ID")

	c.People = c.cleanupSlice(c.People, "PERSON_ID")
	c.ExcludedPeople = c.cleanupSlice(c.ExcludedPeople, "PERSON_ID")

	c.Tags = c.cleanupSlice(c.Tags, "TAG_VALUE")
	c.ExcludedTags = c.cleanupSlice(c.ExcludedTags, "TAG_VALUE")

	c.Dates = c.cleanupSlice(c.Dates, "DATE_RANGE", "YYYY-MM-DD_to_YYYY-MM-DD")

	c.ExcludedPartners = c.cleanupSlice(c.ExcludedPartners, "PARTNER_ID")
}

// checkExcludedAlbums filters out any albums from c.Album that are present in
// c.ExcludedAlbums. It uses a map for O(1) lookups of excluded album IDs and
// filters in-place to avoid unnecessary allocations. If the resulting slice's
// capacity is significantly larger than its length, a new slice is allocated
// to prevent memory leaks.
func (c *Config) checkExcludedAlbums() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.ExcludedAlbums) == 0 || len(c.Albums) == 0 {
		return
	}

	excludeMap := make(map[string]struct{}, len(c.ExcludedAlbums))
	for _, id := range c.ExcludedAlbums {
		excludeMap[id] = struct{}{}
	}

	filtered := c.Albums[:0]
	for _, album := range c.Albums {
		if _, excluded := excludeMap[album]; !excluded {
			filtered = append(filtered, album)
		}
	}

	c.Albums = filtered

	if excess := cap(c.Albums) - len(c.Albums); excess > len(c.Albums) {
		c.Albums = append(make([]string, 0, len(c.Albums)), c.Albums...)
	}
}

// checkWeatherLocations validates the WeatherLocations in the Config.
// It checks each WeatherLocation for required fields (name, latitude, longitude, and API key),
// and logs an error message if any required fields are missing.
func (c *Config) checkWeatherLocations() {
	var validLocations []WeatherLocation

	for _, w := range c.WeatherLocations {
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
				log.Warn("Multiple default weather locations found.")
				w.Default = false
			} else {
				c.HasWeatherDefault = true
			}
		}
		if len(missingFields) == 0 {
			validLocations = append(validLocations, w)
		} else {
			log.Warn("Weather location is missing required fields. Ignoring this location.",
				"missing fields", strings.Join(missingFields, ", "), "name", w.Name)
		}
	}

	c.WeatherLocations = validLocations
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
			if nextRedirect, ok := strings.CutPrefix(targetURL.URL, "/"); ok {
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

// checkAlbumOrder validates the album order value and sets it to the default if invalid.
// The valid values are:
// - "random": Random order (default) - Display albums in random order
// - "asc"/"ascending"/"oldest": Ascending chronological order - Display albums from oldest to newest
// - "desc"/"descending"/"newest": Descending chronological order - Display albums from newest to oldest
// If an invalid value is provided, it will be set to "random" and a warning will be logged.
func (c *Config) checkAlbumOrder() {
	validOrders := []string{
		AlbumOrderRandom,
		AlbumOrderAsc,
		AlbumOrderAscending,
		AlbumOrderOldest,
		AlbumOrderDesc,
		AlbumOrderDescending,
		AlbumOrderNewest,
	}

	isValid := slices.Contains(validOrders, c.AlbumOrder)

	if !isValid {
		log.Warnf("Invalid album_order value: %s. Using default: random", c.AlbumOrder)
		c.AlbumOrder = AlbumOrderRandom
	}
}

func (c *Config) checkOffline() {
	if c.OfflineMode.Enabled {
		if c.OfflineMode.NumberOfAssets <= 0 {
			log.Warn("Invalid number_of_assets value. Using default: 100", "number_of_assets", c.OfflineMode.NumberOfAssets)
			c.OfflineMode.NumberOfAssets = 100
		}

		if c.OfflineMode.MaxSize == "" {
			log.Warn("Invalid max_size value. Using default: 1GB", "max_size", c.OfflineMode.MaxSize)
			c.OfflineMode.MaxSize = "1GB"
		}

		if c.OfflineMode.ParallelDownloads <= 0 {
			log.Warn("Invalid parallel_downloads value. Using default: 1", "parallel_downloads", c.OfflineMode.ParallelDownloads)
			c.OfflineMode.ParallelDownloads = 1
		}

		if c.OfflineMode.ExpirationHours < 0 {
			log.Warn("Invalid expiration_hours value. Using default: 72", "expiration_hours", c.OfflineMode.ExpirationHours)
			c.OfflineMode.ExpirationHours = 72
		}
	}
}

func checkSchema(config map[string]any, level string) bool {
	if strings.EqualFold(level, kiosk.ConfigValidationOff) {
		log.Info("Config validation disabled")
		return true
	}

	if !IsSchemaLoaded() {
		log.Warn("Schema not loaded, skipping validation")
		return true
	}

	// if we are using a config.yaml file but supplying immich_api_key || immich_url via ENVs get them
	if v, ok := config["immich_api_key"]; !ok || v == "" {
		config["immich_api_key"] = os.Getenv("KIOSK_IMMICH_API_KEY")
	}

	if v, ok := config["immich_url"]; !ok || v == "" {
		config["immich_url"] = os.Getenv("KIOSK_IMMICH_URL")
	}

	typed := ConfigTypes(config, Config{})
	for k, v := range config {
		if _, ok := typed[k]; !ok {
			typed[k] = v
		}
	}

	// Load JSON Schema from file
	schemaLoader := gojsonschema.NewStringLoader(SchemaJSON)
	docLoader := gojsonschema.NewGoLoader(typed)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		log.Error("Schema validation setup failed: validate", "err", err)
		return false
	}

	if !result.Valid() {
		switch strings.ToLower(level) {
		case "warning":
			log.Warn("Config validation failed:")
			for _, desc := range result.Errors() {
				log.Warnf("- %s", desc)
			}
		default:
			log.Error("Config validation failed:")
			for _, desc := range result.Errors() {
				log.Errorf("- %s", desc)
			}
		}
		return false
	}

	return true
}

// checkBurnIn validates burn-in prevention configuration values
func (c *Config) checkBurnIn() {
	if c.BurnInOpacity < 0 || c.BurnInOpacity > 100 {
		log.Warn("BurnInOpacity must be between 0 and 100, using default", "value", c.BurnInOpacity)
		c.BurnInOpacity = 70
	}
	if c.BurnInDuration < 1 {
		log.Warn("BurnInDuration must be at least 1 second, using default", "value", c.BurnInDuration)
		c.BurnInDuration = 30
	}
	if c.BurnInInterval < 0 {
		log.Warn("BurnInInterval cannot be negative, disabling", "value", c.BurnInInterval)
		c.BurnInInterval = 0
	}
}

func ConfigTypes(settings map[string]any, cfgStruct any) map[string]any {
	return convertConfigTypes(reflect.TypeOf(cfgStruct), settings)
}

func convertConfigTypes(typ reflect.Type, settings map[string]any) map[string]any {
	result := make(map[string]any)

	// If pointer, get the element
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	for i := range typ.NumField() {
		field := typ.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag == "" {
			tag = field.Name
		}

		raw, rawFound := settings[tag]
		if !rawFound {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			if nestedMap, ok := raw.(map[string]any); ok {
				result[tag] = convertConfigTypes(field.Type, nestedMap)
			}
		case reflect.Int:
			switch v := raw.(type) {
			case string:
				if n, err := strconv.Atoi(v); err == nil {
					result[tag] = n
				}
			default:
				result[tag] = v
			}
		case reflect.Bool:
			switch v := raw.(type) {
			case string:
				if b, err := strconv.ParseBool(v); err == nil {
					result[tag] = b
				}
			case float64:
				result[tag] = v != 0
			default:
				result[tag] = v
			}
		case reflect.Float32, reflect.Float64:
			switch v := raw.(type) {
			case string:
				if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
					if field.Type.Kind() == reflect.Float32 {
						result[tag] = float32(f)
					} else {
						result[tag] = f
					}
				}
			case float64:
				if field.Type.Kind() == reflect.Float32 {
					result[tag] = float32(v)
				} else {
					result[tag] = v
				}
			default:
				result[tag] = v
			}
		case reflect.String:
			if s, ok := raw.(string); ok {
				result[tag] = s
			}
		default:
			result[tag] = raw
		}
	}

	return result
}
