package immich

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestImmichAPICall_Success(t *testing.T) {
	srv := newFakeImmich(t, respondJSON(http.StatusOK, `{"ok":true}`))
	a := newTestAsset(t, testConfig(srv.URL))

	body, contentType, fromCache, err := a.immichAPICall(t.Context(), http.MethodGet, srv.URL+"/api/server/ping", nil)
	if err != nil {
		t.Fatalf("immichAPICall() error = %v", err)
	}

	assertJSONEq(t, `{"ok":true}`, string(body))
	if contentType != "application/json" {
		t.Errorf("contentType = %q, want %q", contentType, "application/json")
	}
	if fromCache {
		t.Error("fromCache = true, want false")
	}

	req := onlyRequest(t, srv)
	if req.Method != http.MethodGet {
		t.Errorf("method = %q, want %q", req.Method, http.MethodGet)
	}
	if req.Path != "/api/server/ping" {
		t.Errorf("path = %q, want %q", req.Path, "/api/server/ping")
	}
	if got := req.Header.Get("x-api-key"); got != "test-api-key" {
		t.Errorf("x-api-key = %q, want %q", got, "test-api-key")
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want %q", got, "application/json")
	}
	if got := req.Header.Get("Content-Type"); got != "" {
		t.Errorf("Content-Type = %q, want none on GET", got)
	}
}

func TestImmichAPICall_PostSendsBodyAndContentType(t *testing.T) {
	srv := newFakeImmich(t, respondJSON(http.StatusOK, `[]`))
	a := newTestAsset(t, testConfig(srv.URL))

	if _, _, _, err := a.immichAPICall(t.Context(), http.MethodPost, srv.URL+"/api/search/random", []byte(`{"size":5}`)); err != nil {
		t.Fatalf("immichAPICall() error = %v", err)
	}

	req := onlyRequest(t, srv)
	if req.Method != http.MethodPost {
		t.Errorf("method = %q, want %q", req.Method, http.MethodPost)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
	assertJSONEq(t, `{"size":5}`, string(req.Body))
}

func TestImmichAPICall_ExtraHeaders(t *testing.T) {
	srv := newFakeImmich(t, respondJSON(http.StatusOK, `{}`))
	a := newTestAsset(t, testConfig(srv.URL))

	_, _, _, err := a.immichAPICall(t.Context(), http.MethodGet, srv.URL+"/api/assets/x/video/playback", nil,
		map[string]string{"Accept": "application/octet-stream"})
	if err != nil {
		t.Fatalf("immichAPICall() error = %v", err)
	}

	req := onlyRequest(t, srv)
	if got := req.Header.Get("Accept"); got != "application/octet-stream" {
		t.Errorf("Accept = %q, want %q (extra headers should override defaults)", got, "application/octet-stream")
	}
}

func TestImmichAPICall_SelectedUserAPIKey(t *testing.T) {
	t.Run("uses the selected user's key", func(t *testing.T) {
		srv := newFakeImmich(t, respondJSON(http.StatusOK, `{}`))
		cfg := testConfig(srv.URL)
		cfg.ImmichUsersAPIKeys = map[string]string{"alice": "alice-key"}
		cfg.SelectedUser = "alice"
		a := newTestAsset(t, cfg)

		if _, _, _, err := a.immichAPICall(t.Context(), http.MethodGet, srv.URL+"/api/users/me", nil); err != nil {
			t.Fatalf("immichAPICall() error = %v", err)
		}

		req := onlyRequest(t, srv)
		if got := req.Header.Get("x-api-key"); got != "alice-key" {
			t.Errorf("x-api-key = %q, want %q", got, "alice-key")
		}
	})

	t.Run("unknown user fails without calling Immich", func(t *testing.T) {
		srv := newFakeImmich(t, respondJSON(http.StatusOK, `{}`))
		cfg := testConfig(srv.URL)
		cfg.SelectedUser = "bob"
		a := newTestAsset(t, cfg)

		_, _, _, err := a.immichAPICall(t.Context(), http.MethodGet, srv.URL+"/api/users/me", nil)
		if err == nil || !strings.Contains(err.Error(), "no API key found for user bob") {
			t.Fatalf("error = %v, want it to mention the missing API key for bob", err)
		}
		if got := srv.RequestCount(); got != 0 {
			t.Errorf("requests = %d, want 0", got)
		}
	})
}

func TestImmichAPICall_ErrorStatuses(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr string
	}{
		{"unauthorised", http.StatusUnauthorized, "received 401 (unauthorised)"},
		{"forbidden", http.StatusForbidden, "received 403 (unauthorised)"},
		{"not found", http.StatusNotFound, "HTTP 404"},
		{"server error", http.StatusInternalServerError, "HTTP 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noRetryBackoff(t)
			srv := newFakeImmich(t, respondJSON(tt.status, `{"message":"nope"}`))
			a := newTestAsset(t, testConfig(srv.URL))

			body, _, _, err := a.immichAPICall(t.Context(), http.MethodGet, srv.URL+"/api/x", nil)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
			}

			// The error body is returned so callers can decode Immich's message.
			assertJSONEq(t, `{"message":"nope"}`, string(body))

			// HTTP error statuses are returned straight away; only transport errors are retried.
			if got := srv.RequestCount(); got != 1 {
				t.Errorf("requests = %d, want 1", got)
			}
		})
	}
}

