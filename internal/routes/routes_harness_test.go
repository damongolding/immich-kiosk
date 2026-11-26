package routes

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/i18n"
	"github.com/labstack/echo/v5"
)

// This file is the test harness for the routes package: a fake Immich server
// that routes by method and path, plus helpers to build a config pointed at it
// and to run a handler against a request. Tests using it must not call
// t.Parallel(): the cache and VideoManager are package-level state.

// recordedRequest is a snapshot of a request received by fakeImmich.
type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Header http.Header
	Body   []byte
}

// fakeImmich is a stand-in Immich server. Register responses with Handle
// using http.ServeMux patterns (e.g. "POST /api/search/random"). Every
// request is recorded; unmatched ones get a 404 and are logged to the test.
type fakeImmich struct {
	*httptest.Server

	t   *testing.T
	mux *http.ServeMux

	mu       sync.Mutex
	requests []recordedRequest
}

// newFakeImmich starts a fake Immich server that is closed when the test ends.
func newFakeImmich(t *testing.T) *fakeImmich {
	t.Helper()

	f := &fakeImmich{t: t, mux: http.NewServeMux()}
	f.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("fakeImmich: no handler for %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	})
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.Close)

	return f
}

// Handle registers h for pattern, using http.ServeMux pattern syntax.
func (f *fakeImmich) Handle(pattern string, h http.HandlerFunc) {
	f.mux.HandleFunc(pattern, h)
}

func (f *fakeImmich) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
		Header: r.Header.Clone(),
		Body:   body,
	})
	f.mu.Unlock()

	f.mux.ServeHTTP(w, r)
}

// Requests returns a copy of every request received so far.
func (f *fakeImmich) Requests() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recordedRequest(nil), f.requests...)
}

// RequestsTo returns the requests whose method and path match exactly.
func (f *fakeImmich) RequestsTo(method, path string) []recordedRequest {
	var out []recordedRequest
	for _, r := range f.Requests() {
		if r.Method == method && r.Path == path {
			out = append(out, r)
		}
	}
	return out
}

// respondJSON returns a handler that writes body as JSON with the given status.
func respondJSON(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}
}

// respondPNG returns a handler that serves a small solid-colour PNG, for
// image preview and thumbnail endpoints.
func respondPNG(t *testing.T, width, height int) http.HandlerFunc {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := range width {
		for y := range height {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("respondPNG: encoding: %v", err)
	}

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(buf.Bytes())
	}
}

// imageAssetJSON returns the JSON for a minimal, valid Immich image asset.
func imageAssetJSON(id string, width, height int) string {
	return fmt.Sprintf(`{"id":%q,"type":"IMAGE","visibility":"timeline","originalMimeType":"image/jpeg",`+
		`"exifInfo":{"exifImageWidth":%d,"exifImageHeight":%d}}`, id, width, height)
}

var i18nOnce sync.Once

// initI18n loads the real locale files from the repo root, as main does from
// its embedded copy. Rendering templates panics without it.
func initI18n(t *testing.T) {
	t.Helper()
	i18nOnce.Do(func() {
		i18n.LocaleFS = os.DirFS("../..")
		if err := i18n.Init("en"); err != nil {
			t.Fatalf("i18n.Init: %v", err)
		}
	})
}

// routeTest bundles what a routes handler needs to run against fakeImmich.
type routeTest struct {
	t      *testing.T
	immich *fakeImmich
	cfg    *config.Config
	com    *common.Common
}

// newRouteTest returns a harness with a fresh fake Immich and a default config
// pointed at it. Prefetch is off so no background work outlives the test, and
// VideoManager is nil (videos unavailable) unless a test sets it.
func newRouteTest(t *testing.T) *routeTest {
	t.Helper()

	initI18n(t)
	cache.Initialize()
	t.Cleanup(cache.Flush)

	prevVideoManager := VideoManager
	VideoManager = nil
	t.Cleanup(func() { VideoManager = prevVideoManager })

	srv := newFakeImmich(t)

	cfg := config.New()
	cfg.ImmichURL = srv.URL
	cfg.ImmichAPIKey = "test-api-key"
	cfg.Kiosk.PreFetch = false

	return &routeTest{t: t, immich: srv, cfg: cfg, com: common.New()}
}

// serve runs h for a request to target and returns the recorder and the
// handler's error. It sets the headers routes read: X-Request-ID (on the
// response, where the request ID middleware puts it) and kiosk-device-id.
func (rt *routeTest) serve(h echo.HandlerFunc, method, target string) (*httptest.ResponseRecorder, error) {
	rt.t.Helper()

	e := echo.New()
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("kiosk-device-id", "test-device-"+strings.ReplaceAll(rt.t.Name(), "/", "-"))
	rec := httptest.NewRecorder()
	rec.Header().Set(echo.HeaderXRequestID, "test-request")

	err := h(e.NewContext(req, rec))
	return rec, err
}
