// Package routes provides HTTP endpoint handlers for the Kiosk application.
//
// It includes functions for rendering pages, handling API requests,
// and managing caching of page data. This package is responsible for
// defining the web routes and their corresponding handler functions.
package routes

import (
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/video"
)

const (
	maxWeatherRetries      = 3
	maxProcessAssetRetries = 10
	maxRedirects           = 10
	redirectCountHeader    = "X-Redirect-Count"
)

var (
	KioskVersion string

	drawFacesOnImages string

	VideoManager *video.Manager

	mu sync.Mutex
)

type PersonOrAlbum struct {
	Type string
	ID   string
}

// requestMetadata holds information about the current request context
type requestMetadata struct {
	requestID string
	deviceID  string
	urlString string
}

func ShouldDrawFacesOnImages() bool {
	return drawFacesOnImages == "true"
}

// InitializeRequestData processes incoming request context and configuration to create RouteRequestData.
// It handles kiosk version checks, client configuration overrides, and request metadata.
//
// Parameters:
//   - c: Echo context containing the HTTP request and response data
//   - baseConfig: Base configuration to be used as template for request-specific config
//
// Returns:
//   - *common.RouteRequestData: Processed request data and configuration
//   - error: Any errors encountered during initialization
func InitializeRequestData(c echo.Context, baseConfig *config.Config) (*common.RouteRequestData, error) {

	kioskDeviceVersion := c.Request().Header.Get("kiosk-version")
	deviceID := c.Request().Header.Get("kiosk-device-id")
	requestID := utils.ColorizeRequestID(c.Response().Header().Get(echo.HeaderXRequestID))
	clientName := c.QueryParams().Get("client")
	if clientName == "" {
		clientName = c.FormValue("client")
	}

	// create a copy of the global config to use with this request
	requestConfig := *baseConfig

	// If kiosk version on client and server do not match refresh client.
	if kioskDeviceVersion != "" && KioskVersion != kioskDeviceVersion {
		c.Response().Header().Set("HX-Refresh", "true")
		return nil, c.NoContent(http.StatusNoContent)
	}

	queryParams := c.QueryParams()
	formParam, err := c.FormParams()
	if err != nil {
		log.Error("initialise request data", "error", err, "path", c.Request().URL.Path)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}

	queries := utils.MergeQueries(queryParams, formParam)

	err = requestConfig.ConfigWithOverrides(queries, c)
	if err != nil {
		log.Error("initialise request data", "error", err, "path", c.Request().URL.Path)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to process request")
	}

	return &common.RouteRequestData{
		RequestConfig: requestConfig,
		DeviceID:      deviceID,
		RequestID:     requestID,
		ClientName:    clientName,
	}, nil
}

func RenderError(c echo.Context, err error, message string, refresh int) error {
	log.Error(message, "err", err)

	retry := refresh > 5

	return Render(c, http.StatusOK, partials.Error(partials.ErrorData{
		Title:   "Error " + message,
		Message: err.Error(),
		Retry:   retry,
	}))
}

func RenderUnauthorized(c echo.Context) error {
	return Render(c, http.StatusUnauthorized, partials.Unauthorized())
}

func RenderMessage(c echo.Context, title, message string) error {
	return Render(c, http.StatusOK, partials.Message(partials.MessageData{
		Title:   title,
		Message: message,
	}))
}

// This custom Render replaces Echo's echo.Context.Render() with templ's templ.Component.Render().
func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	// Set content type manually
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	ctx.Response().WriteHeader(statusCode)

	// Stream the rendered HTML directly
	if err := t.Render(ctx.Request().Context(), ctx.Response().Writer); err != nil {
		log.Warn("rendering view", "err", err)
		return err
	}

	return nil
}
