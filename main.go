package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"text/template"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/damongolding/immich-kiosk/routes"
)

var version string

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func init() {
	debugModeEnv := os.Getenv("DEBUG")
	debugMode, _ := strconv.ParseBool(debugModeEnv)

	if debugMode {
		log.SetLevel(log.DebugLevel)
		log.Debug("DEBUG mode on")
	}

}

func main() {

	fmt.Println(smallBanner)
	fmt.Print("Version ", version, "\n\n")

	e := echo.New()

	e.HideBanner = true

	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.Static("/assets", "public/assets")

	e.GET("/", routes.Home)

	e.GET("/new", routes.NewImage)

	err := e.Start(":3000")
	if err != nil {
		log.Fatal(err)
	}

}
