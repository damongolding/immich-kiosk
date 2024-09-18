package immich

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/config"
	"github.com/stretchr/testify/assert"
)

// TestGetRandomImage testing if no images are found. Should retry 10 times
func TestRandomImage(t *testing.T) {

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

	i := NewImage(*c)

	err := i.RandomImage("TESTING", "TESTING", false)
	assert.NotNil(t, err, "Expected an error, but got nil")
	assert.Equal(t, "no images found", err.Error(), "Unexpected error message")
	assert.Equal(t, 10, i.Retries, "Expected 10 retries")
}

func TestArchiveLogic(t *testing.T) {

	tests := []struct {
		Type                  string
		IsTrashed             bool
		IsArchived            bool
		ArchivedWantedByUser  bool
		WantSimulatedContinue bool
	}{
		{
			Type:                  "IMAGE",
			IsTrashed:             false,
			IsArchived:            false,
			ArchivedWantedByUser:  false,
			WantSimulatedContinue: false,
		},
		{
			Type:                  "IMAGE",
			IsTrashed:             true,
			IsArchived:            false,
			ArchivedWantedByUser:  false,
			WantSimulatedContinue: true,
		},
		{
			Type:                  "IMAGE",
			IsTrashed:             false,
			IsArchived:            true,
			ArchivedWantedByUser:  false,
			WantSimulatedContinue: true,
		},
		{
			Type:                  "IMAGE",
			IsTrashed:             false,
			IsArchived:            true,
			ArchivedWantedByUser:  true,
			WantSimulatedContinue: false,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			simulatedContinueTriggered := false

			if test.Type != "IMAGE" || test.IsTrashed || (test.IsArchived && !test.ArchivedWantedByUser) {
				simulatedContinueTriggered = true
			}

			assert.Equal(t, test.WantSimulatedContinue, simulatedContinueTriggered, "Unexpected simulatedContinueTriggered value")
		})
	}

}
