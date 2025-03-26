package immich

import (
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
			simulatedContinueTriggered := false

			if test.Type != "IMAGE" || test.IsTrashed || (test.IsArchived && !test.ArchivedWantedByUser) {
				simulatedContinueTriggered = true
			}

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
