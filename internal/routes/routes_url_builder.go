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

func BuildURL(baseConfig *config.Config) echo.HandlerFunc {
	const maxURLLength = 2048

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
		formError := ""

		renderURL := kioskURL.String()
		if len(renderURL) > maxURLLength {
			renderURL = truncateURLQueries(renderURL, maxURLLength)
			formError = "This URL is longer than browsers allow. Kiosk has trimmed it, so some of your selected options may not be applied."
		}

		return Render(c, http.StatusOK, partials.URLResult(renderURL, formError))
	}
}

func truncateURLQueries(rawURL string, maxLength int) string {
	parts := strings.SplitN(rawURL, "?", 2)
	if len(parts) < 2 {
		return rawURL
	}

	base := parts[0]
	queryString := parts[1]

	if len(base) >= maxLength {
		return base
	}

	params := strings.Split(queryString, "&")
	result := base + "?" + params[0]

	for _, param := range params[1:] {
		candidate := result + "&" + param
		if len(candidate) > maxLength {
			break
		}
		result = candidate
	}

	return result
}

func URLBuilderPage(baseConfig *config.Config, com *common.Common, extended bool) echo.HandlerFunc {
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

		return Render(c, http.StatusOK, views.URLBuilder(viewData, urlData, extended))
	}
}
