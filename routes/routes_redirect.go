package routes

import (
	"net/http"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/labstack/echo/v4"
)

type RedirectPath struct {
	RedirectName string `param:"redirect"`
}

// Redirect returns an echo.HandlerFunc that handles URL redirections based on configured redirect paths.
// It takes a baseConfig parameter containing the application configuration including redirect mappings.
//
// If the requested redirect name exists in the RedirectsMap, it redirects to the mapped URL.
// Otherwise, it redirects to the root path "/".
//
// The function returns a temporary (307) redirect in both cases.
func Redirect(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r RedirectPath
		if err := c.Bind(&r); err != nil {
			return err
		}

		if url, exists := baseConfig.Kiosk.RedirectsMap[r.RedirectName]; exists {
			return c.Redirect(http.StatusTemporaryRedirect, url)
		}

		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}
