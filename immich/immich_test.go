package immich

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
)

// TestGetRandomImage testing if no images are found. Should retry 10 times
func TestGetRandomImage(t *testing.T) {

	log.SetLevel(log.DebugLevel)

	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		imgRes := make([]ImmichAsset, 5)

		for i := range 5 {
			imgRes[i] = ImmichAsset{
				Type: "VIDEO",
			}
		}

		out, _ := json.Marshal(imgRes)

		// Send response to be tested
		rw.Write(out)
	}))
	// Close the server when test finishes
	defer server.Close()

	c := config.New()
	c.ImmichUrl = server.URL
	c.ImmichApiKey = "123456"

	i := NewImage(c)

	err := i.GetRandomImage("TESTING")
	if err == nil {
		t.Error("A image was found")
		return
	}

	if err.Error() != "No images found" && i.Retries != 10 {
		t.Error(err)
	}

}
