package immich

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestArchiveLogic tests the handling of archived and trashed assets
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
			simulatedContinueTriggered := test.Type != "IMAGE" || test.IsTrashed || (test.IsArchived && !test.ArchivedWantedByUser)

			assert.Equal(t, test.WantSimulatedContinue, simulatedContinueTriggered, "Unexpected simulatedContinueTriggered value")
		})
	}
}

// TestFacesCenterPoint tests the calculation of the center point between detected faces in an asset
func TestFacesCenterPoint(t *testing.T) {

	tests := []struct {
		name  string
		asset Asset
		wantX float64
		wantY float64
	}{
		{
			name: "No people",
			asset: Asset{
				People:          []Person{},
				UnassignedFaces: []Face{},
			},
			wantX: 0,
			wantY: 0,
		},
		{
			name: "People but no faces",
			asset: Asset{
				People: []Person{
					{Faces: []Face{{BoundingBoxX1: 0, BoundingBoxY1: 0, BoundingBoxX2: 0, BoundingBoxY2: 0, ImageWidth: 1000, ImageHeight: 1000}}},
					{Faces: []Face{{BoundingBoxX1: 0, BoundingBoxY1: 0, BoundingBoxX2: 0, BoundingBoxY2: 0, ImageWidth: 1000, ImageHeight: 1000}}},
				},
				UnassignedFaces: []Face{},
			},
			wantX: 0,
			wantY: 0,
		},
		{
			name: "Zero dimensions",
			asset: Asset{
				People: []Person{
					{Faces: []Face{{BoundingBoxX1: 10, BoundingBoxY1: 10, BoundingBoxX2: 20, BoundingBoxY2: 20, ImageWidth: 0, ImageHeight: 0}}},
				},
				UnassignedFaces: []Face{},
			},
			wantX: 0,
			wantY: 0,
		},
		{
			name: "Single face",
			asset: Asset{
				People: []Person{
					{Faces: []Face{{BoundingBoxX1: 100, BoundingBoxY1: 100, BoundingBoxX2: 200, BoundingBoxY2: 200, ImageWidth: 1000, ImageHeight: 1000}}},
				},
				UnassignedFaces: []Face{},
			},
			wantX: 15,
			wantY: 15,
		},
		{
			name: "Multiple faces",
			asset: Asset{
				People: []Person{
					{Faces: []Face{{BoundingBoxX1: 100, BoundingBoxY1: 100, BoundingBoxX2: 200, BoundingBoxY2: 200, ImageWidth: 1000, ImageHeight: 1000}}},
					{Faces: []Face{{BoundingBoxX1: 300, BoundingBoxY1: 300, BoundingBoxX2: 400, BoundingBoxY2: 400, ImageWidth: 1000, ImageHeight: 1000}}},
				},
				UnassignedFaces: []Face{},
			},
			wantX: 25,
			wantY: 25,
		},
		{
			name: "Multiple faces but not on the first person",
			asset: Asset{
				People: []Person{
					{Faces: []Face{{BoundingBoxX1: 0, BoundingBoxY1: 0, BoundingBoxX2: 0, BoundingBoxY2: 0, ImageWidth: 1000, ImageHeight: 1000}}},
					{Faces: []Face{{BoundingBoxX1: 100, BoundingBoxY1: 100, BoundingBoxX2: 200, BoundingBoxY2: 200, ImageWidth: 1000, ImageHeight: 1000}}},
					{Faces: []Face{{BoundingBoxX1: 300, BoundingBoxY1: 300, BoundingBoxX2: 400, BoundingBoxY2: 400, ImageWidth: 1000, ImageHeight: 1000}}},
				},
				UnassignedFaces: []Face{},
			},
			wantX: 25,
			wantY: 25,
		},
		{
			name: "Multiple faces but not on the second person",
			asset: Asset{
				People: []Person{
					{Faces: []Face{{BoundingBoxX1: 100, BoundingBoxY1: 100, BoundingBoxX2: 200, BoundingBoxY2: 200, ImageWidth: 1000, ImageHeight: 1000}}},
					{Faces: []Face{{BoundingBoxX1: 0, BoundingBoxY1: 0, BoundingBoxX2: 0, BoundingBoxY2: 0, ImageWidth: 1000, ImageHeight: 1000}}},
					{Faces: []Face{{BoundingBoxX1: 300, BoundingBoxY1: 300, BoundingBoxX2: 400, BoundingBoxY2: 400, ImageWidth: 1000, ImageHeight: 1000}}},
				},
				UnassignedFaces: []Face{},
			},
			wantX: 25,
			wantY: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := tt.asset.FacesCenterPoint()
			assert.Equal(t, tt.wantX, gotX, "Unexpected X coordinate")
			assert.Equal(t, tt.wantY, gotY, "Unexpected Y coordinate")
		})
	}
}

