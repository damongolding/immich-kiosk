package routes

import (
	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
)

const defaultDateLayout = "02/01/2006"

var (
	KioskVersion string
)

type PersonOrAlbum struct {
	Type string
	ID   string
}

// This custom Render replaces Echo's echo.Context.Render() with templ's templ.Component.Render().
func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		log.Error("err rendering view", "err", err)
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}
