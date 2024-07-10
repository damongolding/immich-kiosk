package config

import (
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ImmichApiKey   string `yaml:"immich_api_key"`
	ImmichUrl      string `yaml:"immich_url"`
	Refresh        int    `yaml:"refresh"`
	Person         string `yaml:"person"`
	Album          string `yaml:"album"`
	FillScreen     bool   `yaml:"fill_screen"`
	ShowDate       bool   `yaml:"show_date"`
	BackgroundBlur bool   `yaml:"background_blur"`
	Transition     string `yaml:"transition"`
}

// Load loads config file
func (c *Config) Load() error {
	config := Config{
		Refresh:    20,
		FillScreen: true,
	}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	*c = config

	return nil
}

func (c *Config) ConfigWithOverrides(queries url.Values) Config {

	configWithOverrides := c

	v := reflect.ValueOf(configWithOverrides).Elem()

	// Loop through the map and update struct fields
	for key, values := range queries {
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
