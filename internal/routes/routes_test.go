package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/config"
	"github.com/damongolding/immich-kiosk/internal/utils"
	"github.com/damongolding/immich-kiosk/internal/video"
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

	prevVideoManager := VideoManager
	videoManager, videoManagerErr := video.New(t.Context())
	if videoManagerErr != nil {
		t.Fatalf("Failed to initialise video manager: %v", videoManagerErr)
	}

	videoManager.MaxAge = time.Minute
	VideoManager = videoManager

	t.Cleanup(func() {
		VideoManager = prevVideoManager
	})

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

func TestTruncateURLQueries(t *testing.T) {
	tests := []struct {
		name      string
		rawURL    string
		maxLength int
		want      string
	}{
		{
			name:      "No query string",
			rawURL:    "https://example.com/path",
			maxLength: 50,
			want:      "https://example.com/path",
		},
		{
			name:      "Base exceeds maxLength",
			rawURL:    "https://example.com/this/is/a/very/long/path",
			maxLength: 20,
			want:      "https://example.com/this/is/a/very/long/path",
		},
		{
			name:      "Single query fits",
			rawURL:    "https://example.com/path?a=1",
			maxLength: 50,
			want:      "https://example.com/path?a=1",
		},
		{
			name:      "Truncate after first query",
			rawURL:    "https://example.com/path?a=1&b=2&c=3",
			maxLength: len("https://example.com/path?a=1"),
			want:      "https://example.com/path?a=1",
		},
		{
			name:      "Include first two queries but not third",
			rawURL:    "https://example.com/path?a=1&b=2&c=3",
			maxLength: len("https://example.com/path?a=1&b=2"),
			want:      "https://example.com/path?a=1&b=2",
		},
		{
			name:      "Exactly full length allowed",
			rawURL:    "https://example.com/path?a=1&b=2",
			maxLength: len("https://example.com/path?a=1&b=2"),
			want:      "https://example.com/path?a=1&b=2",
		},
		{
			name:      "URL with trailing question mark",
			rawURL:    "https://example.com/path?",
			maxLength: 50,
			want:      "https://example.com/path?",
		},
		{
			name:      "URL with empty parameter value",
			rawURL:    "https://example.com/path?a=&b=2",
			maxLength: 50,
			want:      "https://example.com/path?a=&b=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateURLQueries(tt.rawURL, tt.maxLength)
			if got != tt.want {
				t.Errorf("truncateURLQueries() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShouldSkipBlur(t *testing.T) {
	tests := []struct {
		name       string
		config     config.Config
		shouldSkip bool
	}{
		{
			name:       "Blur off disables blur",
			config:     config.Config{BackgroundBlur: false},
			shouldSkip: true,
		},
		{
			name:       "Blur on, image fit cover, live photos enabled",
			config:     config.Config{BackgroundBlur: true, ImageFit: "cover", LivePhotos: true},
			shouldSkip: false,
		},
		{
			name:       "Blur on, image fit cover, live photos disabled",
			config:     config.Config{BackgroundBlur: true, ImageFit: "cover", LivePhotos: false},
			shouldSkip: true,
		},
		{
			name:       "Blur on, image fit cover, image effect zoom",
			config:     config.Config{BackgroundBlur: true, ImageFit: "cover", ImageEffect: "zoom"},
			shouldSkip: true,
		},
		{
			name:       "Blur on, image fit cover, image effect smart-zoom",
			config:     config.Config{BackgroundBlur: true, ImageFit: "cover", ImageEffect: "smart-zoom"},
			shouldSkip: true,
		},
		{
			name:       "Blur on, image fit contain, no effect",
			config:     config.Config{BackgroundBlur: true, ImageFit: "contain", ImageEffect: ""},
			shouldSkip: false,
		},
		{
			name:       "Blur on, image fit cover, effect none",
			config:     config.Config{BackgroundBlur: true, ImageFit: "cover", ImageEffect: "none"},
			shouldSkip: true,
		},
		{
			name:       "Blur on, image fit cover, effect zoom, live photos enabled",
			config:     config.Config{BackgroundBlur: true, ImageFit: "cover", ImageEffect: "zoom", LivePhotos: true},
			shouldSkip: false,
		},
		{
			name:       "Blur on, image fit contain, image effect smart-zoom",
			config:     config.Config{BackgroundBlur: true, ImageFit: "contain", ImageEffect: "smart-zoom"},
			shouldSkip: false,
		},
		{
			name:       "Blur on, layout splitview, image fit contain, image effect zoom, live photos disabled",
			config:     config.Config{BackgroundBlur: true, Layout: "splitview", ImageFit: "contain", ImageEffect: "zoom", LivePhotos: false},
			shouldSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipBlur(tt.config)
			if got != tt.shouldSkip {
				t.Errorf("shouldSkipBlur() = %v, want %v", got, tt.shouldSkip)
			}
		})
	}
}
