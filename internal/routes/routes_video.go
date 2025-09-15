package routes

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/labstack/echo/v4"
)

// NewVideo returns an HTTP handler for serving video files with support for HTTP range requests, caching headers, and partial content delivery.
// If demoMode is enabled, the handler responds with a plain text message indicating demo mode.
// Otherwise, it streams the requested video file, handling range requests for efficient streaming, and sets appropriate HTTP headers for caching and content negotiation.
// Returns 400 if the video ID is missing, 404 if the video is not found, 416 for invalid range requests, and 500 for internal errors.
func NewVideo(demoMode bool) echo.HandlerFunc {
	if demoMode {
		return func(c echo.Context) error {
			return c.String(http.StatusOK, "Demo mode enabled")
		}
	}

	const bufferSize = 1024 * 1024 // Increased to 1MB buffer

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
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get video stats")
		}

		requestID := utils.ColorizeRequestID(c.Response().Header().Get(echo.HeaderXRequestID))

		fileSize := info.Size()

		// Set headers
		c.Response().Header().Set("ETag", vid.ImmichAsset.Checksum)
		c.Response().Header().Set("Cache-Control", "private, max-age=86400, immutable")
		c.Response().Header().Set("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))
		c.Response().Header().Set("Expires", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))

		// Check if-none-match header
		if match := c.Request().Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, vid.ImmichAsset.Checksum) {
				return c.NoContent(http.StatusNotModified)
			}
		}

		// Check if-modified-since header
		if ifModifiedSince := c.Request().Header.Get("If-Modified-Since"); ifModifiedSince != "" {
			if t, tErr := time.Parse(http.TimeFormat, ifModifiedSince); tErr == nil {
				if info.ModTime().Unix() <= t.Unix() {
					return c.NoContent(http.StatusNotModified)
				}
			}
		}

		c.Response().Header().Set("Content-Type", vid.ContentType)
		c.Response().Header().Set("Accept-Ranges", "bytes")

		// Initialize start and end
		var start, end int64

		var statusCode int
		rangeHeader := c.Request().Header.Get("Range")
		start, end, statusCode, err = parseRangeHeader(rangeHeader, fileSize)
		if err != nil {
			log.Error(requestID, "err", err)
			return err
		}

		// Validate ranges more strictly
		if start < 0 || end < 0 || start >= fileSize {
			log.Error(requestID+" Invalid range", "start", start, "end", end, "fileSize", fileSize)
			return echo.NewHTTPError(http.StatusRequestedRangeNotSatisfiable,
				fmt.Sprintf("Invalid range: start=%d, end=%d, fileSize=%d", start, end, fileSize))
		}

		if start > end {
			log.Error(requestID+" Invalid range: start is greater than end", "start", start, "end", end)
			return echo.NewHTTPError(http.StatusRequestedRangeNotSatisfiable,
				fmt.Sprintf("Invalid range: start (%d) is greater than end (%d)", start, end))
		}

		if end >= fileSize {
			end = fileSize - 1
		}

		// Ensure chunk size isn't too large
		chunkSize := end - start + 1
		maxChunkSize := int64(10 * 1024 * 1024) // 10MB
		if chunkSize > maxChunkSize {
			end = start + maxChunkSize - 1
			chunkSize = maxChunkSize
		}

		// Add debug headers
		c.Response().Header().Set("X-Chunk-Size", strconv.FormatInt(chunkSize, 10))
		c.Response().Header().Set("X-Chunk-Start", strconv.FormatInt(start, 10))
		c.Response().Header().Set("X-Chunk-End", strconv.FormatInt(end, 10))

		if _, err = video.Seek(start, io.SeekStart); err != nil {
			log.Error(requestID + " Failed to seek video position")
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to seek video position")
		}

		c.Response().Header().Set("Content-Length", strconv.FormatInt(chunkSize, 10))
		c.Response().Header().Set("Content-Range",
			fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))

		// Use io.Copy instead of buffered reader for large chunks
		if chunkSize > bufferSize {
			return c.Stream(statusCode, vid.ContentType,
				io.NewSectionReader(video, start, chunkSize))
		}

		// Use buffered reader for smaller chunks
		bufferedReader := bufio.NewReaderSize(
			io.NewSectionReader(video, start, chunkSize),
			bufferSize,
		)

		return c.Stream(statusCode, vid.ContentType, bufferedReader)
	}
}

func parseRangeHeader(rangeHeader string, fileSize int64) (int64, int64, int, error) {
	var start, end int64
	var err error
	statusCode := http.StatusOK

	if rangeHeader == "" {
		return start, end, statusCode, nil
	}

	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid range format")
	}

	statusCode = http.StatusPartialContent
	ranges := strings.Split(strings.Replace(rangeHeader, "bytes=", "", 1), "-")
	if len(ranges) != 2 {
		return 0, 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid range format")
	}

	// Handle empty start range
	if ranges[0] == "" {
		end, err = strconv.ParseInt(ranges[1], 10, 64)
		if err != nil {
			return 0, 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid range end")
		}
		start = max(0, fileSize-end)
		end = fileSize - 1
		return start, end, statusCode, nil
	}

	// Parse start range
	start, err = strconv.ParseInt(ranges[0], 10, 64)
	if err != nil {
		return 0, 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid range start")
	}

	// Handle empty end range
	if ranges[1] == "" {
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(ranges[1], 10, 64)
		if err != nil {
			return 0, 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid range end")
		}
	}

	// Validate ranges
	if start >= fileSize {
		return 0, 0, 0, echo.NewHTTPError(http.StatusRequestedRangeNotSatisfiable, "Range start exceeds file size")
	}

	if end >= fileSize {
		end = fileSize - 1
	}

	if start > end {
		return 0, 0, 0, echo.NewHTTPError(http.StatusRequestedRangeNotSatisfiable, "Invalid range: start > end")
	}

	return start, end, statusCode, nil
}

func LivePhoto(demoMode bool, password string) echo.HandlerFunc {
	if demoMode {
		return func(c echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		}
	}

	return func(c echo.Context) error {

		liveID := c.Param("liveID")
		if liveID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Live photo ID is required")
		}

		video, err := VideoManager.GetVideo(liveID)
		if err != nil {
			return c.NoContent(http.StatusNoContent)
		}

		videoOrientation := kiosk.LandscapeOrientation
		if video.ImmichAsset.IsPortrait {
			videoOrientation = kiosk.PortraitOrientation
		}

		return Render(c, http.StatusOK, partials.LivePhoto(video.ID, videoOrientation, password))
	}

}
