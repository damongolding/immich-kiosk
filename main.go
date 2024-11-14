// Package main is the entry point for the Immich Kiosk application.
//
// It sets up the web server, configures routes, and handles the main
// application logic for displaying and managing images in a kiosk mode.
// The package includes functionality for loading configurations, setting up
// middleware, and serving both dynamic content and static assets.

//go:generate go run main_build_time.go
//go:build !generate
// +build !generate

package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/routes"
	"github.com/damongolding/immich-kiosk/weather"
)

// version current build version number
var version string

//go:embed frontend/public
var public embed.FS

func init() {
	routes.KioskVersion = version
}

func main() {

	fmt.Println(kioskBanner)
	versionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5af78e")).Render
	fmt.Print("Version ", versionStyle(version), "\n\n")

	log.SetTimeFormat("15:04:05")

	baseConfig := config.New()

	err := baseConfig.Load()
	if err != nil {
		log.Error("Failed to load config", "err", err)
	}

	if baseConfig.Kiosk.WatchConfig {
		log.Infof("Watching %s for changes", baseConfig.V.ConfigFileUsed())
		baseConfig.WatchConfig()
	}

	if baseConfig.Kiosk.Debug {

		log.SetLevel(log.DebugLevel)
		if baseConfig.Kiosk.DebugVerbose {
			log.Debug("DEBUG VERBOSE mode on")
		} else {
			log.Debug("DEBUG mode on")
		}

		zone, _ := time.Now().Zone()
		log.Debug("🕐", "current_time", time.Now().Format(time.Kitchen), "current_zone", zone)
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	if baseConfig.Kiosk.Password != "" {
		e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Skipper: func(c echo.Context) bool {
				// skip auth for assets
				return strings.HasPrefix(c.Request().URL.String(), "/assets")
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

	e.GET("/", routes.Home(baseConfig, false))

	e.GET("/desktop", routes.Home(baseConfig, true))

	e.GET("/image", routes.NewRawImage(baseConfig))

	e.POST("/image", routes.NewImage(baseConfig))

	e.POST("/image/previous", routes.PreviousImage(baseConfig))

	e.GET("/clock", routes.Clock(baseConfig))

	e.GET("/weather", routes.Weather(baseConfig))

	e.GET("/sleep", routes.Sleep(baseConfig))

	e.GET("/cache/flush", routes.FlushCache)

	e.POST("/refresh/check", routes.RefreshCheck(baseConfig))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for _, w := range baseConfig.WeatherLocations {
		go weather.AddWeatherLocation(ctx, w)
	}

	fmt.Printf("\nKiosk listening on port %s\n\n", versionStyle(fmt.Sprintf("%v", baseConfig.Kiosk.Port)))

	go func() {
		err = e.Start(fmt.Sprintf(":%v", baseConfig.Kiosk.Port))
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	fmt.Println("Kiosk shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

}
