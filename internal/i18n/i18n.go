package i18n

import (
	"embed"

	"github.com/charmbracelet/log"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

var (
	LocaleFS embed.FS
	Bundle   *i18n.Bundle
)

// Init initializes the i18n bundle and loads message files.
func Init() error {
	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	locales := []string{"en", "de", "fr", "es"}
	for _, loc := range locales {
		if _, err := Bundle.LoadMessageFileFS(LocaleFS, "locales/"+loc+".toml"); err != nil {
			return err
		}
	}

	return nil
}

// T returns a translation function for the given locale.
func T(locale string) func(string) string {
	defaultLocalizer := i18n.NewLocalizer(Bundle, "en")
	localizer := i18n.NewLocalizer(Bundle, locale)

	return func(key string) string {
		translated, err := localizer.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
		if err == nil {
			return translated
		}

		defaultTranslated, err := defaultLocalizer.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
		if err != nil {
			log.Error("failed to translate", "error", err)
			return key // fallback
		}
		return defaultTranslated
	}
}
