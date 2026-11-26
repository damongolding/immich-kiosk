package routes

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
	"github.com/labstack/echo/v5"
)

// withoutVideoManager simulates video.New failing at startup, which leaves
// VideoManager nil. Before the fix, every path below panicked.
func withoutVideoManager(t *testing.T) {
	t.Helper()
	prev := VideoManager
	VideoManager = nil
	t.Cleanup(func() { VideoManager = prev })
}

func TestNewVideo_WithoutVideoManager(t *testing.T) {
	withoutVideoManager(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/video/abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "videoID", Value: "abc"}})

	err := NewVideo(false)(c)

	var httpErr *echo.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error = %v, want an *echo.HTTPError", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", httpErr.Code, http.StatusNotFound)
	}
}

func TestLivePhoto_WithoutVideoManager(t *testing.T) {
	withoutVideoManager(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/live/abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "liveID", Value: "abc"}})

	if err := LivePhoto(false, "")(c); err != nil {
		t.Fatalf("LivePhoto() error = %v", err)
	}
	// No video will ever arrive, so htmx should be told to stop polling.
	if rec.Code != kiosk.StatusStopHTMXPolling {
		t.Errorf("status = %d, want %d (stop polling)", rec.Code, kiosk.StatusStopHTMXPolling)
	}
}

func TestProcessVideo_WithoutVideoManager(t *testing.T) {
	withoutVideoManager(t)

	cfg := config.New()
	asset := immich.New(t.Context(), *cfg)
	asset.ID = "video-1"
	asset.Type = immich.VideoType

	_, err := processVideo(&asset, *cfg, "req", "device", "/", false)
	if !errors.Is(err, errVideoUnavailable) {
		t.Errorf("error = %v, want errVideoUnavailable", err)
	}
}

// TestProcessImage_LivePhotoWithoutVideoManager checks a live photo falls back
// to its still image, without fetching the video's info or downloading it.
func TestProcessImage_LivePhotoWithoutVideoManager(t *testing.T) {
	withoutVideoManager(t)

	var pngBuf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.White)
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatalf("encoding test PNG: %v", err)
	}

	var (
		pathsMu sync.Mutex
		paths   []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathsMu.Lock()
		paths = append(paths, r.URL.Path)
		pathsMu.Unlock()

		if !strings.HasPrefix(r.URL.Path, "/api/assets/photo-1/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBuf.Bytes())
	}))
	t.Cleanup(srv.Close)

	cfg := config.New()
	cfg.ImmichURL = srv.URL
	cfg.ImmichAPIKey = "test-api-key"
	cfg.LivePhotos = true

	asset := immich.New(t.Context(), *cfg)
	asset.ID = "photo-1"
	asset.Type = immich.ImageType
	asset.LivePhotoVideoID = "live-video-1"

	got, err := processImage(&asset, *cfg, "req", "device", false)
	if err != nil {
		t.Fatalf("processImage() error = %v", err)
	}
	if got == nil {
		t.Fatal("processImage() returned a nil image")
	}

	pathsMu.Lock()
	defer pathsMu.Unlock()
	for _, p := range paths {
		if strings.Contains(p, "live-video-1") {
			t.Errorf("requested %q: should not touch the live photo's video without a video manager", p)
		}
	}
	if len(paths) != 1 {
		t.Errorf("requests = %v, want exactly one (the still preview)", paths)
	}
}
