// Package views provides HTML templates and view-related functionality
// for rendering web pages in the Immich Kiosk application.
//
// It includes structures and utilities for managing page data,
// handling URL queries, and preparing content for browser display.
package views

import (
	"net/url"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/immich"
)

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
	// instance config
	config.Config
}

func quriesToJson(values url.Values) map[string]any {

	result := make(map[string]any)

	for key, value := range values {
		if len(value) == 1 {
			result[key] = value[0]
		} else {
			result[key] = value
		}
	}

	return result
}
