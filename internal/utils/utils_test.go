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
	baseQueries.Set("duration", "60")
	baseQueries.Set("transition", "fade")
	baseQueries.Set("raw", "false")

	refererQueries := "/demo-url?transition=none&image_fit=cover&raw=true"

	q, err := CombineQueries(baseQueries, refererQueries)
	assert.NoError(t, err)

	assert.Equal(t, []string{"fade", "none"}, q["transition"])
	assert.Equal(t, "cover", q.Get("image_fit"))
	assert.Equal(t, "60", q.Get("duration"))
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

// TestRandomMutate tests that RandomItem does not modify the original slice.
// Ensures the function operates non-destructively on input slices.
func TestRandomMutate(t *testing.T) {
	original := []int{1, 2, 3, 4}
	originalCopy := []int{1, 2, 3, 4}

	RandomItem(original)

	assert.Equal(t, originalCopy, original, "RandomItem should not mutate the original slice")

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
		{
			name:           "outside of sleep time",
			sleepStartTime: "1000",
			sleepEndTime:   "1022",
			currentTime:    time.Date(2023, 1, 2, 10, 23, 0, 0, time.UTC), // 10:23
			want:           false,
			wantErr:        false,
		},
		{
			name:           "inside of sleep time",
			sleepStartTime: "1000",
			sleepEndTime:   "1022",
			currentTime:    time.Date(2023, 1, 2, 10, 21, 0, 0, time.UTC), // 10:21
			want:           true,
			wantErr:        false,
		},
		{
			name:           "overnight: in sleep time",
			sleepStartTime: "22",
			sleepEndTime:   "7",
			currentTime:    time.Date(2023, 1, 2, 2, 00, 0, 0, time.UTC), // 02:00
			want:           true,
			wantErr:        false,
		},

		{
			name:           "overnight: out of sleep time",
			sleepStartTime: "22",
			sleepEndTime:   "7",
			currentTime:    time.Date(2023, 1, 2, 8, 00, 0, 0, time.UTC), // 08:00
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

func TestParseTimeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"1230", "12:30", false},
		{"0130", "01:30", false},
		{"2359", "23:59", false},
		{"0000", "00:00", false},
		{"130", "01:30", false},
		{"12:30", "12:30", false},
		{"1-30", "01:30", false},
		{"25:00", "", true},
		{"12:60", "", true},
		{"abcd", "", true},
		{"", "", true},
		{" ", "", true},
		{"730", "07:30", false},
		{"9:30", "09:30", false},
		{"19:30", "19:30", false},
		{"0930", "09:30", false},
		{"18", "18:00", false},
		{"5", "05:00", false},
		{"23", "23:00", false},
		{"9", "09:00", false},
		{"93015", "09:30", true},
		{"24", "", true},
		{"960", "09:60", true},
	}

	for _, test := range tests {
		parsed, err := parseTimeString(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
			} else if parsed.Format("15:04") != test.expected {
				t.Errorf("For input %s, expected %s, but got %s", test.input, test.expected, parsed.Format("15:04"))
			}
		}
	}
}

