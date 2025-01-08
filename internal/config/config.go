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
	"net/url"
	"os"
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

// Redirect represents a URL redirection configuration with a friendly name.
type Redirect struct {
	// Name is the friendly identifier used to access the redirect
	Name string `mapstructure:"name"`
	// URL is the destination address for the redirect
	URL string `mapstructure:"url"`
	// Type specifies the redirect behaviour (e.g., "internal", "external")
	Type string `mapstructure:"type"`
}

type KioskSettings struct {
	// Redirects defines a list of URL redirections with friendly names
	Redirects []Redirect `mapstructure:"redirects" default:"[]"`
	//RedirectsMap provides O(1) lookup of redirect URLs by their friendly name
	RedirectsMap map[string]Redirect `json:"-"`

	// Port which port to use
	Port int `json:"port" mapstructure:"port" default:"3000"`

	// WatchConfig if kiosk should watch config file for changes
	WatchConfig bool `json:"watchConfig" mapstructure:"watch_config" default:"false"`

	// FetchedAssetsSize the size of assets requests from Immich. min=1 max=1000
	FetchedAssetsSize int `json:"fetchedAssetsSize" mapstructure:"fetched_assets_size" default:"1000"`

	// HTTPTimeout time in seconds before an http request will timeout
	HTTPTimeout int `json:"httpTimeout" mapstructure:"http_timeout" default:"20"`

	// Cache enable/disable api call and image caching
	Cache bool `json:"cache" mapstructure:"cache" default:"true"`

	// PreFetch fetch and cache an image in the background
	PreFetch bool `json:"preFetch" mapstructure:"prefetch" default:"true"`

	// Password the password used to add authentication to the frontend
	Password string `json:"-" mapstructure:"password" default:""`

	// AssetWeighting use weighting when picking assets
	AssetWeighting bool `json:"assetWeighting" mapstructure:"asset_weighting" default:"true"`

	// debug modes
	Debug        bool `json:"debug" mapstructure:"debug" default:"false"`
	DebugVerbose bool `json:"debugVerbose" mapstructure:"debug_verbose" default:"false"`
}

type WeatherLocation struct {
	Name    string `mapstructure:"name"`
	Lat     string `mapstructure:"lat"`
	Lon     string `mapstructure:"lon"`
	API     string `mapstructure:"api"`
	Unit    string `mapstructure:"unit"`
	Lang    string `mapstructure:"lang"`
	Default bool   `mapstructure:"default"`
}

type Webhook struct {
	Url    string `json:"url" mapstructure:"url"`
	Event  string `json:"event" mapstructure:"event"`
	Secret string `json:"secret" mapstructure:"secret"`
}

// ClientData represents the client-specific dimensions received from the frontend.
type ClientData struct {
	// Width represents the client's viewport width in pixels
	Width int `json:"client_width" query:"client_width" form:"client_width"`
	// Height represents the client's viewport height in pixels
	Height int `json:"client_height" query:"client_height" form:"client_height"`
}

