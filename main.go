package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/damongolding/immich-kiosk/routes"
)

// version current build version number
var version string

// TemplateRenderer echos template render
type TemplateRenderer struct {
	templates *template.Template
}

// TemplateFuncs funcs available within template files
var TemplateFuncs = map[string]any{
	"toLower": strings.ToLower,
}

// Render use GOs template engine to render
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.Funcs(TemplateFuncs).ExecuteTemplate(w, name, data)
}

func init() {
	routes.KioskVersion = version

	debugModeEnv := os.Getenv("DEBUG")
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

	// Start template engine
	tmpl := template.New("views").Funcs(TemplateFuncs)
	tmpl, err := tmpl.ParseGlob("public/views/*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	e.Renderer = &TemplateRenderer{
		templates: tmpl,
	}

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// CSS cache busting
	e.File("/assets/css/style.*.css", "public/assets/css/style.css")

	e.Static("/assets", "public/assets")

	e.GET("/", routes.Home)

	e.GET("/image", routes.NewImage)

	e.GET("/clock", routes.Clock)

	err = e.Start(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
