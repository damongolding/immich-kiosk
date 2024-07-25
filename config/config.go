package config

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Config struct {
	// ImmichApiKey Immich key to access assets
	ImmichApiKey string `mapstructure:"immich_api_key"`
	// ImmichUrl Immuch base url
	ImmichUrl string `mapstructure:"immich_url"`

	// DisableUi a shortcut to disable ShowTime, ShowDate, ShowImageTime and ShowImageDate
	DisableUi bool

	// ShowTime whether to display clock
	ShowTime bool `mapstructure:"show_time"`
	// TimeFormat whether to use 12 of 24 hour format for clock
	TimeFormat string `mapstructure:"time_format"`
	// ShowDate whether to display date
	ShowDate bool `mapstructure:"show_date"`
	//  DateFormat format for date
	DateFormat string `mapstructure:"date_format"`

	// Refresh time between fetching new image
	Refresh int `mapstructure:"refresh"`
	// Person ID of person to display
	Person []string `mapstructure:"person"`
	// Album ID of album to display
	Album string `mapstructure:"album"`
	// FillScreen force image to be fullscreen
	FillScreen bool `mapstructure:"fill_screen"`

	// BackgroundBlur whether to display blurred image as background
	BackgroundBlur bool `mapstructure:"background_blur"`
	// BackgroundBlur which transistion to use none|fade|cross-fade
	Transition string `mapstructure:"transition"`
	// ShowProgress display a progress bar
	ShowProgress bool `mapstructure:"show_progress"`

	// ShowImageTime whether to display image time
	ShowImageTime bool `mapstructure:"show_image_time"`
	// ImageTimeFormat  whether to use 12 of 24 hour format
	ImageTimeFormat string `mapstructure:"image_time_format"`
	// ShowImageDate whether to display image date
	ShowImageDate bool `mapstructure:"show_image_date"`
	// ImageDateFormat format for image date
	ImageDateFormat string `mapstructure:"image_date_format"`
}

const (
	defaultImmichPort = "2283"
	defaultScheme     = "http://"
)

// parseUrl checks given url has correct formatting e.g. https://example:2283
func (c *Config) checkUrlFormat(config Config) Config {
	if config.ImmichUrl == "" || config.ImmichApiKey == "" {
		log.Fatal("Either Immich Url or Immich Api Key is missing", "ImmichUrl", config.ImmichUrl, "ImmichApiKey", config.ImmichApiKey)
	}

	// check for correct scheme
	switch {
	case strings.HasPrefix(strings.ToLower(config.ImmichUrl), "http://"):
		break
	case strings.HasPrefix(strings.ToLower(config.ImmichUrl), "https://"):
		break
	default:
		config.ImmichUrl = defaultScheme + config.ImmichUrl
	}

	u, err := url.Parse(config.ImmichUrl)
	if err != nil {
		log.Fatal("Immich URL malformed")
	}

	// Add default Immich port if no port has been specified
	if u.Port() == "" {
		// just in case URL has no port but has a colon, remove it
		host := strings.Replace(u.Host, ":", "", -1)
		// Build URL with scheme and default Immich port
		config.ImmichUrl = fmt.Sprintf("%s://%s:%s", u.Scheme, host, defaultImmichPort)
	}

	return config
}

// Load loads config file
func (c *Config) Load() error {

	var config Config

	// Defaults
	viper.SetDefault("immich_api_key", "")
	viper.SetDefault("immich_url", "")
	viper.SetDefault("disable_ui", false)
	viper.SetDefault("show_time", false)
	viper.SetDefault("time_format", "")
	viper.SetDefault("show_date", false)
	viper.SetDefault("date_format", "")
	viper.SetDefault("refresh", 60)
	viper.SetDefault("album", "")
	viper.SetDefault("person", []string{})
	viper.SetDefault("fill_screen", true)
	viper.SetDefault("background_blur", true)
	viper.SetDefault("transition", "")
	viper.SetDefault("show_progress", false)
	viper.SetDefault("show_image_time", false)
	viper.SetDefault("image_time_format", "")
	viper.SetDefault("show_image_date", false)
	viper.SetDefault("image_date_format", "")

	viper.AddConfigPath(".")
	viper.SetConfigFile("config.yaml")

	viper.SetEnvPrefix("kiosk")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Debug("config.yaml file not being used")
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Environment can't be loaded: ", "err", err)
	}

	config = c.checkUrlFormat(config)

	*c = config

	return nil
}

// ConfigWithOverrides overwrites base config with ones supplied via URL queries
func (c *Config) ConfigWithOverrides(queries url.Values) Config {

	configWithOverrides := c

	v := reflect.ValueOf(configWithOverrides).Elem()

	// Loop through the map and update struct fields
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
				for _, value := range values {

					if value == "" {
						continue
					}

					// Set field (covert to correct type if needed)
					switch field.Kind() {
					case reflect.String: // all string values should be lowercase
						lowercaseValue := strings.ToLower(value)
						field.SetString(lowercaseValue)
						break
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
