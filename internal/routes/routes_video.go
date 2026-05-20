package routes

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"charm.land/log/v2"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/damongolding/immich-kiosk/internal/templates/partials"
	"github.com/labstack/echo/v5"
)

// NewVideo returns an HTTP handler for serving video files with support for HTTP range requests, caching headers, and partial content delivery.
// If demoMode is enabled, the handler responds with a plain text message indicating demo mode.
// Otherwise, it streams the requested video file, handling range requests for efficient streaming, and sets appropriate HTTP headers for caching and content negotiation.
// Returns 400 if the video ID is missing, 404 if the video is not found, 416 for invalid range requests, and 500 for internal errors.
func NewVideo(demoMode bool) echo.HandlerFunc {
	if demoMode {
		return func(c *echo.Context) error {
			return c.String(http.StatusOK, "Demo mode enabled")
		}
	}

	return func(c *echo.Context) error {
		videoID := c.Param("videoID")
		if videoID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Video ID is required")
		}

		vid, err := VideoManager.GetVideo(videoID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Video not found")
		}

		f, err := os.Open(vid.FilePath)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open video file")
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get video stats")
		}

		c.Response().Header().Set("ETag", vid.ImmichAsset.Checksum)
		c.Response().Header().Set("Cache-Control", "private, max-age=86400, immutable")
		c.Response().Header().Set("Content-Type", vid.ContentType)

		http.ServeContent(c.Response(), c.Request(), vid.FilePath, info.ModTime(), f)
		return nil
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
	rangeStart, rangeEnd, ok := strings.Cut(strings.Replace(rangeHeader, "bytes=", "", 1), "-")
	if !ok {
		return 0, 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid range format")
	}

	// Handle empty start range
	if rangeStart == "" {
		end, err = strconv.ParseInt(rangeEnd, 10, 64)
		if err != nil {
			return 0, 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid range end")
		}
		start = max(0, fileSize-end)
		end = fileSize - 1
		return start, end, statusCode, nil
	}

	// Parse start range
	start, err = strconv.ParseInt(rangeStart, 10, 64)
	if err != nil {
		return 0, 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid range start")
	}

	// Handle empty end range
	if rangeEnd == "" {
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(rangeEnd, 10, 64)
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
		return func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		}
	}

	return func(c *echo.Context) error {

		const maxPollCount = 5

		liveID := c.Param("liveID")
		if liveID == "" {
			log.Debug("Live photo ID is required")
			return c.NoContent(kiosk.StatusStopHTMXPolling)
		}

		pollCount := 0
		pollCountQuery := c.QueryParam("poll_count")
		if pollCountQuery != "" {
			var pollCountErr error
			pollCount, pollCountErr = strconv.Atoi(pollCountQuery)
			if pollCountErr != nil || pollCount < 0 {
				log.Warn("Invalid poll_count for live photo", "ID", liveID, "poll_count", pollCountQuery)
				return c.NoContent(kiosk.StatusStopHTMXPolling)
			}
		}

		if pollCount >= maxPollCount {
			log.Warn("Max retries reached for live photo", "ID", liveID)
			return c.NoContent(kiosk.StatusStopHTMXPolling)
		}

		video, err := VideoManager.GetVideo(liveID)
		if err != nil {
			return c.NoContent(http.StatusNoContent)
		}

		videoOrientation := kiosk.LandscapeOrientation
		if video.ImmichAsset.IsPortrait {
			videoOrientation = kiosk.PortraitOrientation
		}

		return Render(c, kiosk.StatusStopHTMXPolling, partials.LivePhoto(video.ID, video.ContentType, videoOrientation, password))
	}

}
