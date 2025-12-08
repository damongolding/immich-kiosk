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

// TestCacheExpirationScenario simulates the real-world scenario of the bug fix:
// - An album is cached with a 6-minute duration
// - Images are selected and cache is updated (using Set) multiple times
// - Verify the cache expiration is maintained properly
func TestCacheExpirationScenario(t *testing.T) {
	const duration = 360 // 6 minutes
	const expectedExpiration = (6 * time.Minute) + time.Minute

	type albumCache struct {
		Assets []string
	}

	key := "album_scenario_test"

	// Initial album with 5 images
	album := albumCache{
		Assets: []string{"img1", "img2", "img3", "img4", "img5"},
	}

	// Initial set
	Set(key, album, duration)

	startTime := time.Now()
	initialExpected := startTime.Add(expectedExpiration)

	_, initialExpiration, initialFound := kioskCache.GetWithExpiration(key)
	if !initialFound {
		t.Fatal("Expected album to be in cache after initial Set")
	}

	// Verify initial expiration
	diff := initialExpiration.Sub(initialExpected)
	const tolerance = 2 * time.Second
	if diff < -tolerance || diff > tolerance {
		t.Errorf("Initial: expected expiration within %v of %v, got %v (diff %v)",
			tolerance, initialExpected, initialExpiration, diff)
	}

	// Simulate showing images one by one
	for i := range 3 {
		// Small delay to simulate time passing
		time.Sleep(10 * time.Millisecond)

		// Get current album
		val, found := Get(key)
		if !found {
			t.Fatalf("Expected album in cache at iteration %d", i)
		}
		currentAlbum, ok := val.(albumCache)
		if !ok {
			t.Fatalf("Expected albumCache type, got %T", val)
		}

		// Remove first image (simulating asset selection)
		if len(currentAlbum.Assets) > 0 {
			currentAlbum.Assets = currentAlbum.Assets[1:]
		}

		// Update cache using Set (simulating the new fix)
		replaceTime := time.Now()
		expectedAfterReplace := replaceTime.Add(expectedExpiration)

		Set(key, currentAlbum, duration)

		// Verify expiration was properly extended
		_, expiration, found := kioskCache.GetWithExpiration(key)
		if !found {
			t.Fatalf("Expected album in cache after Set at iteration %d", i)
		}

		diff = expiration.Sub(expectedAfterReplace)
		if diff < -tolerance || diff > tolerance {
			t.Errorf("Iteration %d: expected expiration within %v of %v, got %v (diff %v)",
				i, tolerance, expectedAfterReplace, expiration, diff)
		}

		// The key assertion: expiration should be ~7 minutes from NOW,
		// not from the original set time
		timeSinceStart := time.Since(startTime)
		timeUntilExpiry := time.Until(expiration)

		// Expiry should be close to expectedExpiration (7 min), regardless of how long we've been running
		expiryDiff := timeUntilExpiry - expectedExpiration
		const expiryTolerance = 3 * time.Second
		if expiryDiff < -expiryTolerance || expiryDiff > expiryTolerance {
			t.Errorf("Iteration %d (after %v): expiration should be ~%v from now, but is %v (diff %v)",
				i, timeSinceStart, expectedExpiration, timeUntilExpiry, expiryDiff)
		}
	}

	// Final verification: ensure remaining assets are still in cache
	val, found := Get(key)
	if !found {
		t.Fatal("Expected album to still be in cache at end")
	}
	finalAlbum, ok := val.(albumCache)
	if !ok {
		t.Fatalf("Expected albumCache type, got %T", val)
	}

	expectedRemaining := 2 // Started with 5, removed 3
	if len(finalAlbum.Assets) != expectedRemaining {
		t.Errorf("Expected %d remaining assets, got %d", expectedRemaining, len(finalAlbum.Assets))
	}
}

// TestOldBugBehavior documents what the old buggy behavior was
// This test would FAIL with the old cache.Replace() implementation
// It now verifies that Set maintains the correct duration
func TestOldBugBehavior(t *testing.T) {
	t.Run("Old bug: Replace reset to default expiration", func(t *testing.T) {
		key := "bug_test"
		longDuration := 720 // 12 minutes

		// Set with long duration
		Set(key, "initial", longDuration)

		// Wait a moment
		time.Sleep(10 * time.Millisecond)

		// Update using Set - should maintain the long expiration calculation
		beforeReplace := time.Now()
		Set(key, "replaced", longDuration)

		_, expiration, found := kioskCache.GetWithExpiration(key)
		if !found {
			t.Fatal("Expected key in cache after replace")
		}

		timeUntilExpiry := time.Until(expiration)
		expectedExpiry := (12 * time.Minute) + time.Minute

		// With the FIX (using Set), this should be ~13 minutes
		const tolerance = 3 * time.Second
		diff := timeUntilExpiry - expectedExpiry
		if diff < -tolerance || diff > tolerance {
			t.Errorf("After Set at %v: expected expiry in ~%v, got %v (diff %v)",
				beforeReplace, expectedExpiry, timeUntilExpiry, diff)
		}

		if timeUntilExpiry < 6*time.Minute {
			t.Errorf("BUG DETECTED: Expiration was reset to default! Got %v, expected ~13 minutes", timeUntilExpiry)
		}
	})
}
