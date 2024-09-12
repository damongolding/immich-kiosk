package routes

import (
	"time"

	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"

	"github.com/damongolding/immich-kiosk/config"
)

var (
	KioskVersion string

	pageDataCache *cache.Cache
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
