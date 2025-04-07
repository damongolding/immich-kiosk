// Package main is the entry point for the Immich Kiosk application.
//
// It sets up the web server, configures routes, and handles the main
// application logic for displaying and managing images in a kiosk mode.
// The package includes functionality for loading configurations, setting up
// middleware, and serving both dynamic content and static assets.
package main

import (
	"embed"
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/damongolding/immich-kiosk/internal/routes"
)

// version current build version number
var version string

//go:embed frontend/public
var public embed.FS

func init() {
	routes.KioskVersion = version
}

func main() {

	fmt.Println(kioskBanner)
	versionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#fe3366")).Render
	fmt.Print("Version ", versionStyle("0.17.2"), "\n\n")

	fmt.Println(versionStyle("⚠️ IMPORTANT NOTICE ⚠️\n"))

	fmt.Println("This Docker image has been deprecated and has been moved to a new location.")
	fmt.Println("")
	fmt.Println("NEW LOCATION:", lipgloss.NewStyle().Bold(true).Render("ghcr.io/damongolding/immich-kiosk:latest"))
	fmt.Println("")
	fmt.Println("Please update your compose file to use the new image location.")
	fmt.Println("This image will no longer receive updates and may be removed in the future.")
	fmt.Println("")
	fmt.Println("For more information, visit:", lipgloss.NewStyle().Bold(true).Render("https://github.com/damongolding/immich-kiosk?tab=readme-ov-file#docker-compose"))
	fmt.Println("")
	fmt.Println("Thank you for using Kiosk and for your understanding 🙇")

}
