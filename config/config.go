package config

import (
	"bytes"
	_ "embed"
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
	// Refresh time between fetching new image
	Refresh int `mapstructure:"refresh"`
	// Person ID of person to display
	Person string `mapstructure:"person"`
	// Album ID of album to display
	Album string `mapstructure:"album"`
	// FillScreen force image to be fullscreen
	FillScreen bool `mapstructure:"fill_screen"`
	// ShowDate whether to display image date
	ShowDate bool `mapstructure:"show_date"`
	//  DateFormat format for date
	DateFormat string `mapstructure:"date_format"`
	// ShowTime whether to display image time
	ShowTime bool `mapstructure:"show_time"`
	// TimeFormat  whether to use 12 of 24 hour format
	TimeFormat string `mapstructure:"time_format"`
	// BackgroundBlur whether to display blurred image as background
	BackgroundBlur bool `mapstructure:"background_blur"`
	// BackgroundBlur which transistion to use none|fade|cross-fade
	Transition string `mapstructure:"transition"`
	// ShowProgress display a progress bar
	ShowProgress bool `mapstructure:"show_progress"`
}

const (
	defaultImmichPort = "2283"
	defaultScheme     = "http://"
)

//go:embed config.example.yaml
var exampleConfig []byte

// Load loads config file
func (c *Config) Load() error {

	config := Config{
		Refresh: 60,
	}

	viper.AddConfigPath(".")
	viper.SetConfigFile("config.yaml")

	viper.SetEnvPrefix("kiosk")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		// no yaml file found so lets load the example file as a base and any ENV will overwrite
		if err := viper.ReadConfig(bytes.NewBuffer(exampleConfig)); err != nil {
			log.Fatal("Config and Example config missing", "err", err)
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Environment can't be loaded: ", "err", err)
	}

	if config.ImmichUrl == "" || config.ImmichApiKey == "" {
		log.Fatal("Either Immich Url or Immich Api Key is missing", "ImmichUrl", config.ImmichUrl, "ImmichApiKey", config.ImmichApiKey)
	}

	// check for correct URL formatting
	if !strings.HasPrefix(config.ImmichUrl, "http://") || !strings.HasPrefix(config.ImmichUrl, "https://") {
		config.ImmichUrl = defaultScheme + config.ImmichUrl
	}

	u, err := url.Parse(config.ImmichUrl)
	if err != nil {
		log.Fatal("Immich URL malformed")
	}

	// Add default Immich port
	if u.Port() == "" {
		host := strings.Replace(u.Host, ":", "", -1)
		// config.ImmichUrl = u.Scheme + host + ":" + defaultImmichPort
		config.ImmichUrl = fmt.Sprintf("%s://%s:%s", u.Scheme, host, defaultImmichPort)
	}

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
			// Lets just use the first given overwrite
			value := values[0]
			if value == "" {
				continue
			}

			// format to match field names
			key = strings.ReplaceAll(key, "_", " ")
			key = cases.Title(language.English, cases.Compact).String(key)
			key = strings.ReplaceAll(key, " ", "")

			// Get the field by name
			field := v.FieldByName(key)
			if field.IsValid() && field.CanSet() {
				// Set field (covert to correct type if needed)
				switch field.Kind() {
				case reflect.String:
					// all string values should be lowercase
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
				}
			}
		}
	}

	return *configWithOverrides
}