// Config represents the main configuration structure for the Immich Kiosk application.
// It contains all the settings that control the behavior and appearance of the kiosk,
// including connection details, display options, image settings, and various feature toggles.
//
// The structure supports configuration through YAML files, environment variables,
// and URL query parameters. Many fields can be dynamically updated through URL queries
// during runtime.
//
// # Tags used in the configuration structure:
//   - mapstructure: field name from yaml file
//   - query: enables URL query parameter binding
//   - form: enables form parameter binding
//   - default: sets default value
//   - lowercase: converts string value to lowercase
type Config struct {
	// V is the viper instance used for configuration management
	V *viper.Viper `json:"-"`
	// mu is a mutex used to ensure thread-safe access to the configuration
	mu *sync.RWMutex `json:"-"`
	// ReloadTimeStamp timestamp for when the last client reload was called for
	ReloadTimeStamp string `json:"-"`
	// configLastModTime stores the last modification time of the configuration file
	configLastModTime time.Time `json:"-"`
	// configHash stores the SHA-256 hash of the configuration file
	configHash string `json:"-"`

	// ImmichApiKey Immich key to access assets
	ImmichApiKey string `json:"-" mapstructure:"immich_api_key" default:""`
	// ImmichUrl Immuch base url
	ImmichUrl string `json:"-" mapstructure:"immich_url" default:""`

	// ImmichExternalUrl specifies an external URL for Immich access. This can be used when
	// the Immich instance is accessed through a different URL externally vs internally
	// (e.g., when using reverse proxies or different network paths)
	ImmichExternalUrl string `json:"-" mapstructure:"immich_external_url" default:""`

	// DisableUi a shortcut to disable ShowTime, ShowDate, ShowImageTime and ShowImageDate
	DisableUi bool `json:"disableUi" mapstructure:"disable_ui" query:"disable_ui" form:"disable_ui" default:"false"`
	// Frameless remove border on frames
	Frameless bool `json:"frameless" mapstructure:"frameless" query:"frameless" form:"frameless" default:"false"`

	// ShowTime whether to display clock
	ShowTime bool `json:"showTime" mapstructure:"show_time" query:"show_time" form:"show_time" default:"false"`
	// TimeFormat whether to use 12 of 24 hour format for clock
	TimeFormat string `json:"timeFormat" mapstructure:"time_format" query:"time_format" form:"time_format" default:""`
	// ShowDate whether to display date
	ShowDate bool `json:"showDate" mapstructure:"show_date" query:"show_date" form:"show_date" default:"false"`
	//  DateFormat format for date
	DateFormat string `json:"dateFormat" mapstructure:"date_format" query:"date_format" form:"date_format" default:""`

	// Refresh time between fetching new image
	Refresh int `json:"refresh" mapstructure:"refresh" query:"refresh" form:"refresh" default:"60"`
	// DisableScreensaver asks browser to disable screensaver
	DisableScreensaver bool `json:"disableScreensaver" mapstructure:"disable_screensaver" query:"disable_screensaver" form:"disable_screensaver" default:"false"`
	// HideCursor hide cursor via CSS
	HideCursor bool `json:"hideCursor" mapstructure:"hide_cursor" query:"hide_cursor" form:"hide_cursor" default:"false"`
	// FontSize the base font size as a percentage
	FontSize int `json:"fontSize" mapstructure:"font_size" query:"font_size" form:"font_size" default:"100"`
	// Theme which theme to use
	Theme string `json:"theme" mapstructure:"theme" query:"theme" form:"theme" default:"fade" lowercase:"true"`
	// Layout which layout to use
	Layout string `json:"layout" mapstructure:"layout" query:"layout" form:"layout" default:"single" lowercase:"true"`

	// SleepStart when to start sleep mode
	SleepStart string `json:"sleepStart" mapstructure:"sleep_start" query:"sleep_start" form:"sleep_start" default:""`
	// SleepEnd when to exit sleep mode
	SleepEnd string `json:"sleepEnd" mapstructure:"sleep_end" query:"sleep_end" form:"sleep_end" default:""`
	// SleepIcon display sleep icon
	SleepIcon bool `json:"sleepIcon" mapstructure:"sleep_icon" query:"sleep_icon" form:"sleep_icon" default:"true"`

	// ShowArchived allow archived image to be displayed
	ShowArchived bool `json:"showArchived" mapstructure:"show_archived" query:"show_archived" form:"show_archived" default:"false"`
	// Person ID of person to display
	Person []string `json:"person" mapstructure:"person" query:"person" form:"person" default:"[]"`
	// Album ID of album(s) to display
	Album          []string `json:"album" mapstructure:"album" query:"album" form:"album" default:"[]"`
	ExcludedAlbums []string `json:"excluded_albums" mapstructure:"excluded_albums" query:"exclude_album" form:"exclude_album" default:"[]"`

	// ImageFit the fit style for main image
	ImageFit string `json:"imageFit" mapstructure:"image_fit" query:"image_fit" form:"image_fit" default:"contain" lowercase:"true"`
	// ImageEffect which effect to apply to image (if any)
	ImageEffect string `json:"imageEffect" mapstructure:"image_effect" query:"image_effect" form:"image_effect" default:"" lowercase:"true"`
	// ImageEffectAmount the amount of effect to apply
	ImageEffectAmount int `json:"imageEffectAmount" mapstructure:"image_effect_amount" query:"image_effect_amount" form:"image_effect_amount" default:"120"`
	// UseOriginalImage use the original image
	UseOriginalImage bool `json:"useOriginalImage" mapstructure:"use_original_image" query:"use_original_image" form:"use_original_image" default:"false"`
	// BackgroundBlur whether to display blurred image as background
	BackgroundBlur bool `json:"backgroundBlur" mapstructure:"background_blur" query:"background_blur" form:"background_blur" default:"true"`
	// BackgroundBlur which transition to use none|fade|cross-fade
	Transition string `json:"transition" mapstructure:"transition" query:"transition" form:"transition" default:"" lowercase:"true"`
	// FadeTransitionDuration sets the length of the fade transition
	FadeTransitionDuration float32 `json:"fadeTransitionDuration" mapstructure:"fade_transition_duration" query:"fade_transition_duration" form:"fade_transition_duration" default:"1"`
	// CrossFadeTransitionDuration sets the length of the cross-fade transition
	CrossFadeTransitionDuration float32 `json:"crossFadeTransitionDuration" mapstructure:"cross_fade_transition_duration" query:"cross_fade_transition_duration" form:"cross_fade_transition_duration" default:"1"`

	// ShowProgress display a progress bar
	ShowProgress bool `json:"showProgress" mapstructure:"show_progress" query:"show_progress" form:"show_progress" default:"false"`
	// CustomCSS use custom css file
	CustomCSS bool `json:"customCSS" mapstructure:"custom_css" query:"custom_css" form:"custom_css" default:"true"`

	// ShowAlbumName whether to display the album name
	ShowAlbumName bool `json:"showAlbumName" mapstructure:"show_album_name" query:"show_album_name" form:"show_album_name" default:"false"`
	// ShowPersonName whether to display the person name
	ShowPersonName bool `json:"showPersonName" mapstructure:"show_person_name" query:"show_person_name" form:"show_person_name" default:"false"`

	// ShowImageTime whether to display image time
	ShowImageTime bool `json:"showImageTime" mapstructure:"show_image_time" query:"show_image_time" form:"show_image_time" default:"false"`
	// ImageTimeFormat  whether to use 12 of 24 hour format
	ImageTimeFormat string `json:"imageTimeFormat" mapstructure:"image_time_format" query:"image_time_format" form:"image_time_format" default:""`
	// ShowImageDate whether to display image date
	ShowImageDate bool `json:"showImageDate" mapstructure:"show_image_date" query:"show_image_date" form:"show_image_date"  default:"false"`
	// ImageDateFormat format for image date
	ImageDateFormat string `json:"imageDateFormat" mapstructure:"image_date_format" query:"image_date_format" form:"image_date_format" default:""`
	// ShowImageDescription isplay image description
	ShowImageDescription bool `json:"showImageDescription" mapstructure:"show_image_description" query:"show_image_description" form:"show_image_description" default:"false"`
	// ShowImageExif display image exif data (f number, iso, shutter speed, Focal length)
	ShowImageExif bool `json:"showImageExif" mapstructure:"show_image_exif" query:"show_image_exif" form:"show_image_exif" default:"false"`
	// ShowImageLocation display image location data
	ShowImageLocation bool `json:"showImageLocation" mapstructure:"show_image_location" query:"show_image_location" form:"show_image_location" default:"false"`
	// HideCountries hide country names in location information
	HideCountries []string `json:"hideCountries" mapstructure:"hide_countries" query:"hide_countries" form:"hide_countries" default:"[]"`
	// ShowImageID display image ID
	ShowImageID bool `json:"showImageID" mapstructure:"show_image_id" query:"show_image_id" form:"show_image_id" default:"false"`

	// ShowMoreInfo enables the display of additional information about the current image
	ShowMoreInfo bool `json:"showMoreInfo" mapstructure:"show_more_info" query:"show_more_info" form:"show_more_info" default:"true"`
	// ShowMoreInfoImageLink shows a link to the original image in the additional information panel
	ShowMoreInfoImageLink bool `json:"showMoreInfoImageLink" mapstructure:"show_more_info_image_link" query:"show_more_info_image_link" form:"show_more_info_image_link" default:"true"`
	// ShowMoreInfoQrCode displays a QR code linking to the original image in the additional information panel
	ShowMoreInfoQrCode bool `json:"showMoreInfoQrCode" mapstructure:"show_more_info_qr_code" query:"show_more_info_qr_code" form:"show_more_info_qr_code" default:"true"`

	// WeatherLocations A list of locations to fetch and display weather data from. Each location
	WeatherLocations []WeatherLocation `json:"weather" mapstructure:"weather" default:"[]"`
	// HasWeatherDefault indicates whether any weather location has been set as the default.
	HasWeatherDefault bool `json:"-" default:"false"`

	// OptimizeImages tells Kiosk to optimize imahes
	OptimizeImages bool `json:"optimize_images" mapstructure:"optimize_images" query:"optimize_images" form:"optimize_images" default:"false"`
	// UseGpu tells Kiosk to use GPU where possible
	UseGpu bool `json:"useGpu" mapstructure:"use_gpu" query:"use_gpu" form:"use_gpu" default:"true"`

	// Webhooks defines a list of webhook endpoints and their associated events that should trigger notifications.
	Webhooks []Webhook `json:"webhooks" mapstructure:"webhooks" default:"[]"`

	// Kiosk settings that are unable to be changed via URL queries
	Kiosk KioskSettings `json:"kiosk" mapstructure:"kiosk"`

	// ClientData data sent from the client with data regarding itself
	ClientData ClientData
	// History past shown images
	History []string `json:"history" form:"history" default:"[]"`
}

