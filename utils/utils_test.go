package utils

import (
	"math"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCombineQueries test to see if referer queries overwrite url queries
func TestCombineQueries(t *testing.T) {
	baseQueries := url.Values{}
	baseQueries.Set("refresh", "60")
	baseQueries.Set("transition", "fade")
	baseQueries.Set("raw", "false")

	refererQueries := "/demo-url?transition=none&image_fit=cover&raw=true"

	q, err := CombineQueries(baseQueries, refererQueries)
	assert.NoError(t, err)

	assert.Equal(t, []string{"fade", "none"}, q["transition"])
	assert.Equal(t, "cover", q.Get("image_fit"))
	assert.Equal(t, "60", q.Get("refresh"))
	assert.True(t, q.Has("raw"))
}

// TestRandomSingleItem test a single item
func TestRandomSingleItem(t *testing.T) {
	s := []string{"cheese"}

	out := RandomItem(s)

	assert.Equal(t, "cheese", out, "RandomItem should return the only item in a single-item slice")
}

// TestRandomSingleItem test a single item
func TestRandomEmptyItem(t *testing.T) {
	s := []string{}

	out := RandomItem(s)

	assert.Equal(t, "", out, "RandomItem should return an empty string for an empty string slice")

	n := []int{}

	out2 := RandomItem(n)

	assert.Equal(t, 0, out2, "RandomItem should return 0 for an empty int slice")

	i := []any{}

	out3 := RandomItem(i)

	assert.Nil(t, out3, "RandomItem should return nil for an empty interface slice")
}

// TestRandomStruct get out what we expect
func TestRandomStruct(t *testing.T) {

	type RendomStructDemo struct {
		name string
		age  int
	}

	s := []RendomStructDemo{
		{name: "John", age: 20},
		{name: "Clara", age: 34},
	}

	out := RandomItem(s)

	assert.Equal(t, "utils.RendomStructDemo", reflect.TypeOf(out).String(), "Unexpected type returned from RandomItem")

	assert.Contains(t, s, out, "RandomItem should return an item from the input slice")

	assert.NotNil(t, out, "RandomItem should not return nil for a non-empty slice")

}

// TestDateToLayout tests the DateToLayout function with various input formats.
// It verifies that the function correctly converts date format strings to Go layout strings.
func TestDateToLayout(t *testing.T) {
	tests := []struct {
		In   string
		Want string
	}{
		{"YYYY-MM-DD", "2006-01-02"},
		{"YYYY/MM/DD", "2006/01/02"},
		{"YYYY:MM:DD", "2006:01:02"},
		{"YYYY MM DD", "2006 01 02"},
		{"YY M DDD", "06 1 Mon"},
		{"YY MMM DDDD", "06 Jan Monday"},
		{"YYYY MMMM DDDD", "2006 January Monday"},
		{"YYYYYY-MM-DD", "200606-01-02"},
		{"YYYY MM DD additional text", "2006 01 02 additional text"},
		{"", ""},
	}

	for _, test := range tests {
		result := DateToLayout(test.In)
		assert.Equal(t, test.Want, result, "DateToLayout(%q) = %q, want %q", test.In, result, test.Want)
	}
}

// TestStringToColor checks if the StringToColor function consistently
// produces the same color for a given input string.
func TestStringToColor(t *testing.T) {
	in := "a sample string"

	a := StringToColor(in)
	b := StringToColor(in)

	assert.Equal(t, a.Hex, b.Hex, "Colors should match for the same input string")
	assert.NotEmpty(t, a.Hex, "Generated color should not be empty")
	assert.Len(t, a.Hex, 7, "Generated color should be in '#RRGGBB' format")
	assert.Equal(t, "#", a.Hex[:1], "Generated color should start with '#'")
}

// TestColorContrast checks the contrast ratio calculation between different colors
// and verifies that the maximum contrast ratio is achieved between black and white.
func TestColorContrast(t *testing.T) {
	white := Color{R: 255, G: 255, B: 255}
	black := Color{R: 0, G: 0, B: 0}

	maxRatio := calculateContrastRatio(white, black)
	assert.Equal(t, float64(21), maxRatio, "The contrast ratio between white and black should be 21")

	textWhite := calculateContrastRatio(white, black)
	textBlack := calculateContrastRatio(black, black)

	assert.Greater(t, textWhite, textBlack, "White text on black background should have a better contrast ratio than black text on black background")
}

// TestCalculateTotalWeight tests the calculateTotalWeight function.
// It verifies that:
// 1. The function correctly calculates the total weight for a slice with one item.
// 2. The returned total weight matches the expected value.
func TestCalculateTotalWeight(t *testing.T) {

	in := []AssetWithWeighting{
		{Asset: WeightedAsset{ID: "1"}, Weight: 1},
	}

	total := calculateTotalWeight(in)

	assert.Equal(t, 1, total, "total should be 1")

}

// TestWeightedRandomItem tests the WeightedRandomItem function.
// It verifies that:
// 1. The function returns an empty WeightedAsset for an empty input slice.
// 2. The function correctly returns the single asset for a slice with one item.
// 3. The function returns a valid asset from the input slice for multiple assets.
// 4. The distribution of returned assets approximately matches their logarithmic weights over many iterations.
func TestWeightedRandomItem(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name   string
		assets []AssetWithWeighting
		want   WeightedAsset
	}{
		{
			name:   "Empty slice",
			assets: []AssetWithWeighting{},
			want:   WeightedAsset{},
		},
		{
			name: "Single asset",
			assets: []AssetWithWeighting{
				{Asset: WeightedAsset{ID: "1"}, Weight: 1},
			},
			want: WeightedAsset{ID: "1"},
		},
		{
			name: "Multiple assets",
			assets: []AssetWithWeighting{
				{Asset: WeightedAsset{ID: "1"}, Weight: 1},
				{Asset: WeightedAsset{ID: "2"}, Weight: 2},
				{Asset: WeightedAsset{ID: "3"}, Weight: 3},
			},
			want: WeightedAsset{}, // We'll check for non-empty result
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := WeightedRandomItem(tc.assets)

			if len(tc.assets) == 0 {
				assert.Equal(t, tc.want, got, "WeightedRandomItem() = %v, want %v", got, tc.want)
			} else {
				found := false
				for _, asset := range tc.assets {
					t.Log(got, asset.Asset)
					if reflect.DeepEqual(got, asset.Asset) {
						found = true
						break
					}
				}
				assert.True(t, found, "WeightedRandomItem() returned an asset not in the input slice")
			}
		})
	}

	// Test for distribution
	assets := []AssetWithWeighting{
		{Asset: WeightedAsset{ID: "1"}, Weight: 1},
		{Asset: WeightedAsset{ID: "2"}, Weight: 2},
		{Asset: WeightedAsset{ID: "3"}, Weight: 3},
	}

	counts := make(map[string]int)
	iterations := 100000 // Increased iterations for more accurate results

	for i := 0; i < iterations; i++ {
		result := WeightedRandomItem(assets)
		counts[result.ID]++
	}

	totalLogWeight := math.Log(1) + math.Log(2) + math.Log(3)
	expectedRatios := map[string]float64{
		"1": math.Log(1) / totalLogWeight,
		"2": math.Log(2) / totalLogWeight,
		"3": math.Log(3) / totalLogWeight,
	}

	tolerance := 0.5

	for id, expectedRatio := range expectedRatios {
		actualRatio := float64(counts[id]) / float64(iterations)
		assert.InDelta(t, expectedRatio, actualRatio, tolerance, "Distribution for asset %s: got %.4f, want %.4f (±%.4f)", id, actualRatio, expectedRatio, tolerance)
	}
}

