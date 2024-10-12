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
	"encoding/json"
	"errors"
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
	// v is the viper instance used for configuration management
	v *viper.Viper
	// mu is a mutex used to ensure thread-safe access to the configuration
	mu *sync.Mutex
	// ReloadTimeStamp timestamp for when the last client reload was called for
	ReloadTimeStamp string
	// configLastModTime stores the last modification time of the configuration file
	configLastModTime time.Time

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
		v:               viper.NewWithOptions(viper.ExperimentalBindStruct()),
		mu:              &sync.Mutex{},
		ReloadTimeStamp: time.Now().Format(time.RFC3339),
	}
	defaults.SetDefaults(c)
	info, err := os.Stat(defaultConfigFile)
	if err == nil {
		c.configLastModTime = info.ModTime()
	}
	return c
}

// hasConfigChanged checks if the configuration file has been modified since the last check.
func (c *Config) hasConfigChanged() bool {
	info, err := os.Stat(c.v.ConfigFileUsed())
	if err != nil {
		log.Errorf("Checking config file: %v", err)
		return false
	}

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

// Load loads yaml config file into memory, then loads ENV vars. ENV vars overwrites yaml settings.
func (c *Config) Load() error {
	return c.load(defaultConfigFile)
}

// Load loads yaml config file into memory with a custom path, then loads ENV vars. ENV vars overwrites yaml settings.
func (c *Config) LoadWithConfigLocation(configPath string) error {
	return c.load(configPath)
}

// WatchConfig starts a goroutine that periodically checks for changes in the configuration file
// and reloads the configuration if changes are detected.
//
// This function performs the following actions:
// 1. Retrieves the initial modification time of the config file.
// 2. Starts a goroutine that runs indefinitely.
// 3. Uses a ticker to check for config changes every 5 seconds.
// 4. If changes are detected, it reloads the configuration and updates the ReloadTimeStamp.
func (c *Config) WatchConfig() {

	configPath := c.v.ConfigFileUsed()

	fileInfo, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return
	}

	if fileInfo.IsDir() {
		log.Errorf("Config file %s is a directory", defaultConfigFile)
		return
	}

	info, err := os.Stat(configPath)
	if err != nil {
		log.Infof("Error getting initial file info: %v", err)
	} else {
		c.configLastModTime = info.ModTime()
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		//nolint:gosimple // Using for-select for ticker and potential future cases
		for {
			select {
			case <-ticker.C:
				if c.hasConfigChanged() {
					log.Info("Config file changed, reloading config")
					c.mu.Lock()
					err := c.Load()
					if err != nil {
						log.Errorf("Reloading config: %v", err)
					} else {
						c.ReloadTimeStamp = time.Now().Format(time.RFC3339)
						info, _ := os.Stat(defaultConfigFile)
						c.configLastModTime = info.ModTime()
					}
					c.mu.Unlock()
				}
			}
		}
	}()
}

// load loads yaml config file into memory, then loads ENV vars. ENV vars overwrites yaml settings.
func (c *Config) load(configFile string) error {

	if err := bindEnvironmentVariables(c.v); err != nil {
		log.Errorf("binding environment variables: %v", err)
	}

	viper.SetConfigName("config")  // Looks for 'config.yaml'
    
	// Optionally, specify the config file type (if you want to load from memory)
	viper.SetConfigType("yaml")

	// Add potential paths for the configuration file
	viper.AddConfigPath(".")        // Look in the current directory
	viper.AddConfigPath("./config") // Look in the 'config/' subdirectory

	c.v.SetEnvPrefix("kiosk")

	c.v.AutomaticEnv()

	err := c.v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Infof("Not using %s", configFile)
		} else if !isValidYAML(configFile) {
			log.Fatal(err)
		}
	}

	err = c.v.Unmarshal(&c)
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

	queries := e.Request().URL.Query()

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