// New returns a new config pointer instance
func New() *Config {
	c := &Config{
		V:               viper.NewWithOptions(viper.ExperimentalBindStruct()),
		mu:              &sync.RWMutex{},
		ReloadTimeStamp: time.Now().Format(time.RFC3339),
	}
	defaults.SetDefaults(c)
	return c
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
		{"kiosk.fetched_assets_size", "KIOSK_FETCHED_ASSETS_SIZE"},
		{"kiosk.http_timeout", "KIOSK_HTTP_TIMEOUT"},
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
	c.V.AddConfigPath("../../")    // Look in the parent directory for testing

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
	c.checkLowercaseTaggedFields()
	c.checkAlbumAndPerson()
	c.checkExcludedAlbums()
	c.checkUrlScheme()
	c.checkHideCountries()
	c.checkWeatherLocations()
	c.checkDebuging()
	c.checkFetchedAssetsSize()
	c.checkRedirects()

	return nil
}

// ConfigWithOverrides overwrites base config with ones supplied via URL queries
func (c *Config) ConfigWithOverrides(queries url.Values, e echo.Context) error {

	// check for person or album in quries and empty baseconfig slice if found
	if queries.Has("person") || queries.Has("album") {
		c.Person = []string{}
		c.Album = []string{}
	}

	err := e.Bind(c)
	if err != nil {
		return err
	}

	c.checkExcludedAlbums()

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
