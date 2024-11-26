package routes

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

	const (
		maxRedirects        = 10
		redirectCountHeader = "X-Redirect-Count"
	)

	return func(c echo.Context) error {

		redirectCount := c.Request().Header.Get(redirectCountHeader)
		count := 0
		if redirectCount != "" {
			var err error
			count, err = strconv.Atoi(redirectCount)
			if err != nil {
				count = 0
			}
		}

		// Check if maximum redirects exceeded
		if count >= maxRedirects {
			return echo.NewHTTPError(http.StatusTooManyRequests, "Too many redirects")
		}

		var r RedirectPath
		if err := c.Bind(&r); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid redirect request")
		}

		if r.RedirectName == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Redirect name is required")
		}

		if redirectItem, exists := baseConfig.Kiosk.RedirectsMap[r.RedirectName]; exists {
			if strings.EqualFold(redirectItem.Type, "internal") {

				parsedUrl, err := url.Parse(redirectItem.URL)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Invalid redirect URL")
				}

				for key, values := range parsedUrl.Query() {
					c.QueryParams().Set(key, values[0])
				}

				return Home(baseConfig)(c)

			}

			c.Response().Header().Set(redirectCountHeader, strconv.Itoa(count+1))

			return c.Redirect(http.StatusTemporaryRedirect, redirectItem.URL)
		}

		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}