// TestMergeQueries tests the MergeQueries function
func TestMergeQueries(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		queryA   url.Values
		queryB   url.Values
		expected url.Values
	}{
		{
			name:     "Empty queries",
			queryA:   url.Values{},
			queryB:   url.Values{},
			expected: url.Values{},
		},
		{
			name:     "One empty query",
			queryA:   url.Values{"key": []string{"value"}},
			queryB:   url.Values{},
			expected: url.Values{"key": []string{"value"}},
		},
		{
			name:     "Distinct keys",
			queryA:   url.Values{"keyA": []string{"valueA"}},
			queryB:   url.Values{"keyB": []string{"valueB"}},
			expected: url.Values{"keyA": []string{"valueA"}, "keyB": []string{"valueB"}},
		},
		{
			name:     "Same keys",
			queryA:   url.Values{"key": []string{"valueA"}},
			queryB:   url.Values{"key": []string{"valueB"}},
			expected: url.Values{"key": []string{"valueB", "valueA"}},
		},
		{
			name:     "Multiple values per key",
			queryA:   url.Values{"key": []string{"valueA1", "valueA2"}},
			queryB:   url.Values{"key": []string{"valueB1", "valueB2"}},
			expected: url.Values{"key": []string{"valueB1", "valueB2", "valueA1", "valueA2"}},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeQueries(tt.queryA, tt.queryB)

			// Compare lengths
			assert.Equal(t, len(tt.expected), len(result), "MergeQueries() got length %v, want %v", len(result), len(tt.expected))

			// Compare values
			for key, expectedVals := range tt.expected {
				resultVals, exists := result[key]
				assert.True(t, exists, "MergeQueries() missing key %v", key)
				assert.Equal(t, len(expectedVals), len(resultVals), "MergeQueries() key %v got %v values, want %v", key, len(resultVals), len(expectedVals))

				for i, val := range expectedVals {
					assert.Equal(t, val, resultVals[i], "MergeQueries() key %v index %v got %v, want %v", key, i, resultVals[i], val)
				}
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "Valid bytes",
			input:   "1024B",
			want:    1024,
			wantErr: false,
		},
		{
			name:    "Valid kilobytes",
			input:   "1KB",
			want:    1024,
			wantErr: false,
		},
		{
			name:    "Valid megabytes",
			input:   "1MB",
			want:    1024 * 1024,
			wantErr: false,
		},
		{
			name:    "Valid gigabytes",
			input:   "1GB",
			want:    1024 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   "1XB",
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Invalid number",
			input:   "abcKB",
			wantErr: true,
		},
		{
			name:    "With space",
			input:   "1 KB",
			want:    1024,
			wantErr: false,
		},
		{
			name:    "Lowercase units",
			input:   "1kb",
			want:    1024,
			wantErr: false,
		},
		{
			name:    "Mixed case",
			input:   "1Kb",
			want:    1024,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssetWeight(t *testing.T) {
	tests := []struct {
		name      string
		asset     AssetWithWeighting
		wantRatio float64 // relative to base (Penalty=1)
	}{
		{
			name: "Normal penalty",
			asset: AssetWithWeighting{
				Weight:  100,
				Penalty: 1.0,
			},
			wantRatio: 1.0,
		},
		{
			name: "Half penalty",
			asset: AssetWithWeighting{
				Weight:  100,
				Penalty: 0.5,
			},
			wantRatio: 0.5,
		},
		{
			name: "Double penalty",
			asset: AssetWithWeighting{
				Weight:  100,
				Penalty: 2.0,
			},
			wantRatio: 2.0,
		},
		{
			name: "Zero penalty defaults",
			asset: AssetWithWeighting{
				Weight:  100,
				Penalty: 0,
			},
			wantRatio: 1.0,
		},
		{
			name: "Negative penalty defaults",
			asset: AssetWithWeighting{
				Weight:  100,
				Penalty: -1,
			},
			wantRatio: 1.0,
		},
	}

	base := assetWeight(AssetWithWeighting{
		Weight:  100,
		Penalty: 1.0,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := assetWeight(tt.asset)
			want := base * tt.wantRatio

			if math.Abs(got-want) > 0.0001 {
				t.Errorf("assetWeight() = %f, want %f", got, want)
			}
		})
	}
}

func TestCalculateTotalWeight(t *testing.T) {
	tests := []struct {
		name   string
		assets []AssetWithWeighting
	}{
		{
			name: "Single asset",
			assets: []AssetWithWeighting{
				{Weight: 10, Penalty: 1.0},
			},
		},
		{
			name: "Multiple assets equal weight",
			assets: []AssetWithWeighting{
				{Weight: 10, Penalty: 1.0},
				{Weight: 10, Penalty: 1.0},
				{Weight: 10, Penalty: 1.0},
			},
		},
		{
			name: "Mixed penalties",
			assets: []AssetWithWeighting{
				{Weight: 10, Penalty: 1.0},
				{Weight: 10, Penalty: 0.5},
				{Weight: 10, Penalty: 2.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expected float64
			for _, a := range tt.assets {
				expected += assetWeight(a)
			}

			got := calculateTotalWeight(tt.assets)

			if math.Abs(got-expected) > 0.0001 {
				t.Errorf("calculateTotalWeight() = %f, want %f", got, expected)
			}
		})
	}
}

func TestWeightedRandomItem(t *testing.T) {
	tests := []struct {
		name       string
		assets     []AssetWithWeighting
		expectMore string
		expectLess string
	}{
		{
			name: "Penalty reduces selection",
			assets: []AssetWithWeighting{
				{
					Asset:   WeightedAsset{ID: "A"},
					Weight:  100,
					Penalty: 1.0,
				},
				{
					Asset:   WeightedAsset{ID: "B"},
					Weight:  100,
					Penalty: 0.25,
				},
			},
			expectMore: "A",
			expectLess: "B",
		},
		{
			name: "Higher penalty increases selection",
			assets: []AssetWithWeighting{
				{
					Asset:   WeightedAsset{ID: "A"},
					Weight:  100,
					Penalty: 2.0,
				},
				{
					Asset:   WeightedAsset{ID: "B"},
					Weight:  100,
					Penalty: 1.0,
				},
			},
			expectMore: "A",
			expectLess: "B",
		},
	}

	const iterations = 50_000

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts := make(map[string]int)

			for range iterations {
				a := WeightedRandomItem(tt.assets)
				counts[a.ID]++
			}

			if counts[tt.expectMore] <= counts[tt.expectLess] {
				t.Errorf(
					"expected %s to be selected more often than %s (%d vs %d)",
					tt.expectMore,
					tt.expectLess,
					counts[tt.expectMore],
					counts[tt.expectLess],
				)
			}
		})
	}
}

func TestWeightedRandomItem_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		assets []AssetWithWeighting
		wantID string
	}{
		{
			name:   "Empty slice",
			assets: nil,
			wantID: "",
		},
		{
			name: "Single asset",
			assets: []AssetWithWeighting{
				{
					Asset:   WeightedAsset{ID: "only"},
					Weight:  1,
					Penalty: 1.0,
				},
			},
			wantID: "only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WeightedRandomItem(tt.assets)
			if got.ID != tt.wantID {
				t.Errorf("WeightedRandomItem() = %q, want %q", got.ID, tt.wantID)
			}
		})
	}
}
