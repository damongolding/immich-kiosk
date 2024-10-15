// Package config provides configuration management for the Immich Kiosk application.
//
// It includes structures and methods for loading, parsing, and managing
// configuration settings from various sources including YAML files,
// environment variables, and URL query parameters.
//
// The package offers functionality to:
// - Define default configuration values
// - Load configuration from files and environment variables
// - Override configuration with URL query parameters
// - Validate and process configuration settings
//
// Key types:
// - Config: The main configuration structure
// - KioskSettings: Settings specific to kiosk mode
//
// Key functions:
// - New: Creates a new Config instance with default values
// - Load: Loads configuration from a file and environment variables
// - ConfigWithOverrides: Applies overrides from URL queries to the configuration
package config

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/labstack/echo/v4"
)

const (
	defaultImmichPort = "2283"
	defaultScheme     = "http://"
	DefaultDateLayout = "02/01/2006"
	defaultConfigFile = "config.yaml"
)

type KioskSettings struct {
	// Port which port to use
	Port int `mapstructure:"port" default:"3000"`

	// WatchConfig if kiosk should watch config file for changes
	WatchConfig bool `mapstructure:"watch_config" default:"false"`

	// Cache enable/disable api call and image caching
	Cache bool `mapstructure:"cache" default:"true"`

	// PreFetch fetch and cache an image in the background
	PreFetch bool `mapstructure:"prefetch" default:"true"`

	// Password the password used to add authentication to the frontend
	Password string `mapstructure:"password" default:""`

	// AssetWeighting use weighting when picking assets
	AssetWeighting bool `mapstructure:"asset_weighting" default:"true"`

	// debug modes
	Debug        bool `mapstructure:"debug" default:"false"`
	DebugVerbose bool `mapstructure:"debug_verbose" default:"false"`
}

