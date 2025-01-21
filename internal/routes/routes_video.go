package routes

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/labstack/echo/v4"
)

func NewVideo(baseConfig *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {

		videoID := c.Param("videoID")
		if videoID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Video ID is required")
		}

		vid, err := VideoManager.GetVideo(videoID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Video not found")
		}

		video, err := os.Open(vid.FilePath)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open video file")
		}
		defer video.Close()

		info, err := video.Stat()
		if err != nil {
			return err
		}

		fileSize := info.Size()

		rangeHeader := c.Request().Header.Get("Range")
		if rangeHeader == "" {
			// If no range requested, send entire file
			c.Response().Header().Set("Content-Type", vid.ImmichAsset.OriginalMimeType)
			c.Response().Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
			return c.Stream(http.StatusOK, vid.ImmichAsset.OriginalMimeType, video)
		}

		// Parse range header
		var start, end int64
		start = 0
		end = fileSize - 1

		if rangeHeader != "" {
			// Remove "bytes=" prefix
			rangeStr := strings.Replace(rangeHeader, "bytes=", "", 1)
			// Split the range into start-end
			parts := strings.Split(rangeStr, "-")

			if len(parts) == 2 {
				// Parse start range
				if parts[0] != "" {
					start, err = strconv.ParseInt(parts[0], 10, 64)
					if err != nil {
						start = 0
					}
				}

				// Parse end range
				if parts[1] != "" {
					end, err = strconv.ParseInt(parts[1], 10, 64)
					if err != nil {
						end = fileSize - 1
					}
				}
			}
		}

		// Validate ranges
		if start >= fileSize {
			return echo.NewHTTPError(416, "Requested range not satisfiable")
		}
		if end >= fileSize {
			end = fileSize - 1
		}

		// Calculate the chunk size
		chunkSize := end - start + 1

		// Seek to start position
		_, err = video.Seek(start, 0)
		if err != nil {
			return err
		}

		// Set headers for partial content
		c.Response().Header().Set("Content-Type", vid.ImmichAsset.OriginalMimeType)
		c.Response().Header().Set("Content-Length", strconv.FormatInt(chunkSize, 10))
		c.Response().Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		c.Response().Header().Set("Accept-Ranges", "bytes")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().Header().Set("Keep-Alive", "timeout=5, max=100")

		// Create a limited reader for the chunk
		chunk := io.LimitReader(video, chunkSize)

		// Stream the chunk
		return c.Stream(http.StatusPartialContent, vid.ImmichAsset.OriginalMimeType, chunk)
	}
}
