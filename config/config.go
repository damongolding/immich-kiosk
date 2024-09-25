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
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/labstack/echo/v4"
)

const (
	defaultImmichPort = "2283"
	defaultScheme     = "http://"
	DefaultDateLayout = "02/01/2006"
)

type KioskSettings struct {
	// Cache enable/disable api call and image caching
	Cache bool `mapstructure:"cache" default:"true"`

	// PreFetch fetch and cache an image in the background
	PreFetch bool `mapstructure:"pre_fetch" default:"true"`

	// Password the password used to add authentication to the frontend
	Password string `mapstructure:"password" default:""`

	// AssetWeighting use weighting when picking assets
	AssetWeighting bool `mapstructure:"asset_weighting" default:"true"`

	// debug modes
	Debug        bool `mapstructure:"debug" default:"false"`
	DebugVerbose bool `mapstructure:"debug_verbose" default:"false"`
}

type Config struct {
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
	// SplitView display two asstes side by side vertically
	SplitView bool `mapstructure:"split_view" query:"split_view" form:"split_view" default:"false"`

	// ShowArchived allow archived image to be displayed
	ShowArchived bool `mapstructure:"show_archived" query:"show_archived" form:"show_archived" default:"false"`
	// Person ID of person to display
	Person []string `mapstructure:"person" query:"person" form:"person" default:"[]"`
	// Album ID of album(s) to display
	Album []string `mapstructure:"album" query:"album" form:"album" default:"[]"`

	// ImageFit the fit style for main image
	ImageFit string `mapstructure:"image_fit" query:"image_fit" form:"image_fit" default:"contain"`
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

	// Kiosk settings that are unable to be changed via URL queries
	Kiosk KioskSettings `mapstructure:"kiosk"`

	// History past shown images
	History []string `form:"history" default:"[]"`
}

// New returns a new config pointer instance
func New() *Config {
	c := &Config{}
	defaults.SetDefaults(c)
	return c
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

// Load loads yaml config file into memory, then loads ENV vars. ENV vars overwrites yaml settings.
func (c *Config) Load() error {
	return c.load("config.yaml")
}

// Load loads yaml config file into memory with a custom path, then loads ENV vars. ENV vars overwrites yaml settings.
func (c *Config) LoadWithConfigLocation(configPath string) error {
	return c.load(configPath)
}

// load loads yaml config file into memory, then loads ENV vars. ENV vars overwrites yaml settings.
func (c *Config) load(configFile string) error {

	v := viper.NewWithOptions(viper.ExperimentalBindStruct())

	v.BindEnv("kiosk.password", "KIOSK_PASSWORD")
	v.BindEnv("kiosk.cache", "KIOSK_CACHE")
	v.BindEnv("kiosk.pre_fetch", "KIOSK_PRE_FETCH")
	v.BindEnv("kiosk.asset_weighting", "KIOSK_ASSET_WEIGHTING")

	v.BindEnv("kiosk.debug", "KIOSK_DEBUG")
	v.BindEnv("kiosk.debug_verbose", "KIOSK_DEBUG_VERBOSE")

	v.AddConfigPath(".")

	v.SetConfigFile(configFile)

	v.SetEnvPrefix("kiosk")

	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		log.Debug("config.yaml file not being used")
	}

	err = v.Unmarshal(&c)
	if err != nil {
		log.Error("Environment can't be loaded", "err", err)
		return err
	}

	c.checkRequiredFields()
	c.checkUrlScheme()
	c.checkDebuging()

	return nil

}

// ConfigWithOverridesOld overwrites base config with ones supplied via URL queries
//
// Deprecated: Keeping for legancy for now
func (c *Config) ConfigWithOverridesOld(queries url.Values) Config {

	configWithOverrides := c

	// check for person or album in quries and empty baseconfig slice if found
	if queries.Has("person") {
		configWithOverrides.Person = []string{}
	}

	if queries.Has("album") {
		configWithOverrides.Album = []string{}
	}

	v := reflect.ValueOf(configWithOverrides).Elem()

	// Loop through the queries and update struct fields
	for key, values := range queries {
		// Disble setting api and url for now
		if strings.ToLower(key) == "immich_api_key" || strings.ToLower(key) == "immich_url" {
			log.Error("tried to set Immich url or Immich api via queries")
			continue
		}

		if len(values) > 0 {
			// format to match field names
			key = strings.ReplaceAll(key, "_", " ")
			key = cases.Title(language.English, cases.Compact).String(key)
			key = strings.ReplaceAll(key, " ", "")

			// Get the field by name
			field := v.FieldByName(key)
			if field.IsValid() && field.CanSet() {

				// Loop values as queries are []string{}
				for _, value := range values {

					// We only want set values
					if value == "" {
						continue
					}

					// Set field (covert to correct type if needed)
					switch field.Kind() {
					case reflect.String: // all string values should be lowercase
						lowercaseValue := strings.ToLower(value)
						field.SetString(lowercaseValue)
					case reflect.Int:
						if n, err := strconv.Atoi(value); err == nil {
							field.SetInt(int64(n))
						}
					case reflect.Bool:
						if b, err := strconv.ParseBool(value); err == nil {
							field.SetBool(b)
						}

					// field type is a string e.g. Person is []string
					case reflect.Slice:
						elemType := field.Type().Elem()
						switch elemType.Kind() {
						case reflect.String:
							field.Set(reflect.Append(field, reflect.ValueOf(value)))
						case reflect.Int:
							if n, err := strconv.Atoi(value); err == nil {
								field.Set(reflect.Append(field, reflect.ValueOf(n)))
							}
						case reflect.Bool:
							if b, err := strconv.ParseBool(value); err == nil {
								field.Set(reflect.Append(field, reflect.ValueOf(b)))
							}
						}
					}
				}
			}
		}
	}

	return *configWithOverrides
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
