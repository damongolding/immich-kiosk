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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/goodsign/monday"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/i18n"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/routes"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/video"
	"github.com/damongolding/immich-kiosk/internal/weather"
)

// version current build version number
var version string

//go:embed frontend/public
var public embed.FS

//go:embed locales/*.toml
var localeFS embed.FS

//go:embed config.schema.json
var SchemaJSON string

func init() {
	routes.KioskVersion = version
	config.SchemaJSON = SchemaJSON
	i18n.LocaleFS = localeFS
}

// main initializes and starts the Immich Kiosk web server, sets up configuration, middleware, routes, and manages graceful shutdown.
func main() {

	var logLevel log.Level
	setLogLevel(&logLevel)
	log.SetLevel(logLevel)

	if logLevel == log.ErrorLevel || logLevel == log.WarnLevel {
		fmt.Println(kioskBanner)
	} else {
		log.Info(kioskBanner)
	}

	versionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5af78e")).Render
	if logLevel == log.ErrorLevel || logLevel == log.WarnLevel {
		fmt.Print("Version ", versionStyle(version), "\n\n")
	} else {
		log.Info("Version", "v", version)
		fmt.Println()
	}

	log.SetTimeFormat("15:04:05")

	c := common.New()

	baseConfig := config.New()
	baseConfig.Kiosk.Version = version

	lang := utils.SystemLanguage()
	systemLang := monday.Locale(lang)
	baseConfig.SystemLang = systemLang
	log.Info("System language", "lang", systemLang)

	i18nErr := i18n.Init(lang)
	if i18nErr != nil {
		log.Error("Failed to initialize i18n", "err", i18nErr)
	}

	configErr := baseConfig.Load()
	if configErr != nil {
		log.Error("Failed to load config", "err", configErr)
	}

	if baseConfig.Kiosk.DemoMode {
		log.Info("Demo mode enabled")
		cache.DemoMode = true
	}

	cache.Initialize()

	immich.HTTPClient.Timeout = time.Second * time.Duration(baseConfig.Kiosk.HTTPTimeout)

	videoManager, videoManagerErr := video.New(c.Context())
	if videoManagerErr != nil {
		log.Error("Failed to initialize video manager", "err", videoManagerErr)
	} else {
		videoManager.MaxAge = time.Minute * 10
		routes.VideoManager = videoManager
	}

	if baseConfig.Kiosk.WatchConfig {
		log.Info("Watching config for changes", "file", baseConfig.V.ConfigFileUsed())
		baseConfig.WatchConfig(c.Context())
	}

	if baseConfig.Kiosk.Debug {
		log.SetLevel(log.DebugLevel)
		if baseConfig.Kiosk.DebugVerbose {
			log.Debug("DEBUG VERBOSE mode on")
		} else {
			log.Debug("DEBUG mode on")
		}
	}

	zone, _ := time.Now().Zone()
	log.Debug("🕐", "current_time", time.Now().Format(time.Kitchen), "current_zone", zone)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	if baseConfig.Kiosk.BehindProxy {
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
	} else {
		e.IPExtractor = echo.ExtractIPDirect()
	}

	// Middleware
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(NoCacheMiddleware)
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
				// skip auth for assets and /health endpoint
				path := c.Request().URL.Path
				return strings.HasPrefix(path, "/assets/") || path == "/health" || path == "/favicon.ico"
			},
			KeyLookup: "header:Authorization,header:X-Api-Key,query:authsecret,query:password,form:authsecret,form:password",
			Validator: func(queryPassword string, _ echo.Context) (bool, error) {
				return queryPassword == baseConfig.Kiosk.Password, nil
			},
			ErrorHandler: func(err error, c echo.Context) error {
				if baseConfig.Kiosk.Debug || baseConfig.Kiosk.DebugVerbose {
					log.Warn("unauthorized request",
						"IP", c.RealIP(),
						"method", c.Request().Method,
						"URL", c.Request().URL.String(),
						"error", err)
				}
				return routes.RenderUnauthorized(c)
			},
		}))
	}

	// CSS cache busting
	e.FileFS("/assets/css/kiosk.*.css", "frontend/public/assets/css/kiosk.css", public, StaticCacheMiddlewareWithConfig(baseConfig))

	// JS cache busting
	e.FileFS("/assets/js/kiosk.*.js", "frontend/public/assets/js/kiosk.js", public, StaticCacheMiddlewareWithConfig(baseConfig))
	e.FileFS("/assets/js/url-builder.*.js", "frontend/public/assets/js/url-builder.js", public, StaticCacheMiddlewareWithConfig(baseConfig))

	// serve embdedd staic assets
	e.StaticFS("/assets", echo.MustSubFS(public, "frontend/public/assets"))

	if !baseConfig.Kiosk.DisableConfigEndpoint {
		e.GET("/config", func(c echo.Context) error {
			return c.String(http.StatusOK, baseConfig.SanitizedYaml())
		})
	}

	e.GET("/", routes.Home(baseConfig, c))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	if baseConfig.Kiosk.EnableURLBuilder {
		e.GET("/url-builder", routes.URLBuilderPage(baseConfig, c, false))
		e.GET("/url-builder/extended", routes.URLBuilderPage(baseConfig, c, true))
		e.POST("/url-builder/build", routes.BuildURL(baseConfig))
	}

	e.GET("/about", routes.About(baseConfig))

	e.GET("/assets/manifest.json", routes.Manifest)

	e.GET("/image", routes.Image(baseConfig, c))
	e.GET("/image/reload", routes.ImageWithReload(baseConfig))

	e.GET("/image/:imageID", routes.ImageWithID(baseConfig, c), AssetCacheMiddlewareWithConfig(baseConfig))

	e.POST("/asset/new", routes.NewAsset(baseConfig, c))

	e.POST("/asset/offline", routes.OfflineMode(baseConfig, c))
	e.POST("/asset/downloading", routes.IsDownloading)

	e.POST("/asset/previous", routes.PreviousHistoryAsset(baseConfig, c))

	e.POST("/asset/tag", routes.TagAsset(baseConfig, c))

	e.POST("/asset/like", routes.LikeAsset(baseConfig, c, true))
	e.POST("/asset/unlike", routes.LikeAsset(baseConfig, c, false))

	e.POST("/asset/hide", routes.HideAsset(baseConfig, c, true))
	e.POST("/asset/unhide", routes.HideAsset(baseConfig, c, false))

	e.POST("/clock", routes.Clock(baseConfig))

	e.POST("/weather", routes.Weather(baseConfig))

	e.GET("/sleep", routes.Sleep(baseConfig))

	e.GET("/cache/flush", routes.FlushCache(baseConfig, c))

	e.POST("/refresh/check", routes.RefreshCheck(baseConfig))

	e.POST("/webhooks", routes.Webhooks(baseConfig, c), middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	e.GET("/live/:liveID", routes.LivePhoto(baseConfig.Kiosk.DemoMode, baseConfig.Kiosk.Password))

	e.GET("/video/:videoID", routes.NewVideo(baseConfig.Kiosk.DemoMode), AssetCacheMiddlewareWithConfig(baseConfig))

	e.GET("/:redirect", routes.Redirect(baseConfig, c))

	for _, w := range baseConfig.WeatherLocations {
		if w.Forecast {
			go weather.AddWeatherLocationWithForecast(c.Context(), w)
		} else {
			go weather.AddWeatherLocation(c.Context(), w)
		}
	}

	if logLevel == log.ErrorLevel || logLevel == log.WarnLevel {
		fmt.Printf("\nKiosk listening on port %s\n\n", versionStyle(strconv.Itoa(baseConfig.Kiosk.Port)))
	} else {
		fmt.Println("")
		log.Info("Kiosk listening on", "port", baseConfig.Kiosk.Port)
		fmt.Println("")
	}

	go func() {
		startErr := e.Start(fmt.Sprintf(":%v", baseConfig.Kiosk.Port))
		if startErr != nil && !errors.Is(startErr, http.ErrServerClosed) {
			log.Fatal(startErr)
		}
	}()

	<-c.Context().Done()

	video.Delete()

	fmt.Println("")
	if logLevel == log.ErrorLevel || logLevel == log.WarnLevel {
		fmt.Println("Kiosk shutting down")
	} else {
		log.Info("Kiosk shutting down")
		fmt.Println("")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if shutdownErr := e.Shutdown(ctx); shutdownErr != nil {
		log.Error(shutdownErr)
	}
}

func setLogLevel(logLevel *log.Level) {
	logLevelStr := os.Getenv("KIOSK_LOG_LEVEL")
	switch strings.ToLower(logLevelStr) {
	case "debug":
		*logLevel = log.DebugLevel
		os.Setenv("KIOSK_DEBUG", "true")
	case "verbose":
		*logLevel = log.DebugLevel
		os.Setenv("KIOSK_DEBUG_VERBOSE", "true")
	case "info":
		*logLevel = log.InfoLevel
	case "warn", "warning":
		*logLevel = log.WarnLevel
	case "error":
		*logLevel = log.ErrorLevel
	default:
		*logLevel = log.WarnLevel
	}
}

// Middleware to set no-store for dynamic endpoints
func NoCacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		return next(c)
	}
}

// Middleware for static routes with access to baseConfig
func StaticCacheMiddlewareWithConfig(baseConfig *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {

		if baseConfig.Kiosk.Debug || baseConfig.Kiosk.DebugVerbose {
			return NoCacheMiddleware(next)
		}

		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			return next(c)
		}
	}
}

// Middleware for asset(s) routes with access to baseConfig
func AssetCacheMiddlewareWithConfig(baseConfig *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {

		if baseConfig.Kiosk.Debug || baseConfig.Kiosk.DebugVerbose {
			return NoCacheMiddleware(next)
		}

		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "private, max-age=86400, no-transform")
			return next(c)
		}
	}
}
