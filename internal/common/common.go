// Package common provides shared types and utilities for the immich-kiosk application
package common

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/utils"
)

var SharedSecret string
var SharedSecretInit sync.Once

// RouteRequestData contains request metadata and configuration used across routes
type RouteRequestData struct {
	RequestConfig config.Config // Configuration for the current request
	DeviceID      string        // Unique identifier for the device making the request
	RequestID     string        // Unique identifier for this specific request
	ClientName    string        // Name of the client making the request
}

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

func init() {
	if err := InitializeSecret(); err != nil {
		log.Fatal("Failed to initialize", "error", err)
	}
}

type ImageData struct {
	// ImmichImage immich asset data
	ImmichImage immich.ImmichAsset
	// ImageData image as base64 data
	ImageData string
	// ImageData blurred image as base64 data
	ImageBlurData string
	// Date image date
	ImageDate string
}

type ViewData struct {
	// KioskVersion the current build version of Kiosk
	KioskVersion string
	// DeviceID unique id for device
	DeviceID string
	// Images the images to display in view
	Images []ImageData
	// URL queries
	Queries url.Values
	// CustomCss
	CustomCss []byte
	// instance config
	config.Config
}