type Config struct {
	// V is the viper instance used for configuration management
	V *viper.Viper
	// mu is a mutex used to ensure thread-safe access to the configuration
	mu *sync.Mutex
	// ReloadTimeStamp timestamp for when the last client reload was called for
	ReloadTimeStamp string
	// configLastModTime stores the last modification time of the configuration file
	configLastModTime time.Time
	// configHash stores the SHA-256 hash of the configuration file
	configHash string

	// ImmichApiKey Immich key to access assets
	ImmichApiKey string `mapstructure:"immich_api_key" default:""`
	// ImmichUrl Immuch base url
	ImmichUrl string `mapstructure:"immich_url" default:""`

	// DisableUi a shortcut to disable ShowTime, ShowDate, ShowImageTime and ShowImageDate
	DisableUi bool `mapstructure:"disable_ui" query:"disable_ui" form:"disable_ui" default:"false"`

	// ShowTime whether to display clock
	ShowTime bool `mapstructure:"show_time" query:"show_time" form:"show_time" default:"false"`
	// TimeFormat whether to use 12 of 24 hour format for clock
	TimeFormat string `mapstructure:"time_format" query:"time_format" form:"time_format" default:""`
	// ShowDate whether to display date
	ShowDate bool `mapstructure:"show_date" query:"show_date" form:"show_date" default:"false"`
	//  DateFormat format for date
	DateFormat string `mapstructure:"date_format" query:"date_format" form:"date_format" default:""`

	// Refresh time between fetching new image
	Refresh int `mapstructure:"refresh" query:"refresh" form:"refresh" default:"60"`
	// DisableScreensaver asks browser to disable screensaver
	DisableScreensaver bool `mapstructure:"disable_screensaver" query:"disable_screensaver" form:"disable_screensaver" default:"false"`
	// HideCursor hide cursor via CSS
	HideCursor bool `mapstructure:"hide_cursor" query:"hide_cursor" form:"hide_cursor" default:"false"`
	// FontSize the base font size as a percentage
	FontSize int `mapstructure:"font_size" query:"font_size" form:"font_size" default:"100"`
	// Theme which theme to use
	Theme string `mapstructure:"theme" query:"theme" form:"theme" default:"fade"`
	// Layout which layout to use
	Layout string `mapstructure:"layout" query:"layout" form:"layout" default:"single"`

	// SleepStart when to start sleep mode
	SleepStart string `mapstructure:"sleep_start" query:"sleep_start" form:"sleep_start" default:""`
	// SleepEnd when to exit sleep mode
	SleepEnd string `mapstructure:"sleep_end" query:"sleep_end" form:"sleep_end" default:""`

	// ShowArchived allow archived image to be displayed
	ShowArchived bool `mapstructure:"show_archived" query:"show_archived" form:"show_archived" default:"false"`
	// Person ID of person to display
	Person []string `mapstructure:"person" query:"person" form:"person" default:"[]"`
	// Album ID of album(s) to display
	Album []string `mapstructure:"album" query:"album" form:"album" default:"[]"`

	// ImageFit the fit style for main image
	ImageFit string `mapstructure:"image_fit" query:"image_fit" form:"image_fit" default:"contain"`
	// ImageZoom add a zoom effect to images
	ImageZoom bool `mapstructure:"image_zoom" query:"image_zoom" form:"image_zoom" default:"false"`
	// ImageZoomAmount the amount to zoom in/out of images
	ImageZoomAmount int `mapstructure:"image_zoom_amount" query:"image_zoom_amount" form:"image_zoom_amount" default:"120"`
	// BackgroundBlur whether to display blurred image as background
	BackgroundBlur bool `mapstructure:"background_blur" query:"background_blur" form:"background_blur" default:"true"`
	// BackgroundBlur which transition to use none|fade|cross-fade
	Transition string `mapstructure:"transition" query:"transition" form:"transition" default:""`
	// FadeTransitionDuration sets the length of the fade transition
	FadeTransitionDuration float32 `mapstructure:"fade_transition_duration" query:"fade_transition_duration" form:"fade_transition_duration" default:"1"`
	// CrossFadeTransitionDuration sets the length of the cross-fade transition
	CrossFadeTransitionDuration float32 `mapstructure:"cross_fade_transition_duration" query:"cross_fade_transition_duration" form:"cross_fade_transition_duration" default:"1"`

	// ShowProgress display a progress bar
	ShowProgress bool `mapstructure:"show_progress" query:"show_progress" form:"show_progress" default:"false"`
	// CustomCSS use custom css file
	CustomCSS bool `mapstructure:"custom_css" query:"custom_css" form:"custom_css" default:"true"`

	// ShowImageTime whether to display image time
	ShowImageTime bool `mapstructure:"show_image_time" query:"show_image_time" form:"show_image_time" default:"false"`
	// ImageTimeFormat  whether to use 12 of 24 hour format
	ImageTimeFormat string `mapstructure:"image_time_format" query:"image_time_format" form:"image_time_format" default:""`
	// ShowImageDate whether to display image date
	ShowImageDate bool `mapstructure:"show_image_date" query:"show_image_date" form:"show_image_date"  default:"false"`
	// ImageDateFormat format for image date
	ImageDateFormat string `mapstructure:"image_date_format" query:"image_date_format" form:"image_date_format" default:""`
	// ShowImageExif display image exif data (f number, iso, shutter speed, Focal length)
	ShowImageExif bool `mapstructure:"show_image_exif" query:"show_image_exif" form:"show_image_exif" default:"false"`
	// ShowImageLocation display image location data
	ShowImageLocation bool `mapstructure:"show_image_location" query:"show_image_location" form:"show_image_location" default:"false"`
	// ShowImageID display image ID
	ShowImageID bool `mapstructure:"show_image_id" query:"show_image_id" form:"show_image_id" default:"false"`

	// Kiosk settings that are unable to be changed via URL queries
	Kiosk KioskSettings `mapstructure:"kiosk"`

	// History past shown images
	History []string `form:"history" default:"[]"`
}

// New returns a new config pointer instance
func New() *Config {
	c := &Config{
		V:               viper.NewWithOptions(viper.ExperimentalBindStruct()),
		mu:              &sync.Mutex{},
		ReloadTimeStamp: time.Now().Format(time.RFC3339),
	}
	defaults.SetDefaults(c)
	return c
}

