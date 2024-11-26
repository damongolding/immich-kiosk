package routes

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
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

	maxRedirects := 10

	return func(c echo.Context) error {

		redirectCount := c.Request().Header.Get("X-Redirect-Count")
		count := 0
		if redirectCount != "" {
			count, _ = strconv.Atoi(redirectCount)
		}

		// Check if maximum redirects exceeded
		if count >= maxRedirects {
			return echo.NewHTTPError(http.StatusTooManyRequests, "Too many redirects")
		}

		var r RedirectPath
		if err := c.Bind(&r); err != nil {
			return err
		}

		if redirectItem, exists := baseConfig.Kiosk.RedirectsMap[r.RedirectName]; exists {

			if strings.EqualFold(redirectItem.Type, "internal") {

				log.Info("INTERNAL")

				parsedUrl, err := url.Parse(redirectItem.URL)
				if err != nil {
					return err
				}

				params := parsedUrl.Query()
				for key, values := range params {
					for _, value := range values {
						c.QueryParams().Add(key, value)
					}
				}

				return Home(baseConfig)(c)

			}

			c.Response().Header().Set("X-Redirect-Count", strconv.Itoa(count+1))
			return c.Redirect(http.StatusTemporaryRedirect, redirectItem.URL)
		}

		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}