// TestRemoveExcludedAlbums tests the functionality to remove specific albums from a list
func TestRemoveExcludedAlbums(t *testing.T) {
	tests := []struct {
		name     string
		albums   Albums
		exclude  []string
		expected Albums
	}{
		{
			name: "removes excluded albums",
			albums: Albums{
				{ID: "1"},
				{ID: "2"},
				{ID: "3"},
			},
			exclude: []string{"2"},
			expected: Albums{
				{ID: "1"},
				{ID: "3"},
			},
		},
		{
			name: "handles empty exclude list",
			albums: Albums{
				{ID: "1"},
				{ID: "2"},
			},
			exclude: []string{},
			expected: Albums{
				{ID: "1"},
				{ID: "2"},
			},
		},
		{
			name:     "handles empty albums list",
			albums:   Albums{},
			exclude:  []string{"1"},
			expected: Albums{},
		},
		{
			name: "handles multiple excludes",
			albums: Albums{
				{ID: "1"},
				{ID: "2"},
				{ID: "3"},
				{ID: "4"},
			},
			exclude: []string{"1", "3", "4"},
			expected: Albums{
				{ID: "2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			albums := tt.albums
			albums.RemoveExcludedAlbums(tt.exclude)
			assert.Equal(t, tt.expected, albums, "RemoveExcludedAlbums returned unexpected result")
		})
	}
}

func TestExtractDays(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "valid number",
			input:   "last_7",
			want:    7,
			wantErr: false,
		},
		{
			name:    "no number",
			input:   "last_",
			want:    0,
			wantErr: true,
		},
		{
			name:    "multiple numbers",
			input:   "last_12_34",
			want:    12,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractDays(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractDays() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractDays() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMergeAssetInfo tests the merging of asset information between two Asset structs.
// It verifies:
// - Empty slices are correctly merged with populated ones
// - Non-empty slices are preserved and not overwritten
// - Boolean fields are always updated regardless of value
// - Zero/empty values are updated while non-zero values are preserved
// - Complex nested structs are merged properly while maintaining existing data
func TestMergeAssetInfo(t *testing.T) {
	tests := []struct {
		name           string
		baseAsset      Asset
		additionalInfo Asset
		wantErr        bool
		expected       Asset
	}{
		{
			name: "merge empty slices",
			baseAsset: Asset{
				People: []Person{},
			},
			additionalInfo: Asset{
				People: []Person{{ID: "1", Name: "Test"}},
			},
			wantErr: false,
			expected: Asset{
				People: []Person{{ID: "1", Name: "Test"}},
			},
		},
		{
			name: "don't overwrite non-empty slices",
			baseAsset: Asset{
				People: []Person{{ID: "1", Name: "Original"}},
			},
			additionalInfo: Asset{
				People: []Person{{ID: "2", Name: "New"}},
			},
			wantErr: false,
			expected: Asset{
				People: []Person{{ID: "1", Name: "Original"}},
			},
		},
		{
			name: "always update booleans",
			baseAsset: Asset{
				IsArchived: false,
			},
			additionalInfo: Asset{
				IsArchived: true,
			},
			wantErr: false,
			expected: Asset{
				IsArchived: true,
			},
		},
		{
			name: "update zero values only",
			baseAsset: Asset{
				ID:   "",
				Type: "image",
			},
			additionalInfo: Asset{
				ID:   "new-id",
				Type: "video",
			},
			wantErr: false,
			expected: Asset{
				ID:   "new-id",
				Type: "image",
			},
		},
		{
			name: "full merge test",
			baseAsset: Asset{
				ID:         "base-id",
				Type:       "image",
				IsArchived: false,
				People:     []Person{{ID: "1", Name: "Original"}},
			},
			additionalInfo: Asset{
				ID:         "new-id",
				Type:       "video",
				IsArchived: true,
				People:     []Person{{ID: "2", Name: "New"}},
				ExifInfo: ExifInfo{
					Make:  "New",
					Model: "New",
				},
			},
			wantErr: false,
			expected: Asset{
				ID:         "base-id",
				Type:       "image",
				IsArchived: true,
				People:     []Person{{ID: "1", Name: "Original"}},
				ExifInfo: ExifInfo{
					Make:  "New",
					Model: "New",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.baseAsset.mergeAssetInfo(tt.additionalInfo)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, tt.baseAsset)
		})
	}
}

func TestTagMatches(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		value    string
		expected bool
	}{
		// Exact match tests
		{
			name:     "exact match - same case",
			pattern:  "parent",
			value:    "parent",
			expected: true,
		},
		{
			name:     "exact match - different case",
			pattern:  "Parent",
			value:    "PARENT",
			expected: true,
		},
		{
			name:     "exact match - with leading/trailing slashes",
			pattern:  "/parent/",
			value:    "/parent/",
			expected: true,
		},
		{
			name:     "exact match - no match",
			pattern:  "parent",
			value:    "other",
			expected: false,
		},
		{
			name:     "exact match - parent vs child",
			pattern:  "parent",
			value:    "parent/child",
			expected: false,
		},

		// Single-level wildcard tests (parent/*)
		{
			name:     "single wildcard - does not match parent itself",
			pattern:  "parent/*",
			value:    "parent",
			expected: false,
		},
		{
			name:     "single wildcard - matches direct child",
			pattern:  "parent/*",
			value:    "parent/child",
			expected: true,
		},
		{
			name:     "single wildcard - does not match grandchild",
			pattern:  "parent/*",
			value:    "parent/child/grandchild",
			expected: false,
		},
		{
			name:     "single wildcard - matches another direct child",
			pattern:  "parent/*",
			value:    "parent/another",
			expected: true,
		},
		{
			name:     "single wildcard - does not match different parent",
			pattern:  "parent/*",
			value:    "other/child",
			expected: false,
		},
		{
			name:     "single wildcard - case insensitive",
			pattern:  "Parent/*",
			value:    "PARENT/CHILD",
			expected: true,
		},

		// Recursive wildcard tests (parent/**)
		{
			name:     "recursive wildcard - does not match parent itself",
			pattern:  "parent/**",
			value:    "parent",
			expected: false,
		},
		{
			name:     "recursive wildcard - matches direct child",
			pattern:  "parent/**",
			value:    "parent/child",
			expected: true,
		},
		{
			name:     "recursive wildcard - matches grandchild",
			pattern:  "parent/**",
			value:    "parent/child/grandchild",
			expected: true,
		},
		{
			name:     "recursive wildcard - matches deep nesting",
			pattern:  "parent/**",
			value:    "parent/a/b/c/d/e",
			expected: true,
		},
		{
			name:     "recursive wildcard - does not match different parent",
			pattern:  "parent/**",
			value:    "other/child",
			expected: false,
		},
		{
			name:     "recursive wildcard - does not match partial prefix",
			pattern:  "parent/**",
			value:    "parental/child",
			expected: false,
		},
		{
			name:     "recursive wildcard - case insensitive",
			pattern:  "Parent/**",
			value:    "PARENT/CHILD/GRANDCHILD",
			expected: true,
		},

		// Edge cases
		{
			name:     "empty pattern and value",
			pattern:  "",
			value:    "",
			expected: true,
		},
		{
			name:     "empty pattern",
			pattern:  "",
			value:    "something",
			expected: false,
		},
		{
			name:     "empty value",
			pattern:  "something",
			value:    "",
			expected: false,
		},
		{
			name:     "nested parent - exact match",
			pattern:  "vacation/2023",
			value:    "vacation/2023",
			expected: true,
		},
		{
			name:     "nested parent - single wildcard",
			pattern:  "vacation/2023/*",
			value:    "vacation/2023/summer",
			expected: true,
		},
		{
			name:     "nested parent - recursive wildcard",
			pattern:  "vacation/2023/**",
			value:    "vacation/2023/summer/beach",
			expected: true,
		},
		{
			name:     "multiple slashes normalized",
			pattern:  "//parent//",
			value:    "parent",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesTagPattern(tt.value, tt.pattern)
			if result != tt.expected {
				t.Errorf("tagMatches(%q, %q) = %v, expected %v",
					tt.pattern, tt.value, result, tt.expected)
			}
		})
	}
}

func TestExpandTags(t *testing.T) {

	tests := []struct {
		name     string
		tag      string
		expected []string
	}{
		// Exact match tests
		{
			name:     "nested parent - single wildcard",
			tag:      "parent/child/*",
			expected: []string{"parent/child/grand-child"},
		},
		{
			name:     "nested parent - recursive wildcard",
			tag:      "parent/**",
			expected: []string{"parent/child", "parent/child/grand-child", "parent/child/grand-child/great-grand-child"},
		},
		{
			name:     "nested child - recursive wildcard",
			tag:      "parent/child/**",
			expected: []string{"parent/child/grand-child", "parent/child/grand-child/great-grand-child"},
		},
	}

	allTags := []Tag{
		{Value: "parent"},
		{Value: "parent/child"},
		{Value: "parent/child/grand-child"},
		{Value: "parent/child/grand-child/great-grand-child"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			expandedTags := []string{}

			expandedTags = addRecursiveTags(tt.tag, expandedTags, allTags)

			if slices.Compare(expandedTags, tt.expected) != 0 {
				t.Errorf("expandedTags = %v, expected %v",
					expandedTags, tt.expected)
			}
		})
	}
}
