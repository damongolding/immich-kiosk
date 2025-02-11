// Package common provides shared types and utilities for the immich-kiosk application
package common

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/labstack/echo/v4"
)

var (
	initOnce sync.Once

	// shared context
	Context context.Context
	cancel  context.CancelFunc

	// SharedSecret stores the application-wide shared secret string
	SharedSecret string
)

// Initialize sets up the application context and shared secret.
// It ensures initialization occurs only once using sync.Once.
// Returns any errors that occurred during initialization.
func Initialize() error {
	var err error

	initOnce.Do(func() {
		err = initialize()
	})

	return err
}

// initialize performs the actual initialization work:
// - Creates cancellable context
// - Initializes shared secret
// - Sets up graceful shutdown handling
// Returns any errors that occurred during initialization.
func initialize() error {
	Context, cancel = context.WithCancel(context.Background())

	if err := InitializeSecret(); err != nil {
		log.Fatal("failed to initialize shared secret", "error", err)
	}

	// Handle graceful shutdown on interrupt signals
	go func() {
		sigChan := make(chan os.Signal, 2)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-sigChan:
			cancel()
		case <-Context.Done():
		}
		signal.Stop(sigChan)
	}()

	return nil
}

// RouteRequestData contains request metadata and configuration used across routes
type RouteRequestData struct {
	RequestConfig config.Config // Configuration for the current request
	DeviceID      string        // Unique identifier for the device making the request
	RequestID     string        // Unique identifier for this specific request
	ClientName    string        // Name of the client making the request
}

// InitializeSecret generates and sets the shared secret used for application security.
// The shared secret is used for authenticating and validating requests between components.
// Generation occurs only once through sync.Once synchronization to prevent duplicate secrets.
// The generated secret is stored in the SharedSecret global variable.
// Returns an error if the secret generation process fails.
func InitializeSecret() error {

	secret, err := utils.GenerateSharedSecret()
	if err != nil {
		return fmt.Errorf("failed to generate shared secret: %w", err)
	}
	SharedSecret = secret

	return nil
}

// ViewImageData contains the image data and metadata for displaying an image in the view
type ViewImageData struct {
	ImmichAsset   immich.ImmichAsset // ImmichAsset contains immich asset data
	ImageData     string             // ImageData contains the image as base64 data
	ImageBlurData string             // ImageBlurData contains the blurred image as base64 data
	ImageDate     string             // ImageDate contains the date of the image
	User          string             // User the user api key used
}

// ViewData contains all the data needed to render a view in the application
type ViewData struct {
	KioskVersion  string          // KioskVersion contains the current build version of Kiosk
	DeviceID      string          // DeviceID contains the unique identifier for the device
	Assets        []ViewImageData // Assets contains the collection of assets to display in view
	Queries       url.Values      // Queries contains the URL query parameters
	CustomCss     []byte          // CustomCss contains custom CSS styling as bytes
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