// Retry tests use POST: Go's transport may silently replay idempotent
// requests (e.g. GET) on a dropped connection, which would skew the counts.

func TestImmichAPICall_RetriesTransportErrors(t *testing.T) {
	t.Run("recovers after a dropped connection", func(t *testing.T) {
		noRetryBackoff(t)
		srv := newFakeImmich(t, dropConnection(t), respondJSON(http.StatusOK, `{"ok":true}`))
		a := newTestAsset(t, testConfig(srv.URL))

		body, _, _, err := a.immichAPICall(t.Context(), http.MethodPost, srv.URL+"/api/x", []byte(`{}`))
		if err != nil {
			t.Fatalf("immichAPICall() error = %v", err)
		}

		assertJSONEq(t, `{"ok":true}`, string(body))
		if got := srv.RequestCount(); got != 2 {
			t.Errorf("requests = %d, want 2", got)
		}
	})

	t.Run("gives up after maxAPIRetries", func(t *testing.T) {
		noRetryBackoff(t)
		srv := newFakeImmich(t, dropConnection(t))
		a := newTestAsset(t, testConfig(srv.URL))

		_, _, _, err := a.immichAPICall(t.Context(), http.MethodPost, srv.URL+"/api/x", []byte(`{}`))
		if err == nil || !strings.Contains(err.Error(), "max retries exceeded") {
			t.Fatalf("error = %v, want it to contain %q", err, "max retries exceeded")
		}
		if got := srv.RequestCount(); got != maxAPIRetries {
			t.Errorf("requests = %d, want %d", got, maxAPIRetries)
		}
	})
}

func TestImmichAPICall_CancelledDuringBackoff(t *testing.T) {
	// Signal when immichAPICall reaches the back-off, and return a delay long
	// enough that the test only finishes if cancellation is honoured.
	reachedBackoff := make(chan struct{})
	var once sync.Once
	orig := retryBackoff
	retryBackoff = func(int) time.Duration {
		once.Do(func() { close(reachedBackoff) })
		return time.Hour
	}
	t.Cleanup(func() { retryBackoff = orig })

	srv := newFakeImmich(t, dropConnection(t))
	a := newTestAsset(t, testConfig(srv.URL))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Cancel only once the first request has failed and the back-off has
	// started, so slow setup (e.g. under -race on CI) can't stop that request
	// being sent. The watchdog fails the test rather than hanging if the
	// back-off is never reached.
	watchdogFired := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-reachedBackoff:
		case <-time.After(30 * time.Second):
			close(watchdogFired)
		case <-stop:
			return
		}
		cancel()
	}()

	_, _, _, err := a.immichAPICall(ctx, http.MethodPost, srv.URL+"/api/x", []byte(`{}`))

	select {
	case <-watchdogFired:
		t.Fatal("watchdog fired: immichAPICall never reached the retry back-off")
	default:
	}

	if err == nil || !strings.Contains(err.Error(), "request cancelled during retry") {
		t.Fatalf("error = %v, want it to contain %q", err, "request cancelled during retry")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want it to wrap context.Canceled", err)
	}
	if got := srv.RequestCount(); got != 1 {
		t.Errorf("requests = %d, want 1 (should not retry once cancelled)", got)
	}
}

