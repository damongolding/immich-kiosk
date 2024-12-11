// Package common provides shared types and utilities for the immich-kiosk application
package common

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

var (
	Context context.Context
	cancel  context.CancelFunc
)

func init() {
	Context, cancel = context.WithCancel(context.Background())

	if err := InitializeSecret(); err != nil {
		log.Fatal("failed to initialize shared secret", "error", err)
	}

	// Handle graceful shutdown on interrupt signals
	go func() {
		sigChan := make(chan os.Signal, 5)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan
		cancel()
	}()
}

// SharedSecret stores the application-wide shared secret string
var SharedSecret string

// SharedSecretInit ensures SharedSecret is initialized only once
var SharedSecretInit sync.Once

// RouteRequestData contains request metadata and configuration used across routes
type RouteRequestData struct {
	RequestConfig config.Config // Configuration for the current request
	DeviceID      string        // Unique identifier for the device making the request
	RequestID     string        // Unique identifier for this specific request
	ClientName    string        // Name of the client making the request
}

// InitializeSecret generates and sets the shared secret for the application.
// It uses sync.Once to ensure the secret is only generated once.
// Returns an error if secret generation fails.
func InitializeSecret() error {
	var initErr error

	SharedSecretInit.Do(func() {
		secret, err := utils.GenerateSharedSecret()
		if err != nil {
			initErr = fmt.Errorf("failed to generate shared secret: %w", err)
			return
		}
		SharedSecret = secret
	})

	return initErr
}

// ViewImageData contains the image data and metadata for displaying an image in the view
type ViewImageData struct {
	ImmichImage   immich.ImmichAsset // ImmichImage contains immich asset data
	ImageData     string             // ImageData contains the image as base64 data
	ImageBlurData string             // ImageBlurData contains the blurred image as base64 data
	ImageDate     string             // ImageDate contains the date of the image
}

// ViewData contains all the data needed to render a view in the application
type ViewData struct {
	KioskVersion  string          // KioskVersion contains the current build version of Kiosk
	DeviceID      string          // DeviceID contains the unique identifier for the device
	Images        []ViewImageData // Images contains the collection of images to display in view
	Queries       url.Values      // Queries contains the URL query parameters
	CustomCss     []byte          // CustomCss contains custom CSS styling as bytes
	config.Config                 // Config contains the instance configuration
}
