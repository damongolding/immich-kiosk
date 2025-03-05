// Package main is the entry point for the Immich Kiosk application.
//
// It sets up the web server, configures routes, and handles the main
// application logic for displaying and managing images in a kiosk mode.
// The package includes functionality for loading configurations, setting up
// middleware, and serving both dynamic content and static assets.
package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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
	"github.com/damongolding/immich-kiosk/internal/video"
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

	c := common.New()

	baseConfig := config.New()

	systemLang := monday.Locale(utils.SystemLanguage())
	baseConfig.SystemLang = systemLang
	log.Infof("System language set as %s", systemLang)

	configErr := baseConfig.Load()
	if configErr != nil {
		log.Error("Failed to load config", "err", configErr)
	}

	videoManager, videoManagerErr := video.New(c.Context())
	if videoManagerErr != nil {
		log.Error("Failed to initialize video manager", "err", videoManagerErr)
	}

	videoManager.MaxAge = time.Duration(10) * time.Minute

	routes.VideoManager = videoManager

	if baseConfig.Kiosk.WatchConfig {
		log.Infof("Watching %s for changes", baseConfig.V.ConfigFileUsed())
		baseConfig.WatchConfig(c.Context())
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
			Validator: func(queryPassword string, _ echo.Context) (bool, error) {
				return queryPassword == baseConfig.Kiosk.Password, nil
			},
			ErrorHandler: func(_ error, c echo.Context) error {
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

	e.GET("/assets/manifest.json", routes.Manifest)

	e.GET("/image", routes.NewRawImage(baseConfig, c))

	e.GET("/image/:imageID", routes.ImageWithID(baseConfig, c))

	e.POST("/asset/new", routes.NewAsset(baseConfig, c))

	e.POST("/asset/previous", routes.PreviousAsset(baseConfig, c))

	e.GET("/clock", routes.Clock(baseConfig))

	e.POST("/weather", routes.Weather(baseConfig))

	e.GET("/sleep", routes.Sleep(baseConfig))

	e.GET("/cache/flush", routes.FlushCache(baseConfig, c))

	e.POST("/refresh/check", routes.RefreshCheck(baseConfig))

	e.POST("/webhooks", routes.Webhooks(baseConfig, c), middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	e.GET("/video/:videoID", routes.NewVideo())

	e.GET("/:redirect", routes.Redirect(baseConfig))

	for _, w := range baseConfig.WeatherLocations {
		go weather.AddWeatherLocation(c.Context(), w)
	}

	fmt.Printf("\nKiosk listening on port %s\n\n", versionStyle(strconv.Itoa(baseConfig.Kiosk.Port)))

	go func() {
		startErr := e.Start(fmt.Sprintf(":%v", baseConfig.Kiosk.Port))
		if startErr != nil && !errors.Is(startErr, http.ErrServerClosed) {
			log.Fatal(startErr)
		}
	}()

	<-c.Context().Done()

	video.Delete()

	fmt.Println("")
	log.Info("Kiosk shutting down")
	fmt.Println("")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if shutdownErr := e.Shutdown(ctx); shutdownErr != nil {
		log.Error(shutdownErr)
	}

}
