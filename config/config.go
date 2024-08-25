package config

import (
	"encoding/json"
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

type KioskSettings struct {
	// Cache enable/disable api call caching
	Cache bool `mapstructure:"cache" default:"true"`

	// Password the password used to add authentication to the frontend
	Password string `mapstructure:"password" default:""`
}

type Config struct {
	// ImmichApiKey Immich key to access assets
	ImmichApiKey string `mapstructure:"immich_api_key" default:""`
	// ImmichUrl Immuch base url
	ImmichUrl string `mapstructure:"immich_url" default:""`

	// DisableUi a shortcut to disable ShowTime, ShowDate, ShowImageTime and ShowImageDate
	DisableUi bool `mapstructure:"disable_ui" default:"false"`

	// ShowTime whether to display clock
	ShowTime bool `mapstructure:"show_time" default:"false"`
	// TimeFormat whether to use 12 of 24 hour format for clock
	TimeFormat string `mapstructure:"time_format" default:""`
	// ShowDate whether to display date
	ShowDate bool `mapstructure:"show_date" default:"false"`
	//  DateFormat format for date
	DateFormat string `mapstructure:"date_format" default:""`

	// Refresh time between fetching new image
	Refresh int `mapstructure:"refresh" default:"60"`

	// Person ID of person to display
	Person []string `mapstructure:"person" default:""`
	// Album ID of album(s) to display
	Album []string `mapstructure:"album" default:""`

	// ImageFit the fit style for main image
	ImageFit string `mapstructure:"image_fit" default:"contain"`
	// BackgroundBlur whether to display blurred image as background
	BackgroundBlur bool `mapstructure:"background_blur" default:"true"`
	// BackgroundBlur which transition to use none|fade|cross-fade
	Transition string `mapstructure:"transition" default:""`
	// ShowProgress display a progress bar
	ShowProgress bool `mapstructure:"show_progress" default:"false"`

	// ShowImageTime whether to display image time
	ShowImageTime bool `mapstructure:"show_image_time" default:"false"`
	// ImageTimeFormat  whether to use 12 of 24 hour format
	ImageTimeFormat string `mapstructure:"image_time_format" default:""`
	// ShowImageDate whether to display image date
	ShowImageDate bool `mapstructure:"show_image_date"  default:"false"`
	// ImageDateFormat format for image date
	ImageDateFormat string `mapstructure:"image_date_format" default:""`

	// Kiosk settings that are unable to be changed via URL queries
	Kiosk KioskSettings `mapstructure:"kiosk"`
}

const (
	defaultImmichPort = "2283"
	defaultScheme     = "http://"
)

// checkUrlScheme checks given url has correct scheme and adds http:// if non if found
func (c *Config) checkUrlScheme(config Config) Config {
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

	return config
}

func setDefaultValue(field reflect.StructField, recursive ...string) {

	mapStructure := field.Tag.Get("mapstructure")

	if len(recursive) != 0 {
		recursive = append(recursive, mapStructure)
		mapStructure = strings.Join(recursive, ".")
	}

	defaultValue := field.Tag.Get("default")

	switch field.Type.Kind() {
	case reflect.Bool:
		value, _ := strconv.ParseBool(defaultValue)
		viper.SetDefault(mapStructure, value)
	case reflect.String:
		viper.SetDefault(mapStructure, defaultValue)
	case reflect.Int:
		value, _ := strconv.ParseInt(defaultValue, 10, 64)
		viper.SetDefault(mapStructure, value)
	case reflect.Float64:
		value, _ := strconv.ParseFloat(defaultValue, 64)
		viper.SetDefault(mapStructure, value)
	default:
		viper.SetDefault(mapStructure, reflect.New(field.Type).Elem())
	}
}

func setDefaults(s interface{}, recursive ...string) {
	val := reflect.ValueOf(s).Elem()
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := val.Field(i)

		mapstructureTag := field.Tag.Get("mapstructure")

		if fieldValue.Kind() == reflect.Struct {
			// Recurse for nested structs
			if len(recursive) != 0 {
				recursive = append(recursive, mapstructureTag)
			}
			setDefaults(fieldValue.Addr().Interface(), recursive...)
		} else {
			defaultTag := field.Tag.Get("default")
			fmt.Printf("Field: %s, mapstructure: %s, default: %s\n", field.Name, mapstructureTag, defaultTag)

			setDefaultValue(field, recursive...) // Set default value based on type
		}
	}
}

// Load loads config file
func (c *Config) Load() error {

	var config Config

	// Defaults
	viper.SetDefault("immich_api_key", "")
	viper.SetDefault("immich_url", "")
	viper.SetDefault("password", "")
	viper.SetDefault("disable_ui", false)
	viper.SetDefault("show_time", false)
	viper.SetDefault("time_format", "")
	viper.SetDefault("show_date", false)
	viper.SetDefault("date_format", "")
	viper.SetDefault("refresh", 60)
	viper.SetDefault("album", "")
	viper.SetDefault("person", []string{})
	viper.SetDefault("image_fit", "contain")
	viper.SetDefault("background_blur", true)
	viper.SetDefault("transition", "")
	viper.SetDefault("show_progress", false)
	viper.SetDefault("show_image_time", false)
	viper.SetDefault("image_time_format", "")
	viper.SetDefault("show_image_date", false)
	viper.SetDefault("image_date_format", "")

	viper.SetDefault("kiosk.cache", true)
	viper.SetDefault("kiosk.password", "")

	viper.BindEnv("kiosk.password", "KIOSK_PASSWORD")
	viper.BindEnv("kiosk.cache", "KIOSK_CACHE")

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

	config = c.checkUrlScheme(config)

	*c = config

	return nil
}

// ConfigWithOverrides overwrites base config with ones supplied via URL queries
func (c *Config) ConfigWithOverrides(queries url.Values) Config {

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

func (c *Config) String() string {
	out, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		log.Error("", "err", err)
	}
	return string(out)
}
