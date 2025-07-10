// Package common provides shared types and utilities for the immich-kiosk application
package common

import (
	"context"
	"fmt"
	"image/color"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/labstack/echo/v4"
)

type Common struct {
	ctx    context.Context
	cancel context.CancelFunc
	secret string
}

func New() *Common {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Common{
		ctx:    ctx,
		cancel: cancel,
	}

	if err := c.initializeSecret(); err != nil {
		log.Fatal("failed to initialize shared secret", "error", err)
	}

	c.handleGracefulShutdown()
	return c
}

// initializeSecret generates and sets a secret token that is shared between application components
// this shared secret is used for secure communication and authentication between services
func (c *Common) initializeSecret() error {

	secret, err := utils.GenerateSharedSecret()
	if err != nil {
		return fmt.Errorf("failed to generate shared secret: %w", err)
	}
	c.secret = secret

	return nil
}

func (c *Common) handleGracefulShutdown() {
	go func() {
		sigChan := make(chan os.Signal, 2)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-sigChan:
			c.cancel()
		case <-c.ctx.Done():
		}
		signal.Stop(sigChan)
	}()
}

func (c *Common) Context() context.Context {
	return c.ctx
}

func (c *Common) Secret() string {
	return c.secret
}

// RouteRequestData contains request metadata and configuration used across routes
type RouteRequestData struct {
	RequestConfig config.Config // Configuration for the current request
	DeviceID      string        // Unique identifier for the device making the request
	RequestID     string        // Unique identifier for this specific request
	ClientName    string        // Name of the client making the request
}

// ViewImageData contains the image data and metadata for displaying an image in the view
type ViewImageData struct {
	ImmichAsset        immich.Asset // ImmichAsset contains immich asset data
	ImageData          string       // ImageData contains the image as base64 data
	ImageBlurData      string       // ImageBlurData contains the blurred image as base64 data
	ImageDominantColor color.RGBA   // ImageDominantColor contains the dominant color of the image
	ImageDate          string       // ImageDate contains the date of the image
	User               string       // User the user api key used
}

// ViewData contains all the data needed to render a view in the application
type ViewData struct {
	KioskVersion  string          // KioskVersion contains the current build version of Kiosk
	RequestID     string          // RequestID contains the unique identifier for the request
	DeviceID      string          // DeviceID contains the unique identifier for the device
	Assets        []ViewImageData // Assets contains the collection of assets to display in view
	Queries       url.Values      // Queries contains the URL query parameters
	CustomCSS     []byte          // CustomCSS contains custom CSS styling as bytes
	config.Config                 // Config contains the instance configuration
}

type ViewImageDataOptions struct {
	RelativeAssetWanted   bool
	RelativeAssetBucket   kiosk.Source
	RelativeAssetBucketID string
	ImageOrientation      immich.ImageOrientation
}

// ContextCopy stores a copy of key HTTP context information including URL and headers
type ContextCopy struct {
	URL            url.URL     // The request URL
	RequestHeader  http.Header // Headers from the incoming request
	ResponseHeader http.Header // Headers for the outgoing response
}

// CopyContext creates a copy of essential context data from an echo.Context
// This allows preserving context information without maintaining a reference to the original context
// Returns a ContextCopy containing the URL and header information
func CopyContext(c echo.Context) ContextCopy {

	ctxCopy := ContextCopy{
		URL:            *c.Request().URL,
		RequestHeader:  c.Request().Header.Clone(),
		ResponseHeader: c.Response().Header().Clone(),
	}

	return ctxCopy
}
