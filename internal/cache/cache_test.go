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
	kioskCache.Flush()
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
			name:     "Less then default expiration",
			duration: 10,
			want:     defaultExpiration,
		},
		{
			name:     "More then default expiration",
			duration: 360, // 6 minutes
			want:     (6 * time.Minute) + time.Minute,
		},
		{
			name:     "30 minutes more then default expiration",
			duration: 1800, // 30 minutes
			want:     (30 * time.Minute) + time.Minute,
		},
		{
			name:     "Negative duration",
			duration: -10,
			want:     defaultExpiration,
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

			expirationStr := expiration.Format("2006-01-02 15:04:05")
			expectedStr := expected.Format("2006-01-02 15:04:05")

			if expirationStr != expectedStr {
				t.Errorf("Expected expiration '%v', got '%v'", expectedStr, expirationStr)
			}
		})
	}

}
