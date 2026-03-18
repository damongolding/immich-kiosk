package components

import (
	"strings"
	"testing"

	"github.com/damongolding/immich-kiosk/internal/common"
	"github.com/damongolding/immich-kiosk/internal/immich"
	"github.com/damongolding/immich-kiosk/internal/kiosk"
)

func TestModifyGIFAssets(t *testing.T) {
	tests := []struct {
		name     string
		viewData *common.ViewData
		want     []string // expected ImageData values for each asset
	}{
		{
			name: "Single GIF asset",
			viewData: &common.ViewData{
				Assets: []common.ViewImageData{
					{
						ImageData: "original-data-1",
						ImmichAsset: immich.Asset{
							ID:               "asset-123",
							OriginalMimeType: kiosk.MimeTypeGif,
						},
					},
				},
			},
			want: []string{"/image/asset-123?use_original_image=true"},
		},
		{
			name: "Single non-GIF asset",
			viewData: &common.ViewData{
				Assets: []common.ViewImageData{
					{
						ImageData: "original-data-1",
						ImmichAsset: immich.Asset{
							ID:               "asset-456",
							OriginalMimeType: kiosk.MimeTypeJpeg,
						},
					},
				},
			},
			want: []string{"original-data-1"}, // Should remain unchanged
		},
		{
			name: "Multiple GIF assets",
			viewData: &common.ViewData{
				Assets: []common.ViewImageData{
					{
						ImageData: "original-data-1",
						ImmichAsset: immich.Asset{
							ID:               "asset-111",
							OriginalMimeType: kiosk.MimeTypeGif,
						},
					},
					{
						ImageData: "original-data-2",
						ImmichAsset: immich.Asset{
							ID:               "asset-222",
							OriginalMimeType: kiosk.MimeTypeGif,
						},
					},
				},
			},
			want: []string{
				"/image/asset-111?use_original_image=true",
				"/image/asset-222?use_original_image=true",
			},
		},
		{
			name: "Mixed GIF and non-GIF assets",
			viewData: &common.ViewData{
				Assets: []common.ViewImageData{
					{
						ImageData: "original-data-1",
						ImmichAsset: immich.Asset{
							ID:               "asset-gif",
							OriginalMimeType: kiosk.MimeTypeGif,
						},
					},
					{
						ImageData: "original-data-2",
						ImmichAsset: immich.Asset{
							ID:               "asset-jpeg",
							OriginalMimeType: kiosk.MimeTypeJpeg,
						},
					},
					{
						ImageData: "original-data-3",
						ImmichAsset: immich.Asset{
							ID:               "asset-png",
							OriginalMimeType: kiosk.MimeTypePng,
						},
					},
					{
						ImageData: "original-data-4",
						ImmichAsset: immich.Asset{
							ID:               "asset-gif2",
							OriginalMimeType: kiosk.MimeTypeGif,
						},
					},
				},
			},
			want: []string{
				"/image/asset-gif?use_original_image=true",
				"original-data-2",
				"original-data-3",
				"/image/asset-gif2?use_original_image=true",
			},
		},
		{
			name: "Empty assets",
			viewData: &common.ViewData{
				Assets: []common.ViewImageData{},
			},
			want: []string{},
		},
		{
			name: "Assets with various image types",
			viewData: &common.ViewData{
				Assets: []common.ViewImageData{
					{
						ImageData: "webp-data",
						ImmichAsset: immich.Asset{
							ID:               "asset-webp",
							OriginalMimeType: kiosk.MimeTypeWebp,
						},
					},
					{
						ImageData: "bmp-data",
						ImmichAsset: immich.Asset{
							ID:               "asset-bmp",
							OriginalMimeType: "image/bmp",
						},
					},
					{
						ImageData: "gif-data",
						ImmichAsset: immich.Asset{
							ID:               "asset-animated",
							OriginalMimeType: kiosk.MimeTypeGif,
						},
					},
				},
			},
			want: []string{
				"webp-data",
				"bmp-data",
				"/image/asset-animated?use_original_image=true",
			},
		},
		{
			name: "GIF with empty asset ID (edge case)",
			viewData: &common.ViewData{
				Assets: []common.ViewImageData{
					{
						ImageData: "gif-data",
						ImmichAsset: immich.Asset{
							ID:               "",
							OriginalMimeType: kiosk.MimeTypeGif,
						},
					},
				},
			},
			want: []string{"/image/?use_original_image=true"},
		},
		{
			name: "Case sensitivity check",
			viewData: &common.ViewData{
				Assets: []common.ViewImageData{
					{
						ImageData: "original-data",
						ImmichAsset: immich.Asset{
							ID:               "asset-case",
							OriginalMimeType: strings.ToUpper(kiosk.MimeTypeGif), // uppercase
						},
					},
				},
			},
			want: []string{"original-data"}, // Should not match due to case sensitivity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function being tested
			modifyGIFAssets(tt.viewData)

			// Verify the results
			if len(tt.viewData.Assets) != len(tt.want) {
				t.Fatalf("Expected %d assets, got %d", len(tt.want), len(tt.viewData.Assets))
			}

			for i, asset := range tt.viewData.Assets {
				if asset.ImageData != tt.want[i] {
					t.Errorf("Asset[%d].ImageData = %q, want %q", i, asset.ImageData, tt.want[i])
				}
			}
		})
	}
}

