package routes

import (
	"net/http"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/views"

	"github.com/labstack/echo/v4"
)

func Wall(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		return Render(c, http.StatusOK, views.Wall())
	}
}
