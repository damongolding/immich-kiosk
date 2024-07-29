package main

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/damongolding/immich-kiosk/routes"
)

// version current build version number
var version string

//go:embed public
var public embed.FS

func init() {
	routes.KioskVersion = version

	debugModeEnv := os.Getenv("KIOSK_DEBUG")
	debugMode, _ := strconv.ParseBool(debugModeEnv)

	if debugMode {
		log.SetLevel(log.DebugLevel)
		log.Debug("DEBUG mode on")
		zone, _ := time.Now().Zone()
		log.Debug("🕐", "current_time", time.Now().Format(time.Kitchen), "current_zone", zone)
	}

}

func main() {

	fmt.Println(smallBanner)
	fmt.Print("Version ", version, "\n\n")

	e := echo.New()

	// hide echos default banner
	e.HideBanner = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// CSS cache busting
	e.FileFS("/assets/css/style.*.css", "public/assets/css/style.css", public)

	// serve embdedd staic assets
	e.StaticFS("/assets", echo.MustSubFS(public, "public/assets"))

	e.GET("/", routes.Home)

	e.GET("/image", routes.NewImage)

	e.GET("/clock", routes.Clock)

	err := e.Start(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
