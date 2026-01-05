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
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/goodsign/monday"
	"github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"

	"github.com/labstack/echo/v4"
)

const (
	defaultScheme     = "http://"
	DefaultDateLayout = "02/01/2006"

	AlbumOrderRandom     = "random"
	AlbumOrderAscending  = "ascending"
	AlbumOrderAsc        = "asc"
	AlbumOrderOldest     = "oldest"
	AlbumOrderDescending = "descending"
	AlbumOrderDesc       = "desc"
	AlbumOrderNewest     = "newest"

	redactedMarker = "REDACTED"

	// Secrets
	systemdCredDirEnv               = "CREDENTIALS_DIRECTORY"
	systemdCredAPIKeyFileEnv        = "kiosk_immich_api_key"
	systemdCredPasswordFileEnv      = "kiosk_password_file"
	systemdCredWeatherAPIKeyFileEnv = "kiosk_weather_api_key_file"

	apiKeyFileEnv        = "KIOSK_IMMICH_API_KEY_FILE"
	passwordFileEnv      = "KIOSK_PASSWORD_FILE"
	weatherAPIKeyFileEnv = "KIOSK_WEATHER_API_KEY_FILE"
)

type OfflineMode struct {
	// MaxSize specifies the maximum storage size for offline assets in a human-readable format e.g. "1GB", "2TB", "500MB"
	MaxSize string `yaml:"max_size" mapstructure:"max_size" default:"0"`
	// NumberOfAssets specifies the maximum number of assets to store in offline mode
	NumberOfAssets int `yaml:"number_of_assets" mapstructure:"number_of_assets" default:"100"`
	// ParallelDownloads specifies the maximum number of concurrent downloads in offline mode
	ParallelDownloads int `yaml:"parallel_downloads" mapstructure:"parallel_downloads" default:"1"`
	// ExpirationHours specifies how long offline assets should be kept before being considered expired (in hours)
	ExpirationHours int `yaml:"expiration_hours" mapstructure:"expiration_hours" default:"0"`
	// Enabled indicates whether offline mode is enabled
	Enabled bool `yaml:"enabled" mapstructure:"enabled" default:"false"`
}

// Redirect represents a URL redirection configuration with a friendly name.
type Redirect struct {
	// Name is the friendly identifier used to access the redirect
	Name string `yaml:"name" mapstructure:"name" redact:"true"`
	// URL is the destination address for the redirect
	URL string `yaml:"url" mapstructure:"url" redact:"true"`
	// Type specifies the redirect behaviour (e.g., "internal", "external")
	Type string `yaml:"type" mapstructure:"type"`
}

type KioskSettings struct {
	// RedirectsMap provides O(1) lookup of redirect URLs by their friendly name
	RedirectsMap map[string]Redirect `json:"-" yaml:"-"`

	// Version
	Version string `json:"version" yaml:"version"`

	ConfigValidationLevel string `json:"configValidationLevel" yaml:"config_validation_level" mapstructure:"config_validation_level" default:"error" lowercase:"true"`

	// Password the password used to add authentication to the frontend
	Password string `json:"-" yaml:"password" mapstructure:"password" default:"" redact:"true"`

	// Redirects defines a list of URL redirections with friendly names
	Redirects []Redirect `yaml:"redirects" mapstructure:"redirects" default:"[]"`

	// Port which port to use
	Port int `json:"port" yaml:"port" mapstructure:"port" default:"3000"`

	// FetchedAssetsSize the size of assets requests from Immich. min=1 max=1000
	FetchedAssetsSize int `json:"fetchedAssetsSize" yaml:"fetched_assets_size" mapstructure:"fetched_assets_size" default:"1000"`

	// HTTPTimeout time in seconds before an http request will timeout
	HTTPTimeout int `json:"httpTimeout" yaml:"http_timeout" mapstructure:"http_timeout" default:"20"`

	// BehindProxy specifies whether the kiosk is behind a proxy
	BehindProxy bool `json:"behindProxy" yaml:"behind_proxy" mapstructure:"behind_proxy" default:"false"`

	// DisableURLQueries disables the ability to override config via URL queries
	DisableURLQueries bool `json:"disableURLQueries" yaml:"disable_url_queries" mapstructure:"disable_url_queries" default:"false"`
	// DisableConfigEndpoint disables the config endpoint
	DisableConfigEndpoint bool `json:"disableConfigEndpoint"  yaml:"disable_config_endpoint" mapstructure:"disable_config_endpoint" default:"false"`

	// EnableURLBuilder enables the URL builder page and endpoints
	EnableURLBuilder bool `json:"enableURLBuilder" yaml:"enable_url_builder" mapstructure:"enable_url_builder" default:"false"`

	// WatchConfig if kiosk should watch config file for changes
	WatchConfig bool `json:"watchConfig" yaml:"watch_config" mapstructure:"watch_config" default:"false"`

	// Cache enable/disable api call and image caching
	Cache bool `json:"cache" yaml:"cache" mapstructure:"cache" default:"true"`

	// PreFetch fetch and cache an image in the background
	PreFetch bool `json:"preFetch" yaml:"prefetch" mapstructure:"prefetch" default:"true"`

	// AssetWeighting use weighting when picking assets
	AssetWeighting bool `json:"assetWeighting" yaml:"asset_weighting" mapstructure:"asset_weighting" default:"true"`

	// debug modes
	Debug        bool `json:"debug" yaml:"debug" mapstructure:"debug" default:"false"`
	DebugVerbose bool `json:"debugVerbose" yaml:"debug_verbose" mapstructure:"debug_verbose" default:"false"`

	DemoMode bool `json:"-" yaml:"-" mapstructure:"demo_mode" default:"false"`
}

