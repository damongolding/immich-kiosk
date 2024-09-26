// Package routes provides HTTP endpoint handlers for the Kiosk application.
//
// It includes functions for rendering pages, handling API requests,
// and managing caching of page data. This package is responsible for
// defining the web routes and their corresponding handler functions.
package routes

import (
	"net/http"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/views"
)

var (
	KioskVersion string

	pageDataCache      *cache.Cache
	pageDataCacheMutex sync.Mutex
)

type PersonOrAlbum struct {
	Type string
	ID   string
}

type RequestData struct {
	History []string `form:"history"`
	config.Config
}

func init() {
	// Setting up Immich api cache
	pageDataCache = cache.New(5*time.Minute, 10*time.Minute)
}

func RenderError(c echo.Context, err error, message string) error {
	log.Error(message, "err", err)
	return Render(c, http.StatusOK, views.Error(views.ErrorData{
		Title:   "Error " + message,
		Message: err.Error(),
	}))
}

// This custom Render replaces Echo's echo.Context.Render() with templ's templ.Component.Render().
func Render(ctx echo.Context, statusCode int, t templ.Component) error {

	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		log.Error("rendering view", "err", err)
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}
