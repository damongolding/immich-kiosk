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
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/goodsign/monday"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/routes"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/weather"
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

	if err := common.Initialize(); err != nil {
		log.Fatal("failed to initialize common package", "error", err)
	}

	baseConfig := config.New()

	systemLang := monday.Locale(utils.SystemLanguage())
	baseConfig.SystemLang = systemLang
	log.Infof("System language set as %s", systemLang)

	err := baseConfig.Load()
	if err != nil {
		log.Error("Failed to load config", "err", err)
	}

	if baseConfig.Kiosk.WatchConfig {
		log.Infof("Watching %s for changes", baseConfig.V.ConfigFileUsed())
		baseConfig.WatchConfig(common.Context)
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

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 6,
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "image")
		},
	}))

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

	e.GET("/", routes.Home(baseConfig))

	e.GET("/image", routes.NewRawImage(baseConfig))

	e.POST("/image", routes.NewImage(baseConfig))

	e.POST("/image/previous", routes.PreviousImage(baseConfig))

	e.GET("/clock", routes.Clock(baseConfig))

	e.GET("/weather", routes.Weather(baseConfig))

	e.GET("/sleep", routes.Sleep(baseConfig))

	e.GET("/cache/flush", routes.FlushCache(baseConfig))

	e.POST("/refresh/check", routes.RefreshCheck(baseConfig))

	e.POST("/webhooks", routes.Webhooks(baseConfig), middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	e.GET("/:redirect", routes.Redirect(baseConfig))

	for _, w := range baseConfig.WeatherLocations {
		go weather.AddWeatherLocation(common.Context, w)
	}

	fmt.Printf("\nKiosk listening on port %s\n\n", versionStyle(fmt.Sprintf("%v", baseConfig.Kiosk.Port)))

	go func() {
		err = e.Start(fmt.Sprintf(":%v", baseConfig.Kiosk.Port))
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-common.Context.Done()

	fmt.Println("")
	log.Info("Kiosk shutting down")
	fmt.Println("")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

}