type WeatherLocation struct {
	Name      string `yaml:"name" mapstructure:"name" redact:"true"`
	Lat       string `yaml:"lat" mapstructure:"lat" redact:"true"`
	Lon       string `yaml:"lon" mapstructure:"lon" redact:"true"`
	API       string `yaml:"api" mapstructure:"api" redact:"true"`
	Unit      string `yaml:"unit" mapstructure:"unit" redact:"true"`
	Lang      string `yaml:"lang" mapstructure:"lang" redact:"true"`
	Forecast  bool   `yaml:"forecast" mapstructure:"forecast" default:"false"`
	RoundTemp bool   `yaml:"round_temperature" mapstructure:"round_temperature" default:"false"`
	Default   bool   `yaml:"default" mapstructure:"default"`
}

type Webhook struct {
	URL    string `json:"url" yaml:"url" mapstructure:"url" redact:"true"`
	Event  string `json:"event" yaml:"event" mapstructure:"event"`
	Secret string `json:"secret" yaml:"secret" mapstructure:"secret" redact:"true"`
}

type Webhooks []Webhook

func (w Webhooks) ContainsEvent(event string) bool {
	for _, webhook := range w {
		if webhook.Event == event {
			return true
		}
	}
	return false
}

// ClientData represents the client-specific dimensions received from the frontend.
type ClientData struct {

	// FullyVersion stores the version info for Fully Kiosk Browser
	FullyVersion string `json:"fully_version" query:"fully_version" form:"fully_version"`
	// FullyWebviewVersion stores the webview version for Fully Kiosk Browser
	FullyWebviewVersion string `json:"fully_webview_version" query:"fully_webview_version" form:"fully_webview_version"`
	// FullyAndroidVersion stores the Android version for Fully Kiosk Browser
	FullyAndroidVersion string `json:"fully_android_version" query:"fully_android_version" form:"fully_android_version"`

	// Width represents the client's viewport width in pixels
	Width int `json:"client_width" query:"client_width" form:"client_width"`
	// Height represents the client's viewport height in pixels
	Height int `json:"client_height" query:"client_height" form:"client_height"`

	// FullyScreenOrientation stores the screen orientation from Fully Kiosk Browser
	FullyScreenOrientation int `json:"fully_screen_orientation" query:"fully_screen_orientation" form:"fully_screen_orientation"`
	// FullyScreenBrightness stores the screen brightness level from Fully Kiosk Browser
	FullyScreenBrightness int `json:"fully_screen_brightness" query:"fully_screen_brightness" form:"fully_screen_brightness"`
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
	// configLastModTime stores the last modification time of the configuration file
	configLastModTime time.Time `json:"-" yaml:"-"`
	// V is the viper instance used for configuration management
	V *viper.Viper `json:"-" yaml:"-"`
	// mu is a mutex used to ensure thread-safe access to the configuration
	mu *sync.RWMutex `json:"-" yaml:"-"`

	// ImmichUsersAPIKeys a map of usernames to their respective api keys for accessing Immich
	ImmichUsersAPIKeys map[string]string `json:"-" yaml:"immich_users_api_keys" mapstructure:"immich_users_api_keys" default:"{}" redact:"true"`
	// User the user from ImmichUsersAPIKeys to use when fetching images. If not set, it will use the default ImmichAPIKey
	User []string `json:"user" yaml:"user" mapstructure:"user" query:"user" form:"user" default:"[]" redact:"true"`
	// ReloadTimeStamp timestamp for when the last client reload was called for
	ReloadTimeStamp string `json:"-" yaml:"-"`
	// configHash stores the SHA-256 hash of the configuration file
	configHash string `json:"-" yaml:"-"`
	// SystemLang the system language
	SystemLang monday.Locale `json:"-" yaml:"-" default:"en_GB"`

	// History past shown images
	History []string `json:"history" yaml:"-" form:"history" default:"[]"`

	// ClientData data sent from the client with data regarding itself
	ClientData ClientData `yaml:"-"`

	// ImmichAPIKey Immich key to access assets
	ImmichAPIKey string `json:"-" yaml:"immich_api_key" mapstructure:"immich_api_key" default:"" redact:"true"`
	// ImmichURL Immuch base url
	ImmichURL string `json:"-" yaml:"immich_url" mapstructure:"immich_url" default:"" redact:"true"`

	// ImmichExternalURL specifies an external URL for Immich access. This can be used when
	// the Immich instance is accessed through a different URL externally vs internally
	// (e.g., when using reverse proxies or different network paths)
	ImmichExternalURL string `json:"-" yaml:"immich_external_url" mapstructure:"immich_external_url" default:"" redact:"true"`

	// ShowTime whether to display clock
	ShowTime bool `json:"showTime" yaml:"show_time" mapstructure:"show_time" query:"show_time" form:"show_time" default:"false"`
	// TimeFormat whether to use 12 of 24 hour format for clock
	TimeFormat string `json:"timeFormat" yaml:"time_format" mapstructure:"time_format" query:"time_format" form:"time_format" default:"24"`
	// ShowDate whether to display date
	ShowDate bool `json:"showDate" yaml:"show_date" mapstructure:"show_date" query:"show_date" form:"show_date" default:"false"`
	//  DateFormat format for date
	DateFormat string `json:"dateFormat" yaml:"date_format" mapstructure:"date_format" query:"date_format" form:"date_format" default:"DD/MM/YYYY"`
	// ClockSource source of clock time
	ClockSource string `json:"clockSource" yaml:"clock_source" mapstructure:"clock_source" query:"clock_source" form:"clock_source" default:"client"`

	// Duration in seconds to display assets
	Duration int `json:"duration" yaml:"duration" mapstructure:"duration" query:"duration" form:"duration" default:"60"`
	// DisableScreensaver asks browser to disable screensaver
	DisableScreensaver bool `json:"disableScreensaver" yaml:"disable_screensaver" mapstructure:"disable_screensaver" query:"disable_screensaver" form:"disable_screensaver" default:"false"`

	// SelectedUser selected user from User for the specific request
	SelectedUser string `json:"selectedUser" yaml:"-" default:""`
	// MenuPosition position of menu
	MenuPosition string `json:"menuPosition" yaml:"menu_position" mapstructure:"menu_position" query:"menu_position" form:"menu_position" default:"top"`
	// OptimizeImages tells Kiosk to optimize images
	OptimizeImages bool `json:"optimize_images" yaml:"optimize_images" mapstructure:"optimize_images" query:"optimize_images" form:"optimize_images" default:"false"`
	// UseGpu tells Kiosk to use GPU where possible
	UseGpu bool `json:"useGpu" yaml:"use_gpu" mapstructure:"use_gpu" query:"use_gpu" form:"use_gpu" default:"true"`
	// BurnInInterval specifies (in minutes) the interval between burn-in cycles. 0 disables.
	BurnInInterval int `json:"burnInInterval" yaml:"burn_in_interval" mapstructure:"burn_in_interval" query:"burn_in_interval" form:"burn_in_interval" default:"0"`
	// BurnInDuration specifies how long (in seconds) the burn-in effect lasts
	BurnInDuration int `json:"burnInDuration" yaml:"burn_in_duration" mapstructure:"burn_in_duration" query:"burn_in_duration" form:"burn_in_duration" default:"30"`
	// BurnInOpacity specifies the opacity (0-100) of UI elements during burn-in shift
	BurnInOpacity int `json:"burnInOpacity" yaml:"burn_in_opacity" mapstructure:"burn_in_opacity" query:"burn_in_opacity" form:"burn_in_opacity" default:"70"`

	// ShowArchived allow archived image to be displayed
	ShowArchived bool `json:"showArchived" yaml:"show_archived" mapstructure:"show_archived" query:"show_archived" form:"show_archived" default:"false"`

	// IDs of people to display
	People           []string `json:"people" yaml:"people" mapstructure:"people" query:"person" form:"person" default:"[]" redact:"true"`
	RequireAllPeople bool     `json:"requireAllPeople" yaml:"require_all_people" mapstructure:"require_all_people" query:"require_all_people" form:"require_all_people" default:"false"`
	ExcludedPeople   []string `json:"excludedPeople" yaml:"excluded_people" mapstructure:"excluded_people" query:"exclude_person" form:"exclude_person" default:"[]" redact:"true"`

	// IDs of album(s) to display
	Albums []string `json:"albums" yaml:"albums" mapstructure:"albums" query:"album" form:"album" default:"[]" redact:"true"`
	// AlbumOrder specifies the order in which album assets are displayed.
	AlbumOrder     string   `json:"album_order" yaml:"album_order" mapstructure:"album_order" query:"album_order" form:"album_order" default:"random"`
	ExcludedAlbums []string `json:"excluded_albums" yaml:"excluded_albums" mapstructure:"excluded_albums" query:"exclude_album" form:"exclude_album" default:"[]" redact:"true"`

	// Dates date filter
	Dates []string `json:"dates" yaml:"dates" mapstructure:"dates" query:"date" form:"date" default:"[]"`

	// Tags Name of tag to display
	Tags         []string `json:"tags" yaml:"tags" mapstructure:"tags" query:"tag" form:"tag" default:"[]" lowercase:"true" redact:"true"`
	ExcludedTags []string `json:"excluded_tags" yaml:"excluded_tags" mapstructure:"excluded_tags" query:"exclude_tag" form:"exclude_tag" default:"[]" lowercase:"true" redact:"true"`

	// ExcludedPartners ID(s) of partner to exclude
	ExcludedPartners []string `json:"excluded_partners" yaml:"excluded_partners" mapstructure:"excluded_partners" query:"exclude_partner" form:"exclude_partner" default:"[]" redact:"true"`

	// Memories show memories
	Memories       bool    `json:"memories" yaml:"memories" mapstructure:"memories" query:"memories" form:"memories" default:"false"`
	PastMemoryDays int     `json:"pastMemoryDays" yaml:"past_memory_days" mapstructure:"past_memory_days" query:"past_memory_days" form:"past_memory_days" default:"0"`
	MemoryWeight   float64 `json:"memoryWeight" yaml:"memory_weight" mapstructure:"memory_weight" default:"1.0"`

	// DateFilter filter certain asset bucket assets by date
	DateFilter string `json:"dateFilter" yaml:"date_filter" mapstructure:"date_filter" query:"date_filter" form:"date_filter" default:""`

	// ShowClearCacheButton display a button to clear cache
	ShowClearCacheButton bool `json:"showClearCacheButton" yaml:"show_clear_cache_button" mapstructure:"show_clear_cache_button" query:"show_clear_cache_button" form:"show_clear_cache_button" default:"false"`
	// ShowProgressBar display a progress bar
	ShowProgressBar bool `json:"showProgressBar" yaml:"show_progress_bar" mapstructure:"show_progress_bar" query:"show_progress_bar" form:"show_progress_bar" default:"false"`
	// ProgressBarPosition
	ProgressBarPosition string `json:"progressBarPosition" yaml:"progress_bar_position" mapstructure:"progress_bar_position" query:"progress_bar_position" form:"progress_bar_position" default:"top"`
	// DisableUI a shortcut to disable ShowTime, ShowDate, ShowImageTime and ShowImageDate
	DisableUI bool `json:"disableUi" yaml:"disable_ui" mapstructure:"disable_ui" query:"disable_ui" form:"disable_ui" default:"false"`
	// DisableNavigation remove navigation
	DisableNavigation bool `json:"disableNavigation" yaml:"disable_navigation" mapstructure:"disable_navigation" query:"disable_navigation" form:"disable_navigation" default:"false"`
	// Frameless remove border on frames
	Frameless bool `json:"frameless" yaml:"frameless" mapstructure:"frameless" query:"frameless" form:"frameless" default:"false"`
	// FramePadding add padding to Kiosk
	FramePadding []int `json:"framePadding" yaml:"frame_padding" mapstructure:"frame_padding" query:"frame_padding" form:"frame_padding" default:"[]"`
	// HideCursor hide cursor via CSS
	HideCursor bool `json:"hideCursor" yaml:"hide_cursor" mapstructure:"hide_cursor" query:"hide_cursor" form:"hide_cursor" default:"false"`
	// FontSize the base font size as a percentage
	FontSize int `json:"fontSize" yaml:"font_size" mapstructure:"font_size" query:"font_size" form:"font_size" default:"100"`
	// BackgroundBlur whether to display blurred image as background
	BackgroundBlur bool `json:"backgroundBlur" yaml:"background_blur" mapstructure:"background_blur" query:"background_blur" form:"background_blur" default:"true"`
	// BackgroundBlurAmount the amount of blur to apply
	BackgroundBlurAmount int `json:"backgroundBlurAmount" yaml:"background_blur_amount" mapstructure:"background_blur_amount" query:"background_blur_amount" form:"background_blur_amount" default:"10"`
	// Theme which theme to use
	Theme string `json:"theme" yaml:"theme" mapstructure:"theme" query:"theme" form:"theme" default:"fade" lowercase:"true"`
	// Layout which layout to use
	Layout string `json:"layout" yaml:"layout" mapstructure:"layout" query:"layout" form:"layout" default:"single" lowercase:"true"`

	// SleepStart when to start sleep mode
	SleepStart string `json:"sleepStart" yaml:"sleep_start" mapstructure:"sleep_start" query:"sleep_start" form:"sleep_start" default:""`
	// SleepEnd when to exit sleep mode
	SleepEnd string `json:"sleepEnd" yaml:"sleep_end" mapstructure:"sleep_end" query:"sleep_end" form:"sleep_end" default:""`
	// SleepIcon display sleep icon
	SleepIcon bool `json:"sleepIcon" yaml:"sleep_icon" mapstructure:"sleep_icon" query:"sleep_icon" form:"sleep_icon" default:"true"`
	// SleepDimScreen dim screen when sleep mode is active (for Fully Kiosk Browser)
	SleepDimScreen bool `json:"sleepDimScreen" yaml:"sleep_dim_screen" mapstructure:"sleep_dim_screen" query:"sleep_dim_screen" form:"sleep_dim_screen" default:"false"`
	// SleepDisable disable sleep via url queries
	DisableSleep bool `json:"disableSleep" yaml:"disable_sleep" query:"disable_sleep" form:"disable_sleep" default:"false"`

	// Transition which transition to use none|fade|cross-fade
	Transition string `json:"transition" yaml:"transition" mapstructure:"transition" query:"transition" form:"transition" default:"" lowercase:"true"`
	// FadeTransitionDuration sets the length of the fade transition
	FadeTransitionDuration float32 `json:"fadeTransitionDuration" yaml:"fade_transition_duration" mapstructure:"fade_transition_duration" query:"fade_transition_duration" form:"fade_transition_duration" default:"1"`
	// CrossFadeTransitionDuration sets the length of the cross-fade transition
	CrossFadeTransitionDuration float32 `json:"crossFadeTransitionDuration" yaml:"cross_fade_transition_duration" mapstructure:"cross_fade_transition_duration" query:"cross_fade_transition_duration" form:"cross_fade_transition_duration" default:"1"`

	// ImageFit the fit style for main image
	ImageFit string `json:"imageFit" yaml:"image_fit" mapstructure:"image_fit" query:"image_fit" form:"image_fit" default:"contain" lowercase:"true"`
	// ImageEffect which effect to apply to image (if any)
	ImageEffect string `json:"imageEffect" yaml:"image_effect" mapstructure:"image_effect" query:"image_effect" form:"image_effect" default:"" lowercase:"true"`
	// ImageEffectAmount the amount of effect to apply
	ImageEffectAmount int `json:"imageEffectAmount" yaml:"image_effect_amount" mapstructure:"image_effect_amount" query:"image_effect_amount" form:"image_effect_amount" default:"120"`
	// UseOriginalImage use the original image
	UseOriginalImage bool `json:"useOriginalImage" yaml:"use_original_image" mapstructure:"use_original_image" query:"use_original_image" form:"use_original_image" default:"false"`

	// ShowVideos whether to display videos
	ShowVideos bool `json:"showVideos" yaml:"show_videos" mapstructure:"show_videos" query:"show_videos" form:"show_videos" default:"false"`
	// LivePhotos show live photos
	LivePhotos         bool `json:"livePhotos" yaml:"live_photos" mapstructure:"live_photos" query:"live_photos" form:"live_photos" default:"false"`
	LivePhotoLoopDelay int  `json:"livePhotoLoopDelay" yaml:"live_photo_loop_delay" mapstructure:"live_photo_loop_delay" query:"live_photo_loop_delay" form:"live_photo_loop_delay" default:"0"`

	// ShowOwner whether to display owner
	ShowOwner bool `json:"showOwner" yaml:"show_owner" mapstructure:"show_owner" query:"show_owner" form:"show_owner" default:"false"`
	// ShowAlbumName whether to display the album name
	ShowAlbumName bool `json:"showAlbumName" yaml:"show_album_name" mapstructure:"show_album_name" query:"show_album_name" form:"show_album_name" default:"false"`
	// ShowPersonName whether to display the person name
	ShowPersonName bool `json:"showPersonName" yaml:"show_person_name" mapstructure:"show_person_name" query:"show_person_name" form:"show_person_name" default:"false"`
	// ShowPersonAge whether to display the person age
	ShowPersonAge bool `json:"showPersonAge" yaml:"show_person_age" mapstructure:"show_person_age" query:"show_person_age" form:"show_person_age" default:"false"`
	// ShowAgeYearUnit whether to show year unit
	ShowAgeYearUnit bool `json:"showAgeYearUnit" yaml:"show_age_year_unit" mapstructure:"show_age_year_unit" query:"show_age_year_unit" form:"show_age_year_unit" default:"false"`
	// AgeSwitchToYearsAfter when to switch from months to years
	AgeSwitchToYearsAfter int `json:"ageSwitchToYearsAfter" yaml:"age_switch_to_years_after" mapstructure:"age_switch_to_years_after" query:"age_switch_to_years_after" form:"age_switch_to_years_after" default:"1"`
	// ShowImageTime whether to display image time
	ShowImageTime bool `json:"showImageTime" yaml:"show_image_time" mapstructure:"show_image_time" query:"show_image_time" form:"show_image_time" default:"false"`
	// ImageTimeFormat  whether to use 12 of 24 hour format
	ImageTimeFormat string `json:"imageTimeFormat" yaml:"image_time_format" mapstructure:"image_time_format" query:"image_time_format" form:"image_time_format" default:""`
	// ShowImageDate whether to display image date
	ShowImageDate bool `json:"showImageDate" yaml:"show_image_date" mapstructure:"show_image_date" query:"show_image_date" form:"show_image_date"  default:"false"`
	// ImageDateFormat format for image date
	ImageDateFormat string `json:"imageDateFormat" yaml:"image_date_format" mapstructure:"image_date_format" query:"image_date_format" form:"image_date_format" default:""`
	// ShowImageDescription isplay image description
	ShowImageDescription bool `json:"showImageDescription" yaml:"show_image_description" mapstructure:"show_image_description" query:"show_image_description" form:"show_image_description" default:"false"`
	// ShowImageCamera display camera model
	ShowImageCamera bool `json:"showImageCamera" yaml:"show_image_camera" mapstructure:"show_image_camera" query:"show_image_camera" form:"show_image_camera" default:"false"`
	// ShowImageExif display image exif data (f number, iso, shutter speed, Focal length)
	ShowImageExif bool `json:"showImageExif" yaml:"show_image_exif" mapstructure:"show_image_exif" query:"show_image_exif" form:"show_image_exif" default:"false"`
	// ShowImageLocation display image location data
	ShowImageLocation bool `json:"showImageLocation" yaml:"show_image_location" mapstructure:"show_image_location" query:"show_image_location" form:"show_image_location" default:"false"`
	// ShowImageQR display image QR code
	ShowImageQR bool `json:"showImageQR" yaml:"show_image_qr" mapstructure:"show_image_qr" query:"show_image_qr" form:"show_image_qr" default:"false"`
	// HideCountries hide country names in location information
	HideCountries []string `json:"hideCountries" yaml:"hide_countries" mapstructure:"hide_countries" query:"hide_countries" form:"hide_countries" default:"[]"`
	// ShowImageID display image ID
	ShowImageID bool `json:"showImageID" yaml:"show_image_id" mapstructure:"show_image_id" query:"show_image_id" form:"show_image_id" default:"false"`
	// ShowUser whether to display user
	ShowUser bool `json:"showUser" yaml:"show_user" mapstructure:"show_user" query:"show_user" form:"show_user" default:"false"`

	// ShowMoreInfo enables the display of additional information about the current image
	ShowMoreInfo bool `json:"showMoreInfo" yaml:"show_more_info" mapstructure:"show_more_info" query:"show_more_info" form:"show_more_info" default:"true"`
	// ShowMoreInfoImageLink shows a link to the original image in the additional information panel
	ShowMoreInfoImageLink bool `json:"showMoreInfoImageLink" yaml:"show_more_info_image_link" mapstructure:"show_more_info_image_link" query:"show_more_info_image_link" form:"show_more_info_image_link" default:"true"`
	// ShowMoreInfoQrCode displays a QR code linking to the original image in the additional information panel
	ShowMoreInfoQrCode bool `json:"showMoreInfoQrCode" yaml:"show_more_info_qr_code" mapstructure:"show_more_info_qr_code" query:"show_more_info_qr_code" form:"show_more_info_qr_code" default:"true"`

	// LikeButtonAction indicates the action to take when the like button is clicked
	LikeButtonAction []string `json:"likeButtonAction" yaml:"like_button_action" mapstructure:"like_button_action" query:"like_button_action" form:"like_button_action" default:"[favorite]"`
	// HideButtonAction indicates the action to take when the hide button is clicked
	HideButtonAction []string `json:"hideButtonAction" yaml:"hide_button_action" mapstructure:"hide_button_action" query:"hide_button_action" form:"hide_button_action" default:"[tag]"`

	// WeatherLocations A list of locations to fetch and display weather data from. Each location
	WeatherLocations []WeatherLocation `json:"weather" yaml:"weather" mapstructure:"weather" default:"[]"`
	// HasWeatherDefault indicates whether any weather location has been set as the default.
	HasWeatherDefault bool `json:"-" yaml:"-" default:"false"`

	Iframe []string `json:"iframe" yaml:"iframe" mapstructure:"iframe" query:"iframe" form:"iframe" default:"[]"`

	// CustomCSSClass add a class to the body tag
	CustomCSSClass string `json:"customCSSClass" yaml:"custom_css_class" mapstructure:"custom_css_class" query:"custom_css_class" form:"custom_css_class" default:""`
	// CustomCSS use custom css file
	CustomCSS bool `json:"customCSS" yaml:"custom_css" mapstructure:"custom_css" query:"custom_css" form:"custom_css" default:"true"`

	// Webhooks defines a list of webhook endpoints and their associated events that should trigger notifications.
	Webhooks Webhooks `json:"webhooks" yaml:"webhooks" mapstructure:"webhooks" default:"[]"`

	// Blacklist define a list of assets to skip
	Blacklist []string `json:"blacklist" yaml:"blacklist" mapstructure:"blacklist" default:"[]" redact:"true"`

	OfflineMode    OfflineMode `json:"offlineMode" yaml:"offline_mode" mapstructure:"offline_mode"`
	UseOfflineMode bool        `json:"useOfflineMode" yaml:"use_offline_mode" mapstructure:"use_offline_mode" query:"use_offline_mode" form:"use_offline_mode" default:"false"`

	// Kiosk settings that are unable to be changed via URL queries
	Kiosk KioskSettings `json:"kiosk" yaml:"kiosk" mapstructure:"kiosk"`
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
		{"kiosk.behind_proxy", "KIOSK_BEHIND_PROXY"},
		{"kiosk.watch_config", "KIOSK_WATCH_CONFIG"},
		{"kiosk.disable_url_queries", "KIOSK_DISABLE_URL_QUERIES"},
		{"kiosk.disable_config_endpoint", "KIOSK_DISABLE_CONFIG_ENDPOINT"},
		{"kiosk.enable_url_builder", "KIOSK_ENABLE_URL_BUILDER"},
		{"kiosk.fetched_assets_size", "KIOSK_FETCHED_ASSETS_SIZE"},
		{"kiosk.http_timeout", "KIOSK_HTTP_TIMEOUT"},
		{"kiosk.password", "KIOSK_PASSWORD"},
		{"kiosk.cache", "KIOSK_CACHE"},
		{"kiosk.prefetch", "KIOSK_PREFETCH"},
		{"kiosk.asset_weighting", "KIOSK_ASSET_WEIGHTING"},
		{"kiosk.debug", "KIOSK_DEBUG"},
		{"kiosk.debug_verbose", "KIOSK_DEBUG_VERBOSE"},
		{"kiosk.demo_mode", "KIOSK_DEMO_MODE"},
		{"kiosk.config_validation_level", "KIOSK_CONFIG_VALIDATION_LEVEL"},
		{"offline_mode.enabled", "KIOSK_OFFLINE_MODE_ENABLED"},
		{"offline_mode.number_of_assets", "KIOSK_OFFLINE_MODE_NUMBER_OF_ASSETS"},
		{"offline_mode.max_size", "KIOSK_OFFLINE_MODE_MAX_SIZE"},
		{"offline_mode.expiration_hours", "KIOSK_OFFLINE_MODE_EXPIRATION_HOURS"},
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
func isValidYAML(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var data any
	if err = yaml.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	return nil
}

// load loads yaml config file into memory, then loads ENV vars. ENV vars overwrites yaml settings.
func (c *Config) Load() error {

	if bindErr := bindEnvironmentVariables(c.V); bindErr != nil {
		log.Error("binding environment variables", "err", bindErr)
	}

	c.V.SetConfigName("config")
	c.V.SetConfigType("yaml")

	// Add potential paths for the configuration file
	c.V.AddConfigPath(".")         // Look in the current directory
	c.V.AddConfigPath("./config/") // Look in the 'config/' subdirectory
	c.V.AddConfigPath("../../")    // Look in the parent directory for testing

	if os.Getenv("KIOSK_DEMO_MODE") != "" {
		c.V.SetConfigFile("./demo.config.yaml") // use demo config file
	}

	c.V.SetEnvPrefix("kiosk")

	c.V.AutomaticEnv()

	readInConfigErr := c.V.ReadInConfig()
	if readInConfigErr != nil {
		var configFileNotFoundErr viper.ConfigFileNotFoundError
		switch {
		case errors.As(readInConfigErr, &configFileNotFoundErr):
			log.Info("Not using config.yaml")
		case isValidYAML(c.V.ConfigFileUsed()) != nil:
			log.Fatal(readInConfigErr)
		}
	} else {
		level := strings.ToLower(strings.TrimSpace(c.V.GetString("kiosk.config_validation_level")))
		if level != kiosk.ConfigValidationWarning && level != kiosk.ConfigValidationError && level != kiosk.ConfigValidationOff {
			level = kiosk.ConfigValidationError
		}

		valid := checkSchema(c.V.AllSettings(), level)
		if !valid && level != kiosk.ConfigValidationWarning {
			log.Fatal("Invalid configuration")
		}
	}

	if err := c.V.Unmarshal(c); err != nil {
		log.Error("Environment can't be loaded", "err", err)
		return err
	}

	c.checkSecrets()
	c.checkRequiredFields()
	c.checkLowercaseTaggedFields()
	c.checkAssetBuckets()
	c.checkAlbumOrder()
	c.checkExcludedAlbums()
	c.checkURLScheme()
	c.checkHideCountries()
	c.checkWeatherLocations()
	c.checkDebuging()
	c.checkFetchedAssetsSize()
	c.checkRedirects()
	c.checkOffline()
	c.checkBurnIn()

	return nil
}

// ResetBuckets clears all the asset bucket slice fields (Person, Album, Date)
// in the Config structure. This is typically used when applying new query parameters
// to ensure old values don't persist. When querying specific buckets, the previous
// values need to be cleared to avoid mixing unintended assets.
func (c *Config) ResetBuckets() {
	c.People = []string{}
	c.Albums = []string{}
	c.Dates = []string{}
	c.Tags = []string{}
}

func getHistory(queries url.Values) []string {
	h := make([]string, 0, len(queries))

	for key, query := range queries {
		if key == "history" {
			h = append(h, query...)
		}
	}

	return h
}

// ConfigWithOverrides overwrites base config with ones supplied via URL queries
func (c *Config) ConfigWithOverrides(queries url.Values, e echo.Context) error {

	if c.Kiosk.DisableURLQueries {
		c.History = getHistory(queries)
		return nil
	}

	// check for person or album in quries and empty baseconfig slice if found
	if queries.Has("person") || queries.Has("album") || queries.Has("date") || queries.Has("tag") || queries.Has("memories") {
		c.ResetBuckets()
	}

	const none = "none"

	if queries.Get("excluded_person") == none || queries.Get("excluded_people") == none {
		c.ExcludedPeople = []string{}
	}

	if queries.Get("excluded_album") == none || queries.Get("excluded_albums") == none {
		c.ExcludedAlbums = []string{}
	}

	if queries.Get("excluded_partner") == none || queries.Get("excluded_partners") == none {
		c.ExcludedPartners = []string{}
	}

	err := e.Bind(c)
	if err != nil {
		return err
	}

	c.checkExcludedAlbums()

	// Disabled features in demo mode
	if c.Kiosk.DemoMode {
		c.ShowVideos = false
		c.UseOriginalImage = false
		c.OptimizeImages = false
		c.Memories = false
		c.Kiosk.FetchedAssetsSize = 100
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

func (c *Config) SanitizedYaml() string {

	red := RedactedCopy(*c) // deep redacted clone
	out, err := yaml.Marshal(red)
	if err != nil {
		log.Error("yaml marshal", "err", err)
		return ""
	}
	return string(out)
}

// RedactedCopy returns a *new* value of the same concrete type as in,
// with any field tagged `redact:"true"` masked.  It never touches the
// original and uses only the standard library.
func RedactedCopy[T any](in T) T {
	src := reflect.ValueOf(in)
	dst := reflect.New(src.Type()).Elem()
	seen := make(map[unsafe.Pointer]reflect.Value) // cycle-guard
	copyRec(src, dst, seen)
	out, ok := dst.Interface().(T)
	if !ok {
		var zero T
		return zero
	}
	return out
}

func copyRec(src, dst reflect.Value, seen map[unsafe.Pointer]reflect.Value) {
	if !src.IsValid() {
		return
	}

	switch src.Kind() {

	case reflect.Pointer:
		if src.IsNil() {
			return
		}
		addr := unsafe.Pointer(src.Pointer())
		if v, ok := seen[addr]; ok {
			dst.Set(v)
			return
		}
		dst.Set(reflect.New(src.Elem().Type()))
		seen[addr] = dst
		copyRec(src.Elem(), dst.Elem(), seen)

	case reflect.Struct:
		for i := range src.NumField() {
			fsrc, fdst := src.Field(i), dst.Field(i)
			fieldInfo := src.Type().Field(i)
			if fieldInfo.Tag.Get("redact") == "true" {
				maskValue(fsrc, fdst)
				continue
			}
			if fdst.CanSet() {
				copyRec(fsrc, fdst, seen)
			}
		}

	case reflect.Slice, reflect.Array:
		l := src.Len()
		dslice := reflect.MakeSlice(src.Type(), l, l)
		for i := range l {
			copyRec(src.Index(i), dslice.Index(i), seen)
		}
		dst.Set(dslice)

	case reflect.Map:
		dmap := reflect.MakeMapWithSize(src.Type(), src.Len())
		for _, k := range src.MapKeys() {
			v := reflect.New(src.Type().Elem()).Elem()
			copyRec(src.MapIndex(k), v, seen)
			dmap.SetMapIndex(k, v)
		}
		dst.Set(dmap)

	default: // primitives, interfaces, chans, funcs
		dst.Set(src)
	}
}

func maskValue(src, dst reflect.Value) {
	switch dst.Kind() {
	case reflect.String:
		if src.String() == "" {
			return
		}
		dst.SetString(redactedMarker)

	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.String {
			if src.Len() == 0 {
				dst.Set(reflect.MakeSlice(src.Type(), 0, 0))
				return
			}
			out := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
			for i := range src.Len() {
				out.Index(i).SetString(redactedMarker)
			}
			dst.Set(out)
			return
		}
		dst.Set(reflect.Zero(dst.Type()))

	case reflect.Map:
		if dst.Type().Key().Kind() == reflect.String && dst.Type().Elem().Kind() == reflect.String {
			if src.Len() == 0 {
				dst.Set(reflect.MakeMap(src.Type()))
				return
			}
			out := reflect.MakeMapWithSize(src.Type(), src.Len())
			i := 0
			for range src.MapKeys() {
				key := fmt.Sprintf("%s_%d", redactedMarker, i+1)
				out.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(redactedMarker))
				i++
			}
			dst.Set(out)
			return
		}
		dst.Set(reflect.Zero(dst.Type()))

	default:
		dst.Set(reflect.Zero(dst.Type()))
	}
}
