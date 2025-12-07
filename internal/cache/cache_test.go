package cache

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	Initialize()
	code := m.Run()
	Flush()
	os.Exit(code)
}

func TestCacheSet(t *testing.T) {

	tests := []struct {
		name     string
		duration int
		want     time.Duration
	}{
		{
			name:     "Zero duration",
			duration: 0,
			want:     defaultExpiration,
		},
		{
			name:     "Less than default expiration",
			duration: 10,
			want:     defaultExpiration,
		},
		{
			name:     "More than default expiration",
			duration: 360, // 6 minutes
			want:     (6 * time.Minute) + time.Minute,
		},
		{
			name:     "30 minutes. More than default expiration",
			duration: 1800, // 30 minutes
			want:     (30 * time.Minute) + time.Minute,
		},
		{
			name:     "Negative duration",
			duration: -10,
			want:     defaultExpiration,
		},
		{
			name:     "Exactly default expiration",
			duration: 300,
			want:     defaultExpiration + time.Minute,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			key := fmt.Sprintf("test_%d", i)

			expected := time.Now().Add(tt.want)

			Set(key, key, tt.duration)
			_, expiration, found := kioskCache.GetWithExpiration(key)
			if !found {
				t.Errorf("Expected key '%s' to be found in cache", key)
			}

			diff := expiration.Sub(expected)
			const tolerance = 2 * time.Second
			if diff < -tolerance || diff > tolerance {
				t.Errorf("expected expiration within %v of %v, got %v (diff %v)", tolerance, expected, expiration, diff)
			}

		})
	}

}

func TestCacheReplace(t *testing.T) {
	tests := []struct {
		name           string
		setupDuration  int
		replaceAfter   time.Duration
		expectPreserve bool
	}{
		{
			name:           "Preserves expiration for duration > 5 minutes",
			setupDuration:  360, // 6 minutes
			replaceAfter:   1 * time.Second,
			expectPreserve: true,
		},
		{
			name:           "Preserves expiration for default duration",
			setupDuration:  60, // 1 minute (uses default 5 min)
			replaceAfter:   1 * time.Second,
			expectPreserve: true,
		},
		{
			name:           "Handles expired items gracefully",
			setupDuration:  0,
			replaceAfter:   0,
			expectPreserve: false, // Will use default expiration
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := fmt.Sprintf("replace_test_%d", i)

			// Set initial value with specific duration
			Set(key, "initial_value", tt.setupDuration)

			// Get the original expiration time
			_, originalExpiration, found := kioskCache.GetWithExpiration(key)
			if !found {
				t.Fatalf("Expected key '%s' to be found in cache after Set", key)
			}

			// Wait a bit if needed
			if tt.replaceAfter > 0 {
				time.Sleep(tt.replaceAfter)
			}

			// Replace the value
			err := Replace(key, "replaced_value")
			if err != nil {
				t.Fatalf("Replace failed: %v", err)
			}

			// Get the new expiration time
			value, newExpiration, found := kioskCache.GetWithExpiration(key)
			if !found {
				t.Fatalf("Expected key '%s' to be found in cache after Replace", key)
			}

			// Verify the value was replaced
			if value != "replaced_value" {
				t.Errorf("Expected value 'replaced_value', got '%v'", value)
			}

			if tt.expectPreserve {
				// Verify expiration time is preserved (within tolerance)
				diff := newExpiration.Sub(originalExpiration)
				const tolerance = 2 * time.Second
				if diff < -tolerance || diff > tolerance {
					t.Errorf("Expected expiration to be preserved. Original: %v, New: %v, Diff: %v",
						originalExpiration, newExpiration, diff)
				}
			}
		})
	}
}

func TestCacheReplaceNonExistent(t *testing.T) {
	err := Replace("non_existent_key", "value")
	if err == nil {
		t.Error("Expected error when replacing non-existent key, got nil")
	}
	if err.Error() != "key not found: non_existent_key" {
		t.Errorf("Expected 'key not found' error, got: %v", err)
	}
}
