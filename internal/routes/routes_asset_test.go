package routes

import (
	"bytes"
	"image/jpeg"
	"net/http"
	"strings"
	"testing"
)

func TestImage_ServesRandomAssetAsJPEG(t *testing.T) {
	rt := newRouteTest(t)
	rt.immich.Handle("POST /api/search/random", respondJSON(http.StatusOK, "["+imageAssetJSON("img-1", 4, 3)+"]"))
	rt.immich.Handle("GET /api/assets/img-1/thumbnail", respondPNG(t, 4, 3))

	rec, err := rt.serve(Image(rt.cfg, rt.com), http.MethodGet, "/image")
	if err != nil {
		t.Fatalf("Image() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("X-Immich-Asset-ID"); got != "img-1" {
		t.Errorf("X-Immich-Asset-ID = %q, want %q", got, "img-1")
	}
	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", got)
	}
	if _, decodeErr := jpeg.Decode(bytes.NewReader(rec.Body.Bytes())); decodeErr != nil {
		t.Errorf("body is not a valid JPEG: %v", decodeErr)
	}

	if n := len(rt.immich.RequestsTo(http.MethodPost, "/api/search/random")); n != 1 {
		t.Errorf("random search requests = %d, want 1", n)
	}
	thumbs := rt.immich.RequestsTo(http.MethodGet, "/api/assets/img-1/thumbnail")
	if len(thumbs) != 1 {
		t.Fatalf("thumbnail requests = %d, want 1", len(thumbs))
	}
	if got := thumbs[0].Header.Get("x-api-key"); got != "test-api-key" {
		t.Errorf("x-api-key = %q, want %q", got, "test-api-key")
	}
}

func TestNewAsset_RendersRandomImage(t *testing.T) {
	rt := newRouteTest(t)
	rt.immich.Handle("POST /api/search/random", respondJSON(http.StatusOK, "["+imageAssetJSON("img-1", 4, 3)+"]"))
	rt.immich.Handle("GET /api/assets/img-1/thumbnail", respondPNG(t, 4, 3))

	rec, err := rt.serve(NewAsset(rt.cfg, rt.com), http.MethodGet, "/asset/new")
	if err != nil {
		t.Fatalf("NewAsset() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", got)
	}
	if body := rec.Body.String(); !strings.Contains(body, "data:image/jpeg;base64,") {
		t.Errorf("rendered page has no inline JPEG image; body starts: %.200s", body)
	}
	if n := len(rt.immich.RequestsTo(http.MethodGet, "/api/assets/img-1/thumbnail")); n != 1 {
		t.Errorf("thumbnail requests = %d, want 1", n)
	}
}

func TestNewAsset_RendersErrorWhenImmichFails(t *testing.T) {
	rt := newRouteTest(t)
	rt.immich.Handle("POST /api/search/random", respondJSON(http.StatusInternalServerError, `{"message":"immich is down"}`))

	rec, err := rt.serve(NewAsset(rt.cfg, rt.com), http.MethodGet, "/asset/new")
	if err != nil {
		t.Fatalf("NewAsset() error = %v, want the error rendered as a page", err)
	}

	// Errors are rendered as an HTML partial (status 200) so htmx swaps it in.
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); !strings.Contains(body, "immich is down") {
		t.Errorf("error page does not include Immich's message; body starts: %.300s", body)
	}
	if n := len(rt.immich.RequestsTo(http.MethodGet, "/api/assets/img-1/thumbnail")); n != 0 {
		t.Errorf("thumbnail requests = %d, want 0", n)
	}
}
