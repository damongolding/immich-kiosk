package routes

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/damongolding/immich-kiosk/internal/common"
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

// TestProcessViewImageData_SkipsVideosWithoutVideoManager covers an album that
// isn't shuffled (album_order: asc) with caching off, whose first asset is a
// video. Nothing advances past that video between retries, so without a video
// manager every retry used to pick it again and the request failed - even
// though the album has an image to show.
func TestProcessViewImageData_SkipsVideosWithoutVideoManager(t *testing.T) {
	withoutVideoManager(t)

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, image.NewRGBA(image.Rect(0, 0, 4, 3))); err != nil {
		t.Fatalf("encoding test PNG: %v", err)
	}

	const albumAssets = `{"assets":{"items":[` +
		`{"id":"video-1","type":"VIDEO","duration":5000,"exifInfo":{"exifImageWidth":4,"exifImageHeight":3}},` +
		`{"id":"image-1","type":"IMAGE","exifInfo":{"exifImageWidth":4,"exifImageHeight":3}}` +
		`],"nextPage":""}}`

	var (
		pathsMu sync.Mutex
		paths   []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathsMu.Lock()
		paths = append(paths, r.URL.Path)
		pathsMu.Unlock()

		switch {
		case r.URL.Path == "/api/search/metadata":
			// Deliberately ignores the request's type filter, like a stale
			// cache would: the code must not rely on Immich omitting videos.
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(albumAssets))
		case strings.HasPrefix(r.URL.Path, "/api/assets/image-1/"):
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngBuf.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	cfg := config.New()
	cfg.ImmichURL = srv.URL
	cfg.ImmichAPIKey = "test-api-key"
	cfg.Albums = []string{"album-1"}
	cfg.AlbumOrder = "asc"
	cfg.Kiosk.Cache = false
	cfg.ShowVideos = true

	ctxCopy := common.ContextCopy{
		RequestHeader:  http.Header{},
		ResponseHeader: http.Header{},
		URL:            url.URL{Path: "/asset/new"},
		Ctx:            t.Context(),
	}

	got, err := ProcessViewImageData(*cfg, ctxCopy, false)
	if err != nil {
		t.Fatalf("ProcessViewImageData() error = %v, want the album's image to be shown", err)
	}
	if got.ImmichAsset.ID != "image-1" {
		t.Errorf("shown asset = %q, want %q", got.ImmichAsset.ID, "image-1")
	}

	pathsMu.Lock()
	defer pathsMu.Unlock()
	for _, p := range paths {
		if strings.Contains(p, "video-1") {
			t.Errorf("requested %q: the unavailable video should never be processed", p)
		}
	}
}