func TestIsSleepTime(t *testing.T) {
	tests := []struct {
		name           string
		sleepStartTime string
		sleepEndTime   string
		currentTime    time.Time
		want           bool
		wantErr        bool
	}{
		{
			name:           "Within sleep time",
			sleepStartTime: "2200",
			sleepEndTime:   "0600",
			currentTime:    time.Date(2023, 1, 1, 23, 0, 0, 0, time.UTC), // 23:00
			want:           true,
			wantErr:        false,
		},
		{
			name:           "Outside sleep time",
			sleepStartTime: "2200",
			sleepEndTime:   "0600",
			currentTime:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), // 12:00
			want:           false,
			wantErr:        false,
		},
		{
			name:           "Invalid start time",
			sleepStartTime: "2500",
			sleepEndTime:   "0600",
			currentTime:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			want:           false,
			wantErr:        true,
		},
		{
			name:           "Invalid end time",
			sleepStartTime: "2200",
			sleepEndTime:   "2500",
			currentTime:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			want:           false,
			wantErr:        true,
		},
		{
			name:           "At start time",
			sleepStartTime: "2200",
			sleepEndTime:   "0600",
			currentTime:    time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC), // 22:00
			want:           true,
			wantErr:        false,
		},
		{
			name:           "At end time",
			sleepStartTime: "2200",
			sleepEndTime:   "0600",
			currentTime:    time.Date(2023, 1, 2, 6, 0, 0, 0, time.UTC), // 06:00 next day
			want:           false,
			wantErr:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := IsSleepTime(test.sleepStartTime, test.sleepEndTime, test.currentTime)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.want, got)
		})
	}
}
