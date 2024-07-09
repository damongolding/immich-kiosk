package main

import (
	"io"
	"text/template"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/damongolding/immich-frame/routes"
)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {

	log.SetLevel(log.DebugLevel)

	e := echo.New()

	e.HideBanner = true

	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e.Use(middleware.Recover())

	e.Static("/css", "public/css")

	e.GET("/", routes.Home)

	e.GET("/new", routes.NewImage)

	err := e.Start(":3000")
	if err != nil {
		log.Fatal(err)
	}

}