// hasConfigChanged checks if the configuration file has been modified since the last check.
func (c *Config) hasConfigChanged() bool {
	info, err := os.Stat(defaultConfigFile)
	if err != nil {
		log.Errorf("Checking config file: %v", err)
		return false
	}

	log.Debug("Config file mod stats", "stored", c.configLastModTime, "live", info.ModTime(), "changed?", info.ModTime().After(c.configLastModTime))

	return info.ModTime().After(c.configLastModTime)
}

// bindEnvironmentVariables binds specific environment variables to their corresponding
// configuration keys in the Viper instance. This function allows for easy mapping
// between environment variables and configuration settings.
//
// It iterates through a predefined list of mappings between config keys and
// environment variable names, binding each pair using Viper's BindEnv method.
//
// If any errors occur during the binding process, they are collected and
// returned as a single combined error.
//
// Parameters:
//   - v: A pointer to a viper.Viper instance to which the environment variables will be bound.
//
// Returns:
//   - An error if any binding operations fail, or nil if all bindings are successful.
func bindEnvironmentVariables(v *viper.Viper) error {
	var errs []error

	bindVars := []struct {
		configKey string
		envVar    string
	}{
		{"kiosk.port", "KIOSK_PORT"},
		{"kiosk.watch_config", "KIOSK_WATCH_CONFIG"},
		{"kiosk.password", "KIOSK_PASSWORD"},
		{"kiosk.cache", "KIOSK_CACHE"},
		{"kiosk.prefetch", "KIOSK_PREFETCH"},
		{"kiosk.asset_weighting", "KIOSK_ASSET_WEIGHTING"},
		{"kiosk.debug", "KIOSK_DEBUG"},
		{"kiosk.debug_verbose", "KIOSK_DEBUG_VERBOSE"},
	}

	for _, bv := range bindVars {
		if err := v.BindEnv(bv.configKey, bv.envVar); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// isValidYAML checks if the given file is a valid YAML file.
func isValidYAML(filename string) bool {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Errorf("Error reading file: %v", err)
		return false
	}

	var data interface{}
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
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

// hasConfigMtimeChanged checks if the configuration file has been modified since the last check.
func (c *Config) hasConfigMtimeChanged() bool {
	info, err := os.Stat(c.V.ConfigFileUsed())
	if err != nil {
		log.Errorf("Checking config file: %v", err)
		return false
	}

	return info.ModTime().After(c.configLastModTime)
}

// Function to calculate the SHA-256 hash of a file
func (c *Config) configFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// checkUrlScheme checks given url has correct scheme and adds http:// if non if found
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

// checkRequiredFields check is required config files are set.
func (c *Config) checkRequiredFields() {
	switch {
	case c.ImmichUrl == "":
		log.Fatal("Immich Url is missing")
	case c.ImmichApiKey == "":
		log.Fatal("Immich API is missing")
	}
}

func (c *Config) checkDebuging() {
	if c.Kiosk.DebugVerbose {
		c.Kiosk.Debug = true
	}
}

func (c *Config) checkAlbumAndPerson() {
	newAlbum := []string{}
	for _, album := range c.Album {
		if album != "" && album != "ALBUM_ID" {
			newAlbum = append(newAlbum, strings.TrimSpace(album))
		}
	}
	c.Album = newAlbum

	newPerson := []string{}
	for _, person := range c.Person {
		if person != "" && person != "PERSON_ID" {
			newPerson = append(newPerson, strings.TrimSpace(person))
		}
	}
	c.Person = newPerson
}

// WatchConfig sets up a configuration file watcher that monitors for changes
// and reloads the configuration when necessary.
func (c *Config) WatchConfig() {
	configPath := c.V.ConfigFileUsed()

	if err := validateConfigFile(configPath); err != nil {
		log.Error(err)
		return
	}

	if err := c.initializeConfigState(); err != nil {
		log.Error("Failed to initialize config state:", err)
		return
	}

	go c.watchConfigChanges()
}

// initializeConfigState sets up the initial state of the configuration,
// including the last modification time and hash of the config file.
func (c *Config) initializeConfigState() error {
	info, err := os.Stat(c.V.ConfigFileUsed())
	if err != nil {
		return fmt.Errorf("getting initial file mTime: %v", err)
	}
	c.configLastModTime = info.ModTime()

	configHash, err := c.configFileHash(c.V.ConfigFileUsed())
	if err != nil {
		return fmt.Errorf("getting initial file hash: %v", err)
	}
	c.configHash = configHash

	return nil
}

// watchConfigChanges continuously monitors the configuration file for changes
// and triggers a reload when necessary.
func (c *Config) watchConfigChanges() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	hashCheckCount := 0
	const hashCheckInterval = 12

	for range ticker.C {
		if c.hasConfigMtimeChanged() {
			c.reloadConfig("mTime changed")
			hashCheckCount = 0
			continue
		}

		if hashCheckCount >= hashCheckInterval {
			if c.hasConfigHashChanged() {
				c.reloadConfig("hash changed")
			}
			hashCheckCount = 0
		}

		hashCheckCount++
	}
}

// hasConfigHashChanged checks if the hash of the config file has changed.
func (c *Config) hasConfigHashChanged() bool {
	configHash, err := c.configFileHash(c.V.ConfigFileUsed())
	if err != nil {
		log.Error("configFileHash", "err", err)
		return false
	}
	return c.configHash != configHash
}

// reloadConfig reloads the configuration when a change is detected.
func (c *Config) reloadConfig(reason string) {
	log.Infof("Config file %s, reloading config", reason)
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.Load(); err != nil {
		log.Error("Failed to reload config:", err)
	}

	c.updateConfigState()
}

// updateConfigState updates the configuration state after a reload.
func (c *Config) updateConfigState() {
	configHash, _ := c.configFileHash(c.V.ConfigFileUsed())
	c.configHash = configHash
	c.ReloadTimeStamp = time.Now().Format(time.RFC3339)
	info, _ := os.Stat(c.V.ConfigFileUsed())
	c.configLastModTime = info.ModTime()
}

// load loads yaml config file into memory, then loads ENV vars. ENV vars overwrites yaml settings.
func (c *Config) Load() error {

	if err := bindEnvironmentVariables(c.V); err != nil {
		log.Errorf("binding environment variables: %v", err)
	}

	c.V.SetConfigName("config")
	c.V.SetConfigType("yaml")

	// Add potential paths for the configuration file
	c.V.AddConfigPath(".")         // Look in the current directory
	c.V.AddConfigPath("./config/") // Look in the 'config/' subdirectory
	c.V.AddConfigPath("../")       // Look in the parent directory for testing

	c.V.SetEnvPrefix("kiosk")

	c.V.AutomaticEnv()

	err := c.V.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("Not using config.yaml")
		} else if !isValidYAML(c.V.ConfigFileUsed()) {
			log.Fatal(err)
		}
	}

	err = c.V.Unmarshal(&c)
	if err != nil {
		log.Error("Environment can't be loaded", "err", err)
		return err
	}

	c.checkRequiredFields()
	c.checkAlbumAndPerson()
	c.checkUrlScheme()
	c.checkDebuging()

	return nil
}

// ConfigWithOverrides overwrites base config with ones supplied via URL queries
func (c *Config) ConfigWithOverrides(e echo.Context) error {

	queries := e.QueryParams()

	// check for person or album in quries and empty baseconfig slice if found
	if queries.Has("person") {
		c.Person = []string{}
	}

	if queries.Has("album") {
		c.Album = []string{}
	}

	err := e.Bind(c)
	if err != nil {
		return err
	}

	return nil

}

// String returns a string representation of the Config structure.
// If debug_verbose is not enabled, it returns a message prompting to enable it.
// Otherwise, it returns a JSON-formatted string of the entire Config structure.
//
// This method is useful for debugging and logging purposes, providing a
// detailed view of the current configuration when verbose debugging is enabled.
//
// Returns:
//   - A string containing either a prompt to enable debug_verbose or
//     the JSON representation of the Config structure.
func (c *Config) String() string {
	if !c.Kiosk.DebugVerbose {
		return "use debug_verbose for more info"
	}

	out, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		log.Error("", "err", err)
	}
	return string(out)
}
