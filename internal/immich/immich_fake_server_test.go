package immich

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/config"
)

// TestMain prepares package-level state that the code under test relies on.
// The cache is a package global, so tests using it must not run in parallel.
func TestMain(m *testing.M) {
	cache.Initialize()
	os.Exit(m.Run())
}

// recordedRequest is a snapshot of a request received by fakeImmich.
type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Header http.Header
	Body   []byte
}

// fakeImmich is a stand-in Immich server for tests. Each incoming request is
// recorded, then answered by the next handler in the queue; once the queue is
// down to its last handler, that handler answers every remaining request.
type fakeImmich struct {
	*httptest.Server

	mu       sync.Mutex
	requests []recordedRequest
	handlers []http.HandlerFunc
}

// newFakeImmich starts a fake Immich server that answers with the given
// handlers in order. It is closed automatically when the test ends.
func newFakeImmich(t *testing.T, handlers ...http.HandlerFunc) *fakeImmich {
	t.Helper()

	if len(handlers) == 0 {
		t.Fatal("newFakeImmich: at least one handler is required")
	}

	f := &fakeImmich{handlers: handlers}
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.Close)

	return f
}

func (f *fakeImmich) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
		Header: r.Header.Clone(),
		Body:   body,
	})
	h := f.handlers[0]
	if len(f.handlers) > 1 {
		f.handlers = f.handlers[1:]
	}
	f.mu.Unlock()

	h(w, r)
}

// Requests returns a copy of every request received so far.
func (f *fakeImmich) Requests() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recordedRequest(nil), f.requests...)
}

// RequestCount returns how many requests have been received so far.
func (f *fakeImmich) RequestCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

// onlyRequest fails the test unless exactly one request was received, and
// returns it.
func onlyRequest(t *testing.T, f *fakeImmich) recordedRequest {
	t.Helper()
	reqs := f.Requests()
	if len(reqs) != 1 {
		t.Fatalf("received %d requests, want 1", len(reqs))
	}
	return reqs[0]
}

// assertJSONEq reports an error if want and got are not equivalent JSON
// documents (key order and whitespace are ignored).
func assertJSONEq(t *testing.T, want, got string) {
	t.Helper()

	var w, g any
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("assertJSONEq: invalid expected JSON %q: %v", want, err)
	}
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		t.Errorf("got invalid JSON %q: %v", got, err)
		return
	}
	if !reflect.DeepEqual(w, g) {
		t.Errorf("JSON mismatch:\n got: %s\nwant: %s", got, want)
	}
}

// respondJSON returns a handler that writes body as JSON with the given status.
func respondJSON(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}
}

// dropConnection returns a handler that closes the connection without
// responding, which the client sees as a transport error.
func dropConnection(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, _ *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("dropConnection: response writer does not support hijacking")
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Errorf("dropConnection: hijack: %v", err)
			return
		}
		_ = conn.Close()
	}
}

// testConfig returns a default config pointed at the fake server.
func testConfig(serverURL string) config.Config {
	c := config.New()
	c.ImmichURL = serverURL
	c.ImmichAPIKey = "test-api-key"
	return *c
}

// newTestAsset returns an Asset that talks to the fake server.
func newTestAsset(t *testing.T, cfg config.Config) Asset {
	t.Helper()
	return New(t.Context(), cfg)
}

// noRetryBackoff removes the delay between retries for the current test.
func noRetryBackoff(t *testing.T) {
	t.Helper()
	setRetryBackoff(t, 0)
}

// setRetryBackoff sets a fixed delay between retries for the current test.
func setRetryBackoff(t *testing.T, d time.Duration) {
	t.Helper()
	orig := retryBackoff
	retryBackoff = func(int) time.Duration { return d }
	t.Cleanup(func() { retryBackoff = orig })
}

// flushCache empties the shared cache before and after the current test.
func flushCache(t *testing.T) {
	t.Helper()
	cache.Flush()
	t.Cleanup(cache.Flush)
}