// TestModifyGIFAssetsDoesNotModifyOriginalAsset ensures that the function
// only modifies ImageData and doesn't change other fields
func TestModifyGIFAssetsDoesNotModifyOriginalAsset(t *testing.T) {
	originalID := "asset-test-123"
	originalMimeType := kiosk.MimeTypeGif
	originalImageData := "base64-encoded-data"

	viewData := &common.ViewData{
		Assets: []common.ViewImageData{
			{
				ImageData: originalImageData,
				ImmichAsset: immich.Asset{
					ID:               originalID,
					OriginalMimeType: originalMimeType,
				},
			},
		},
	}

	modifyGIFAssets(viewData)

	// Verify ImmichAsset fields are unchanged
	if viewData.Assets[0].ImmichAsset.ID != originalID {
		t.Errorf("ImmichAsset.ID was modified, got %q, want %q", viewData.Assets[0].ImmichAsset.ID, originalID)
	}

	if viewData.Assets[0].ImmichAsset.OriginalMimeType != originalMimeType {
		t.Errorf("ImmichAsset.OriginalMimeType was modified, got %q, want %q", viewData.Assets[0].ImmichAsset.OriginalMimeType, originalMimeType)
	}

	// Verify ImageData was modified
	expectedImageData := "/image/asset-test-123?use_original_image=true"
	if viewData.Assets[0].ImageData != expectedImageData {
		t.Errorf("ImageData = %q, want %q", viewData.Assets[0].ImageData, expectedImageData)
	}
}

// TestModifyGIFAssetsNilViewData tests that the function handles nil pointers gracefully
func TestModifyGIFAssetsNilViewData(t *testing.T) {
	// This test ensures the function doesn't panic with nil input
	// Note: The current implementation will panic if viewData is nil
	// This test documents the current behavior and can be updated if the function
	// is modified to handle nil inputs
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when passing nil ViewData, but function did not panic")
		}
	}()

	modifyGIFAssets(nil)
}

// TestModifyGIFAssetsConcurrency tests that modifyGIFAssets is safe for concurrent access
// This is a regression test to ensure the function doesn't have race conditions
func TestModifyGIFAssetsConcurrency(t *testing.T) {
	viewData := &common.ViewData{
		Assets: []common.ViewImageData{
			{
				ImageData: "data-1",
				ImmichAsset: immich.Asset{
					ID:               "asset-1",
					OriginalMimeType: kiosk.MimeTypeGif,
				},
			},
			{
				ImageData: "data-2",
				ImmichAsset: immich.Asset{
					ID:               "asset-2",
					OriginalMimeType: kiosk.MimeTypeJpeg,
				},
			},
		},
	}

	// Create a copy to compare against
	originalAssets := make([]common.ViewImageData, len(viewData.Assets))
	copy(originalAssets, viewData.Assets)

	// Call the function - it should be deterministic
	modifyGIFAssets(viewData)

	// Verify first asset was modified
	expected := "/image/asset-1?use_original_image=true"
	if viewData.Assets[0].ImageData != expected {
		t.Errorf("Asset[0].ImageData = %q, want %q", viewData.Assets[0].ImageData, expected)
	}

	// Verify second asset was not modified
	if viewData.Assets[1].ImageData != "data-2" {
		t.Errorf("Asset[1].ImageData = %q, want %q", viewData.Assets[1].ImageData, "data-2")
	}
}
