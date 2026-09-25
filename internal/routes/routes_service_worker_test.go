package routes

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestRenderServiceWorkerFallbackHTML(t *testing.T) {
	sw, err := os.ReadFile("../../frontend/public/assets/js/sw.js.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	fallbackHTML := "<!doctype html><html><body>`\\${\"'<> &\u2028\u2029</body></html>"
	version := "0.44.0\"\nself.skipWaiting();"

	var output bytes.Buffer
	if err = renderServiceWorker(&output, sw, version, fallbackHTML); err != nil {
		t.Fatal(err)
	}

	worker := output.String()
	if strings.Contains(worker, "&lt;!doctype html&gt;") {
		t.Fatal("fallback HTML was HTML-escaped")
	}
	if got := workerString(t, worker, "FALLBACK_HTML"); got != fallbackHTML {
		t.Errorf("FALLBACK_HTML = %q, want %q", got, fallbackHTML)
	}
	if got := workerString(t, worker, "CACHE"); got != "immich-kiosk-"+version {
		t.Errorf("CACHE = %q, want %q", got, "immich-kiosk-"+version)
	}
}

func workerString(t *testing.T, worker, name string) string {
	t.Helper()
	pattern := regexp.MustCompile("(?m)^var " + name + " = (.+);$")
	match := pattern.FindStringSubmatch(worker)
	if match == nil {
		t.Fatalf("%s assignment not found", name)
	}

	var value string
	if err := json.Unmarshal([]byte(match[1]), &value); err != nil {
		t.Fatalf("%s is not a valid JavaScript string: %v", name, err)
	}
	return value
}
