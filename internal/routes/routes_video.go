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

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/labstack/echo/v4"
)

func NewVideo(baseConfig *config.Config) echo.HandlerFunc {
	const bufferSize = 64 * 1024 // 64KB buffer

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
			if t, err := time.Parse(http.TimeFormat, ifModifiedSince); err == nil {
				if info.ModTime().Unix() <= t.Unix() {
					return c.NoContent(http.StatusNotModified)
				}
			}
		}

		c.Response().Header().Set("Content-Type", vid.ImmichAsset.OriginalMimeType)
		c.Response().Header().Set("Accept-Ranges", "bytes")

		rangeHeader := c.Request().Header.Get("Range")
		if rangeHeader == "" {
			c.Response().Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
			bufferedReader := bufio.NewReaderSize(video, bufferSize)
			return c.Stream(http.StatusOK, vid.ImmichAsset.OriginalMimeType, bufferedReader)
		}

		// Parse range
		start, end := int64(0), fileSize-1
		if rangeHeader != "" {
			rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
			parts := strings.Split(rangeStr, "-")
			if len(parts) == 2 {
				if parts[0] != "" {
					start, _ = strconv.ParseInt(parts[0], 10, 64)
				}
				if parts[1] != "" {
					end, _ = strconv.ParseInt(parts[1], 10, 64)
				}
			}
		}

		// Validate ranges
		if start >= fileSize || start > end {
			return echo.NewHTTPError(http.StatusRequestedRangeNotSatisfiable,
				"Requested range not satisfiable")
		}
		if end >= fileSize {
			end = fileSize - 1
		}

		// Seek and prepare chunk
		if _, err = video.Seek(start, io.SeekStart); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to seek video position")
		}

		chunkSize := end - start + 1
		c.Response().Header().Set("Content-Length", strconv.FormatInt(chunkSize, 10))
		c.Response().Header().Set("Content-Range",
			fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))

		chunk := io.LimitReader(video, chunkSize)

		bufferedReader := bufio.NewReaderSize(chunk, bufferSize)

		return c.Stream(http.StatusPartialContent, vid.ImmichAsset.OriginalMimeType,
			bufferedReader)
	}
}
