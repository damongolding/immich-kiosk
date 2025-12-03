package i18n

import (
	"embed"
	"io/fs"

	"github.com/charmbracelet/log"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

var (
	LocaleFS         embed.FS
	Bundle           *i18n.Bundle
	lang             string
	localizer        *i18n.Localizer
	defaultLocalizer *i18n.Localizer
)

// Init initializes the i18n bundle and loads message files.
func Init(systemLang string) error {
	lang = systemLang

	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	err := fs.WalkDir(LocaleFS, "locales", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		_, bundleErr := Bundle.LoadMessageFileFS(LocaleFS, path)
		if bundleErr != nil {
			return bundleErr
		}

		return nil
	})

	if err != nil {
		return err
	}

	defaultLocalizer = i18n.NewLocalizer(Bundle, "en")
	localizer = i18n.NewLocalizer(Bundle, lang)

	return nil
}

// T returns a translation function for the given locale.
func T() func(string) string {
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
