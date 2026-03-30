package routes

import (
	"errors"
	"net/http"
	"net/url"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/labstack/echo/v5"
)

// ManifestJSON represents the JSON structure for a web app manifest
type ManifestJSON struct {
	Name            string          `json:"name"`             // Full name of the web application
	ShortName       string          `json:"short_name"`       // Short name for the app, used on home screens
	Description     string          `json:"description"`      // Description of the web application
	StartURL        string          `json:"start_url"`        // URL that loads when app is launched
	Scope           string          `json:"scope"`            // Navigation scope that remains within the app
	Display         string          `json:"display"`          // Preferred display mode for the website
	BackgroundColor string          `json:"background_color"` // Background color for splash screen
	ThemeColor      string          `json:"theme_color"`      // Theme color for the application
	Icons           []ManifestIcons `json:"icons"`            // Array of icons for different sizes
}

// ManifestIcons represents the icon information in the web app manifest
type ManifestIcons struct {
	Src   string `json:"src"`   // Path to the icon file
	Sizes string `json:"sizes"` // Dimensions of the icon
	Type  string `json:"type"`  // MIME type of the icon
}

// manifestStartURL builds a safe start_url from the referer, keeping only the
// "client" query parameter so transient or sensitive params are not baked into
// the installed PWA launch URL.
func manifestStartURL(referer *url.URL) string {
	startURL := referer.EscapedPath()
	if startURL == "" {
		startURL = "/"
	}
	filtered := url.Values{}
	if client := referer.Query().Get("client"); client != "" {
		filtered.Set("client", client)
	}
	if encoded := filtered.Encode(); encoded != "" {
		startURL += "?" + encoded
	}
	return startURL
}

// Manifest generates and returns a web app manifest JSON response
// based on the request referer URL. It sets appropriate headers
// and formats the manifest data according to the Web App Manifest spec.
func Manifest(c *echo.Context) error {
	refererURL := c.Request().Referer()
	if refererURL == "" {
		refererURL = "/"
	}

	referer, err := url.Parse(refererURL)
	if err != nil {
		log.Error("parsing URL", "url", refererURL, "err", err)
		return errors.New("could not read URL. Is it formatted correctly?")
	}

	manifest := &ManifestJSON{
		Name:            "Immich Kiosk",
		ShortName:       "Kiosk",
		Description:     "Immich Kiosk is a lightweight slideshow for running on kiosk devices and browsers that uses Immich as a data source.",
		StartURL:        manifestStartURL(referer),
		Scope:           "/",
		Display:         "fullscreen",
		BackgroundColor: "#000000",
		ThemeColor:      "#1f262f",
		Icons: []ManifestIcons{
			{
				Src:   "/assets/images/android-chrome-192x192.png",
				Sizes: "192x192",
				Type:  kiosk.MimeTypePng,
			},
			{
				Src:   "/assets/images/android-chrome-512x512.png",
				Sizes: "512x512",
				Type:  kiosk.MimeTypePng,
			},
		},
	}

	c.Response().Header().Set("Content-Type", "application/json")

	return c.JSON(http.StatusOK, *manifest)
}
