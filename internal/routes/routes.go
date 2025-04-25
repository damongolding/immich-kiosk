// Package routes provides HTTP endpoint handlers for the Kiosk application.
//
// It includes functions for rendering pages, handling API requests,
// and managing caching of page data. This package is responsible for
// defining the web routes and their corresponding handler functions.
package routes

import (
	"context"
	"net/http"
	"os"
	"sync/atomic"

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
	maxProcessAssetRetries = 3
	maxRedirects           = 10
	redirectCountHeader    = "X-Redirect-Count"
)

var (
	KioskVersion string

	drawFacesOnImages string

	VideoManager *video.Manager
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

func RenderError(c echo.Context, err error, message string) error {
	log.Error(message, "err", err)
	return Render(c, http.StatusOK, partials.Error(partials.ErrorData{
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

func SaveOfflineAsset(ctx context.Context, filename string, t templ.Component, maxOfflineSize int64, offlineSize *atomic.Int64) error {

	if offlineSize.Load() >= maxOfflineSize {
		log.Debug("SaveOfflineAsset: max storage size reached", "offlineSize", offlineSize.Load(), "maxOfflineSize", maxOfflineSize)
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Error("creating file", "err", err)
		return err
	}
	defer file.Close()

	if err := t.Render(ctx, file); err != nil {
		log.Error("saving view", "err", err)
		return err
	}

	s, statErr := file.Stat()
	if statErr != nil {
		log.Error("getting file size", "err", statErr)
		return err
	}

	offlineSize.Add(s.Size())

	return nil
}
