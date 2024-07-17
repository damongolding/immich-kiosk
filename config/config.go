package config

import (
	"bytes"
	_ "embed"
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
	ImmichApiKey string `mapstructure:"immich_api_key"`
	ImmichUrl    string `mapstructure:"immich_url"`

	Refresh    int    `mapstructure:"refresh"`
	Person     string `mapstructure:"person"`
	Album      string `mapstructure:"album"`
	FillScreen bool   `mapstructure:"fill_screen"`

	ShowDate   bool   `mapstructure:"show_date"`
	DateFormat string `mapstructure:"date_format"`

	ShowTime   bool   `mapstructure:"show_time"`
	TimeFormat string `mapstructure:"time_format"`

	BackgroundBlur bool   `mapstructure:"background_blur"`
	Transition     string `mapstructure:"transition"`
}

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

	*c = config

	return nil
}

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
			// Lets just use the first used override
			value := values[0]
			if value == "" {
				continue
			}

			key = strings.ReplaceAll(key, "_", " ")
			key = cases.Title(language.English, cases.Compact).String(key)
			key = strings.ReplaceAll(key, " ", "")

			// Get the field by name
			field := v.FieldByName(key)
			if field.IsValid() && field.CanSet() {
				// Set field (covert to correct type if needed)
				switch field.Kind() {
				case reflect.String:
					field.SetString(value)
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
