package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/damongolding/immich-kiosk/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestNewRawImage tests the NewRawImage handler function.
// It skips the test in CI environments, sets up a test HTTP request,
// loads the configuration, and asserts that the handler responds
// with a 200 OK status code.
func TestNewRawImage(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI environment")
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/image", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	req.Header.Set(echo.HeaderXRequestID, "TESTING")

	baseConfig := config.New()

	err := baseConfig.Load("../config.yaml")
	if err != nil {
		t.Error("Failed to load config", "err", err)
	}

	h := NewRawImage(baseConfig)

	// Assertions
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}
