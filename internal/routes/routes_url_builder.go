package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/templates/views"
	"github.com/google/go-querystring/query"
	"github.com/labstack/echo/v4"
)

// BuildURL returns an Echo handler that constructs a kiosk URL from the incoming request and renders it.
// 
// The handler determines the kiosk host from the X-Forwarded-Host header (falling back to the request Host)
// and the scheme from the request or the X-Forwarded-Proto header (only "http" or "https" are accepted).
// It parses and cleans form values (removing single empty values), binds them into a common.URLBuilderRequest,
// treats zero or negative durations as unset, encodes the request as query parameters, and injects
// baseConfig.Kiosk.Password as the "password" query parameter when present. The final URL is rendered
// using the UrlResult partial. The handler responds with appropriate HTTP errors for form parsing, binding,
// URL parsing, or query encoding failures.
func BuildURL(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		kioskHost := c.Request().Header.Get("X-Forwarded-Host")
		if kioskHost == "" {
			kioskHost = c.Request().Host
		}
		scheme := c.Scheme()
		if xf := c.Request().Header.Get("X-Forwarded-Proto"); xf != "" {
			s := strings.ToLower(xf)
			if s == "http" || s == "https" {
				scheme = s
			}
		}
		kioskURL, err := url.Parse(fmt.Sprintf("%s://%s", scheme, kioskHost))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "parsing url from request")
		}

		if err = c.Request().ParseForm(); err != nil {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid form")
		}

		// remove empty form values so optional fields parse as nil a.k.a config defaults
		for k, v := range c.Request().Form {
			if len(v) == 1 && v[0] == "" {
				delete(c.Request().Form, k)
			}
		}

		var req common.URLBuilderRequest
		if err = c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "parsing form")
		}

		// Treat invalid durations (zero or negative) as unset to preserve config defaults
		if req.Duration != nil && *req.Duration <= 0 {
			req.Duration = nil
		}

		queries, err := query.Values(req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("encoding query parameters: %v", err))
		}

		if baseConfig.Kiosk.Password != "" {
			queries.Add("password", baseConfig.Kiosk.Password)
		}

		kioskURL.RawQuery = queries.Encode()

		return Render(c, http.StatusOK, partials.UrlResult(kioskURL.String()))
	}
}

// URLBuilderPage returns an echo.HandlerFunc that prepares data for the URL builder page and renders the URL builder view.
// The handler initializes request context, fetches named people, albums, and tags for the current request and device, assembles view data, and renders the URL builder; any errors from initialization or data retrieval are propagated.
func URLBuilderPage(baseConfig *config.Config, com *common.Common) echo.HandlerFunc {
	return func(c echo.Context) error {

		requestData, err := InitializeRequestData(c, baseConfig)
		if err != nil {
			return err
		}

		if requestData == nil {
			log.Info("Refreshing clients")
			return nil
		}

		requestConfig := requestData.RequestConfig
		requestID := requestData.RequestID
		deviceID := requestData.DeviceID

		log.Debug(
			requestID,
			"method", c.Request().Method,
			"path", c.Request().URL.String(),
			"requestConfig", requestConfig.String(),
		)

		im := immich.New(com.Context(), requestConfig)

		ppl, pplErr := im.AllNamedPeople(requestID, deviceID)
		if pplErr != nil {
			return pplErr
		}

		albs, albErr := im.AllAlbums(requestID, deviceID)
		if albErr != nil {
			return albErr
		}

		tags, _, tagsErr := im.AllTags(requestID, deviceID)
		if tagsErr != nil {
			return tagsErr
		}

		viewData := common.ViewData{
			KioskVersion: KioskVersion,
			RequestID:    requestID,
			DeviceID:     deviceID,
			Config:       requestConfig,
		}

		urlData := common.URLViewData{
			People: ppl,
			Albums: albs,
			Tags:   tags,
		}

		return Render(c, http.StatusOK, views.URLBuilder(viewData, urlData))
	}
}