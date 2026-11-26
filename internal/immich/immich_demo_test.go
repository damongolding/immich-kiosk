package immich

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/demo"
)

// demoConfig points demo mode at the fake server. demo.Login honours
// KIOSK_IMMICH_URL, so no request ever reaches the real demo instance.
func demoConfig(t *testing.T, srv *fakeImmich) config.Config {
	t.Helper()
	t.Setenv("KIOSK_IMMICH_URL", srv.URL)

	prevToken := demo.DemoToken
	demo.DemoToken = ""
	t.Cleanup(func() { demo.DemoToken = prevToken })

	cfg := testConfig(srv.URL)
	cfg.Kiosk.DemoMode = true
	return cfg
}

func TestImmichAPICall_DemoModeUsesBearerToken(t *testing.T) {
	srv := newFakeImmich(t,
		respondJSON(http.StatusOK, `{"accessToken":"demo-token"}`), // login
		respondJSON(http.StatusOK, `{"ok":true}`),                  // the API call
	)
	a := newTestAsset(t, demoConfig(t, srv))

	if _, _, _, err := a.immichAPICall(t.Context(), http.MethodGet, srv.URL+"/api/x", nil); err != nil {
		t.Fatalf("immichAPICall() error = %v", err)
	}

	reqs := srv.Requests()
	if len(reqs) != 2 {
		t.Fatalf("requests = %d, want 2 (login, then the call)", len(reqs))
	}
	if reqs[0].Path != "/api/auth/login" {
		t.Errorf("first request path = %q, want /api/auth/login", reqs[0].Path)
	}
	if got := reqs[1].Header.Get("Authorization"); got != "Bearer demo-token" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer demo-token")
	}
	if got := reqs[1].Header.Get("x-api-key"); got != "" {
		t.Errorf("x-api-key = %q, want none in demo mode", got)
	}
}

func TestImmichAPICall_DemoModeReloginOnExpiredToken(t *testing.T) {
	srv := newFakeImmich(t,
		respondJSON(http.StatusOK, `{"accessToken":"old-token"}`),  // login
		respondJSON(http.StatusUnauthorized, `{"message":"nope"}`), // call with old token
		respondJSON(http.StatusOK, `{"authStatus":false}`),         // validateToken
		respondJSON(http.StatusOK, `{"accessToken":"new-token"}`),  // re-login
		respondJSON(http.StatusOK, `{"ok":true}`),                  // retried call
	)
	a := newTestAsset(t, demoConfig(t, srv))

	body, _, _, err := a.immichAPICall(t.Context(), http.MethodGet, srv.URL+"/api/x", nil)
	if err != nil {
		t.Fatalf("immichAPICall() error = %v", err)
	}
	assertJSONEq(t, `{"ok":true}`, string(body))

	reqs := srv.Requests()
	if len(reqs) != 5 {
		t.Fatalf("requests = %d, want 5", len(reqs))
	}
	if reqs[2].Path != "/api/auth/validateToken" {
		t.Errorf("third request path = %q, want /api/auth/validateToken", reqs[2].Path)
	}
	if got := reqs[4].Header.Get("Authorization"); got != "Bearer new-token" {
		t.Errorf("retried call Authorization = %q, want %q", got, "Bearer new-token")
	}
}

// TestImmichAPICall_DemoLoginUsesCallerContext checks demo login honours the
// ctx passed to immichAPICall, not the context the Asset was created with.
func TestImmichAPICall_DemoLoginUsesCallerContext(t *testing.T) {
	srv := newFakeImmich(t, respondJSON(http.StatusOK, `{"accessToken":"demo-token"}`))
	cfg := demoConfig(t, srv)

	// The asset's own context is never cancelled...
	a := New(context.Background(), cfg)

	// ...but the caller's is.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, _, _, err := a.immichAPICall(ctx, http.MethodGet, srv.URL+"/api/x", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if got := srv.RequestCount(); got != 0 {
		t.Errorf("requests = %d, want 0: login should not run with a cancelled context", got)
	}
}