func TestWithImmichAPICache(t *testing.T) {
	// PersonStatistics is used as the cache shape: it's the smallest real
	// APIResponse type ({"assets": N}).
	type shape = PersonStatistics

	t.Run("disabled cache always calls Immich", func(t *testing.T) {
		flushCache(t)
		srv := newFakeImmich(t, respondJSON(http.StatusOK, `{"assets":1}`))
		cfg := testConfig(srv.URL)
		cfg.Kiosk.Cache = false
		a := newTestAsset(t, cfg)

		call := withImmichAPICache(a.immichAPICall, "req", "device-a", cfg, shape{})
		for i := range 2 {
			_, _, fromCache, err := call(t.Context(), http.MethodGet, srv.URL+"/api/thing", nil)
			if err != nil {
				t.Fatalf("call %d: error = %v", i, err)
			}
			if fromCache {
				t.Errorf("call %d: fromCache = true, want false", i)
			}
		}

		if got := srv.RequestCount(); got != 2 {
			t.Errorf("requests = %d, want 2", got)
		}
	})

	t.Run("second call is served from cache, trimmed to shape", func(t *testing.T) {
		flushCache(t)
		srv := newFakeImmich(t, respondJSON(http.StatusOK, `{"assets":1,"unused":"dropped"}`))
		cfg := testConfig(srv.URL)
		cfg.Kiosk.Cache = true
		a := newTestAsset(t, cfg)

		call := withImmichAPICache(a.immichAPICall, "req", "device-a", cfg, shape{})

		first, _, fromCache, err := call(t.Context(), http.MethodGet, srv.URL+"/api/thing", nil)
		if err != nil {
			t.Fatalf("first call: error = %v", err)
		}
		if fromCache {
			t.Error("first call: fromCache = true, want false")
		}
		// Fields not in the shape are dropped before caching.
		assertJSONEq(t, `{"assets":1}`, string(first))

		second, _, fromCache, err := call(t.Context(), http.MethodGet, srv.URL+"/api/thing", nil)
		if err != nil {
			t.Fatalf("second call: error = %v", err)
		}
		if !fromCache {
			t.Error("second call: fromCache = false, want true")
		}
		if !bytes.Equal(first, second) {
			t.Errorf("cached body = %s, want %s", second, first)
		}

		if got := srv.RequestCount(); got != 1 {
			t.Errorf("requests = %d, want 1", got)
		}
	})

	t.Run("entries are per device", func(t *testing.T) {
		flushCache(t)
		srv := newFakeImmich(t, respondJSON(http.StatusOK, `{"assets":1}`))
		cfg := testConfig(srv.URL)
		cfg.Kiosk.Cache = true
		a := newTestAsset(t, cfg)

		for _, device := range []string{"device-a", "device-b"} {
			call := withImmichAPICache(a.immichAPICall, "req", device, cfg, shape{})
			_, _, fromCache, err := call(t.Context(), http.MethodGet, srv.URL+"/api/thing", nil)
			if err != nil {
				t.Fatalf("%s: error = %v", device, err)
			}
			if fromCache {
				t.Errorf("%s: fromCache = true, want a miss", device)
			}
		}

		if got := srv.RequestCount(); got != 2 {
			t.Errorf("requests = %d, want 2", got)
		}
	})

	t.Run("errors are not cached", func(t *testing.T) {
		flushCache(t)
		srv := newFakeImmich(t,
			respondJSON(http.StatusInternalServerError, `{"message":"boom"}`),
			respondJSON(http.StatusOK, `{"assets":1}`),
		)
		cfg := testConfig(srv.URL)
		cfg.Kiosk.Cache = true
		a := newTestAsset(t, cfg)

		call := withImmichAPICache(a.immichAPICall, "req", "device-a", cfg, shape{})

		if _, _, _, err := call(t.Context(), http.MethodGet, srv.URL+"/api/thing", nil); err == nil {
			t.Fatal("first call: error = nil, want an error for HTTP 500")
		}

		body, _, fromCache, err := call(t.Context(), http.MethodGet, srv.URL+"/api/thing", nil)
		if err != nil {
			t.Fatalf("second call: error = %v", err)
		}
		if fromCache {
			t.Error("second call: fromCache = true, want false")
		}
		assertJSONEq(t, `{"assets":1}`, string(body))

		if got := srv.RequestCount(); got != 2 {
			t.Errorf("requests = %d, want 2", got)
		}
	})
}

// TestMe exercises a public Asset method end to end against the fake server.
func TestMe(t *testing.T) {
	flushCache(t)
	srv := newFakeImmich(t, respondJSON(http.StatusOK, `{"id":"user-1","name":"Test User","email":"test@example.com"}`))
	cfg := testConfig(srv.URL)
	cfg.Kiosk.Cache = true
	a := newTestAsset(t, cfg)

	for i := range 2 {
		me, err := a.Me("req", "device-a")
		if err != nil {
			t.Fatalf("call %d: Me() error = %v", i, err)
		}
		if me.ID != "user-1" {
			t.Errorf("call %d: ID = %q, want %q", i, me.ID, "user-1")
		}
		if me.Name != "Test User" {
			t.Errorf("call %d: Name = %q, want %q", i, me.Name, "Test User")
		}
	}

	// The second Me() should come from cache.
	req := onlyRequest(t, srv)
	if req.Path != "/api/users/me" {
		t.Errorf("path = %q, want %q", req.Path, "/api/users/me")
	}
}
