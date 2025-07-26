package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestRawImage tests the NewRawImage handler function.
// It skips the test in CI environments, sets up a test HTTP request,
// loads the configuration, and asserts that the handler responds
// with a 200 OK status code.
func TestRawImage(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI environment")
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/image", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	req.Header.Set(echo.HeaderXRequestID, "TESTING")

	baseConfig := config.New()

	err := baseConfig.Load()
	if err != nil {
		t.Error("Failed to load config", "err", err)
	}

	cache.Initialize()

	h := Image(baseConfig, common.New())

	// Assertions
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestTrimHistory(t *testing.T) {
	testCases := []struct {
		name      string
		history   []string
		maxLength int
		expected  []string
	}{
		{
			name:      "Empty history",
			history:   []string{},
			maxLength: 5,
			expected:  []string{},
		},
		{
			name:      "History shorter than maxLength",
			history:   []string{"a", "b", "c"},
			maxLength: 5,
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "History equal to maxLength",
			history:   []string{"a", "b", "c", "d", "e"},
			maxLength: 5,
			expected:  []string{"a", "b", "c", "d", "e"},
		},
		{
			name:      "History longer than maxLength",
			history:   []string{"a", "b", "c", "d", "e", "f", "g"},
			maxLength: 5,
			expected:  []string{"c", "d", "e", "f", "g"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			history := tc.history
			utils.TrimHistory(&history, tc.maxLength)
			assert.Equal(t, tc.expected, history, "Trimmed history does not match expected result")
		})
	}
}
