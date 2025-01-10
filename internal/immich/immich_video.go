package immich

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
)

var customTempVideoDir = filepath.Join(os.TempDir(), "immich-kiosk", "videos")

func VideoInit(ctx context.Context) {

	// Create custom temp directory if it doesn't exist
	err := os.MkdirAll(customTempVideoDir, 0755)
	if err != nil {
		fmt.Println("Error creating custom temp directory:", err)
		return
	}

	log.Debug("created video tmp dir at", "path", customTempVideoDir)

	go VideoCleanup(ctx)
}

func VideoCleanup(ctx context.Context) {
	// Run cleanup every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("Remove custom temp video directory")
			err := os.Remove(customTempVideoDir)
			if err != nil {
				log.Error("Error removing custom temp directory", "err", err)
			}
			return
		case <-ticker.C:
			cleanup()
		}
	}
}

func cleanup() {
	entries, err := os.ReadDir(customTempVideoDir)
	if err != nil {
		fmt.Printf("Error reading temp directory: %v\n", err)
		return
	}

	now := time.Now()
	maxAge := 5 * time.Minute

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		filePath := filepath.Join(customTempVideoDir, entry.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Error getting file info for %s: %v\n", filePath, err)
			continue
		}

		// Check if file is older than maxAge
		if now.Sub(fileInfo.ModTime()) > maxAge {
			err := os.Remove(filePath)
			if err != nil {
				fmt.Printf("Error deleting file %s: %v\n", filePath, err)
			} else {
				fmt.Printf("Deleted old file: %s\n", filePath)
			}
		}
	}
}
