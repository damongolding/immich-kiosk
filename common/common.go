// Package common provides shared types and utilities for the immich-kiosk application
package common

import (
	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/damongolding/immich-kiosk/utils"
)

var SharedSecret string

// RouteRequestData contains request metadata and configuration used across routes
type RouteRequestData struct {
	RequestConfig config.Config // Configuration for the current request
	DeviceID      string        // Unique identifier for the device making the request
	RequestID     string        // Unique identifier for this specific request
	ClientName    string        // Name of the client making the request
}

func init() {
	secret, err := utils.GenerateSharedSecret()
	if err != nil {
		log.Fatal("Failed to generate shared secret", "error", err)
	}
	SharedSecret = secret
}
