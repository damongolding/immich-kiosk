package cache

import (
	"slices"
	"sync"
	"testing"

	"github.com/damongolding/immich-kiosk/internal/config"
)

// testViewConfig returns a fresh config per caller: assetToCache trims
// requestConfig.History in place, so goroutines must not share one.
func testViewConfig() *config.Config {
	c := config.New()
	c.Duration = 60
	return c
}

// cachedView returns the view queue stored for url/deviceID.
func cachedView[T any](t *testing.T, url, deviceID string) []T {
	t.Helper()
	data, found := Get(ViewCacheKey(url, deviceID))
	if !found {
		return nil
	}
	s, ok := data.([]T)
	if !ok {
		t.Fatalf("cached view has type %T, want %T", data, []T(nil))
	}
	return s
}

func TestAssetToCache_Order(t *testing.T) {
	Flush()
	t.Cleanup(Flush)
	const url, device = "/asset/new", "device-order"

	AssetToCache(1, testViewConfig(), device, url)
	AssetToCache(2, testViewConfig(), device, url)
	AssetToCacheWithPosition(0, testViewConfig(), device, url, PREPEND)

	if got, want := cachedView[int](t, url, device), []int{0, 1, 2}; !slices.Equal(got, want) {
		t.Errorf("queue = %v, want %v", got, want)
	}
}

// TestAssetToCache_ConcurrentWritesAreNotLost reproduces concurrent prefetches
// (and video-ready prepends) for the same device. Every write must survive.
func TestAssetToCache_ConcurrentWritesAreNotLost(t *testing.T) {
	Flush()
	t.Cleanup(Flush)
	const url, device = "/asset/new", "device-concurrent"
	const writers = 200

	var wg sync.WaitGroup
	for i := range writers {
		wg.Go(func() {
			pos := APPEND
			if i%2 == 0 {
				pos = PREPEND
			}
			AssetToCacheWithPosition(i, testViewConfig(), device, url, pos)
		})
	}
	wg.Wait()

	got := cachedView[int](t, url, device)
	if len(got) != writers {
		t.Fatalf("queue has %d items, want %d (concurrent writes were lost)", len(got), writers)
	}
	slices.Sort(got)
	for i, v := range got {
		if v != i {
			t.Fatalf("queue is missing or duplicating items: sorted[%d] = %d", i, v)
		}
	}
}

func TestPopView(t *testing.T) {
	t.Run("empty queue", func(t *testing.T) {
		Flush()
		t.Cleanup(Flush)

		if v, ok := PopView[int](testViewConfig(), "device-empty", "/asset/new"); ok {
			t.Errorf("PopView() = %v, true; want ok = false on an empty queue", v)
		}
	})

	t.Run("pops in order and removes the key when drained", func(t *testing.T) {
		Flush()
		t.Cleanup(Flush)
		const url, device = "/asset/new", "device-pop"

		for i := range 3 {
			AssetToCache(i, testViewConfig(), device, url)
		}

		for want := range 3 {
			got, ok := PopView[int](testViewConfig(), device, url)
			if !ok || got != want {
				t.Fatalf("PopView() = %v, %v; want %v, true", got, ok, want)
			}
		}

		if _, found := Get(ViewCacheKey(url, device)); found {
			t.Error("key still present after the queue was drained")
		}
		if v, ok := PopView[int](testViewConfig(), device, url); ok {
			t.Errorf("PopView() on drained queue = %v, true; want ok = false", v)
		}
	})

	t.Run("wrong type is dropped", func(t *testing.T) {
		Flush()
		t.Cleanup(Flush)
		const url, device = "/asset/new", "device-type"

		AssetToCache("not an int", testViewConfig(), device, url)

		if _, ok := PopView[int](testViewConfig(), device, url); ok {
			t.Error("PopView[int]() ok = true for a queue of strings")
		}
		if _, found := Get(ViewCacheKey(url, device)); found {
			t.Error("mistyped queue should be removed")
		}
	})
}

// TestPopView_ConcurrentWithWrites reproduces a request popping the next
// asset while prefetches append to the same queue: every item must be
// delivered exactly once - none lost, none shown twice.
func TestPopView_ConcurrentWithWrites(t *testing.T) {
	Flush()
	t.Cleanup(Flush)
	const url, device = "/asset/new", "device-pop-concurrent"
	const items = 200

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		popped []int
	)

	for i := range items {
		wg.Go(func() {
			AssetToCache(i, testViewConfig(), device, url)
		})
		wg.Go(func() {
			if v, ok := PopView[int](testViewConfig(), device, url); ok {
				mu.Lock()
				popped = append(popped, v)
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	all := slices.Concat(popped, cachedView[int](t, url, device))
	if len(all) != items {
		t.Fatalf("popped %d + remaining %d = %d items, want %d", len(popped), len(all)-len(popped), len(all), items)
	}
	slices.Sort(all)
	for i, v := range all {
		if v != i {
			t.Fatalf("items lost or delivered twice: sorted[%d] = %d", i, v)
		}
	}
}
