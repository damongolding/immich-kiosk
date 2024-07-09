package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type PageData struct {
	ImageUrl   string
	Date       string
	FillScreen bool
}

type ErrorData struct {
	Message string
}

var (
	immichApiKey string
	immichUrl    string

	config Config
)

func ImageToBase64(imgBtyes []byte) (string, error) {

	var base64Encoding string

	mimeType := http.DetectContentType(imgBtyes)

	base64Encoding += fmt.Sprintf("data:%s;base64,", mimeType)

	base64Encoding += base64.StdEncoding.EncodeToString(imgBtyes)

	return base64Encoding, nil
}

func showErrorTemplate(w io.Writer, errToShow error) {
	templateFile := "templates/error.html"
	tmpl, err := template.New("error.html").ParseFiles(templateFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = tmpl.Execute(w, ErrorData{Message: errToShow.Error()})
	if err != nil {
		log.Fatal(err)
	}

}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	immichApiKey = os.Getenv("IMMICH_API_KEY")
	immichUrl = os.Getenv("IMMICH_URL")

	log.SetLevel(log.DebugLevel)

	config.Load()
	log.Info("Config loaded", "config", config)
}

func main() {

	e := echo.New()

	e.HideBanner = true

	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e.Use(middleware.Recover())

	e.Static("/css", "public/css")

	e.GET("/", home)

	e.GET("/new", newImage)

	err := e.Start(":3000")
	if err != nil {
		log.Fatal(err)
	}

}
