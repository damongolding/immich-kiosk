// Package main is the entry point for the Immich Kiosk application.
//
// It sets up the web server, configures routes, and handles the main
// application logic for displaying and managing images in a kiosk mode.
// The package includes functionality for loading configurations, setting up
// middleware, and serving both dynamic content and static assets.
package main

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/routes"
)

// version current build version number
var version string

//go:embed frontend/public
var public embed.FS

func init() {
	routes.KioskVersion = version
}

func main() {

	baseConfig := config.New()

	err := baseConfig.Load()
	if err != nil {
		log.Error("Failed to load config", "err", err)
	}

	if baseConfig.Kiosk.Debug {
		log.SetTimeFormat("15:04:05")

		log.SetLevel(log.DebugLevel)
		if baseConfig.Kiosk.DebugVerbose {
			log.Debug("DEBUG VERBOSE mode on")
		} else {
			log.Debug("DEBUG mode on")
		}

		zone, _ := time.Now().Zone()
		log.Debug("🕐", "current_time", time.Now().Format(time.Kitchen), "current_zone", zone)
	}

	fmt.Println(kioskBanner)
	versionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5af78e")).Render(version)
	fmt.Print("Version ", versionStyle, "\n\n")

	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	if baseConfig.Kiosk.Password != "" {
		e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Skipper: func(c echo.Context) bool {
				// skip auth for assets
				if strings.HasPrefix(c.Request().URL.String(), "/assets") {
					return true
				}
				return false
			},
			KeyLookup: "query:password,form:password",
			Validator: func(queryPassword string, c echo.Context) (bool, error) {
				return queryPassword == baseConfig.Kiosk.Password, nil
			},
			ErrorHandler: func(err error, c echo.Context) error {
				return c.String(http.StatusUnauthorized, "Unauthorized")
			},
		}))
	}

	// CSS cache busting
	e.FileFS("/assets/css/kiosk.*.css", "frontend/public/assets/css/kiosk.css", public)

	// JS cache busting
	e.FileFS("/assets/js/kiosk.*.js", "frontend/public/assets/js/kiosk.js", public)

	// serve embdedd staic assets
	e.StaticFS("/assets", echo.MustSubFS(public, "frontend/public/assets"))

	e.GET("/", routes.Home(baseConfig))

	e.GET("/image", routes.NewRawImage(baseConfig))

	e.POST("/image", routes.NewImage(baseConfig))

	e.GET("/clock", routes.Clock(baseConfig))

	err = e.Start(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
