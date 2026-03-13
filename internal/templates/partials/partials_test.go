package partials

import (
	"testing"

	"github.com/damongolding/immich-kiosk/internal/immich"
)

func TestAssetCameraData(t *testing.T) {
	tests := []struct {
		name string
		exif immich.ExifInfo
		want string
	}{
		{
			name: "Make already in model (whole word)",
			exif: immich.ExifInfo{
				Make:  "Canon",
				Model: "Canon EOS 5D Mark IV",
			},
			want: "Canon EOS 5D Mark IV",
		},
		{
			name: "Make not in model",
			exif: immich.ExifInfo{
				Make:  "Canon",
				Model: "EOS 5D Mark IV",
			},
			want: "Canon EOS 5D Mark IV",
		},
		{
			name: "Make is substring but not whole word",
			exif: immich.ExifInfo{
				Make:  "Canon",
				Model: "Canonic EOS 5D Mark IV",
			},
			want: "Canon Canonic EOS 5D Mark IV",
		},
		{
			name: "Extra whitespace trimmed",
			exif: immich.ExifInfo{
				Make:  " Canon ",
				Model: " EOS 5D Mark IV ",
			},
			want: "Canon EOS 5D Mark IV",
		},
		{
			name: "Empty make",
			exif: immich.ExifInfo{
				Make:  "",
				Model: "EOS 5D Mark IV",
			},
			want: "EOS 5D Mark IV",
		},
		{
			name: "Empty model",
			exif: immich.ExifInfo{
				Make:  "Canon",
				Model: "",
			},
			want: "Canon",
		},
		{
			name: "Both empty",
			exif: immich.ExifInfo{
				Make:  "",
				Model: "",
			},
			want: "",
		},
		{
			name: "Case insensitive whole word match",
			exif: immich.ExifInfo{
				Make:  "canon",
				Model: "Canon EOS 5D Mark IV",
			},
			want: "Canon EOS 5D Mark IV",
		},
		{
			name: "My camera",
			exif: immich.ExifInfo{
				Make:  "NIKON CORPORATION",
				Model: "NIKON D90",
			},
			want: "NIKON CORPORATION NIKON D90",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AssetCameraData(tt.exif)
			if got != tt.want {
				t.Errorf("AssetCameraData() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAssetExif(t *testing.T) {
	tests := []struct {
		name string
		exif immich.ExifInfo
		want string
	}{
		{
			name: "Empty EXIF data",
			exif: immich.ExifInfo{},
			want: "",
		},
		{
			name: "All zeros",
			exif: immich.ExifInfo{
				FNumber:      0,
				ExposureTime: "",
				FocalLength:  0,
				Iso:          0,
			},
			want: "",
		},
		{
			name: "Only FNumber",
			exif: immich.ExifInfo{
				FNumber: 2.8,
			},
			want: `<span class="asset--metadata--exif--fnumber">&#402;</span>/2.8`,
		},
		{
			name: "Only ExposureTime",
			exif: immich.ExifInfo{
				ExposureTime: "1/250",
			},
			want: `1/250<small>s</small>`,
		},
		{
			name: "Only ISO",
			exif: immich.ExifInfo{
				Iso: 800,
			},
			want: `ISO 800`,
		},
		{
			name: "Focal length - whole number (no trailing zeros)",
			exif: immich.ExifInfo{
				FocalLength: 50.00,
			},
			want: `50mm`,
		},
		{
			name: "Focal length - one decimal place",
			exif: immich.ExifInfo{
				FocalLength: 35.5,
			},
			want: `35.5mm`,
		},
		{
			name: "Focal length - two decimal places",
			exif: immich.ExifInfo{
				FocalLength: 24.25,
			},
			want: `24.25mm`,
		},
		{
			name: "Focal length - trailing zero removed",
			exif: immich.ExifInfo{
				FocalLength: 85.10,
			},
			want: `85.1mm`,
		},
		{
			name: "Focal length - small value with decimals",
			exif: immich.ExifInfo{
				FocalLength: 16.35,
			},
			want: `16.35mm`,
		},
		{
			name: "Focal length - large whole number",
			exif: immich.ExifInfo{
				FocalLength: 200.0,
			},
			want: `200mm`,
		},
		{
			name: "Focal length - very precise value",
			exif: immich.ExifInfo{
				FocalLength: 70.99,
			},
			want: `70.99mm`,
		},
		{
			name: "FNumber and ExposureTime",
			exif: immich.ExifInfo{
				FNumber:      1.8,
				ExposureTime: "1/60",
			},
			want: `<span class="asset--metadata--exif--fnumber">&#402;</span>/1.8<span class="asset--metadata--exif--seperator">&#124;</span>1/60<small>s</small>`,
		},
		{
			name: "FNumber and FocalLength",
			exif: immich.ExifInfo{
				FNumber:     2.8,
				FocalLength: 50.0,
			},
			want: `<span class="asset--metadata--exif--fnumber">&#402;</span>/2.8<span class="asset--metadata--exif--seperator">&#124;</span>50mm`,
		},
		{
			name: "FNumber and ISO",
			exif: immich.ExifInfo{
				FNumber: 4.0,
				Iso:     400,
			},
			want: `<span class="asset--metadata--exif--fnumber">&#402;</span>/4.0<span class="asset--metadata--exif--seperator">&#124;</span>ISO 400`,
		},
		{
			name: "ExposureTime and FocalLength",
			exif: immich.ExifInfo{
				ExposureTime: "1/125",
				FocalLength:  35.0,
			},
			want: `1/125<small>s</small><span class="asset--metadata--exif--seperator">&#124;</span>35mm`,
		},
		{
			name: "ExposureTime and ISO",
			exif: immich.ExifInfo{
				ExposureTime: "1/500",
				Iso:          200,
			},
			want: `1/500<small>s</small><span class="asset--metadata--exif--seperator">&#124;</span>ISO 200`,
		},
		{
			name: "FocalLength and ISO",
			exif: immich.ExifInfo{
				FocalLength: 85.0,
				Iso:         1600,
			},
			want: `85mm<span class="asset--metadata--exif--seperator">&#124;</span>ISO 1600`,
		},
		{
			name: "All EXIF fields populated",
			exif: immich.ExifInfo{
				FNumber:      1.4,
				ExposureTime: "1/1000",
				FocalLength:  85.0,
				Iso:          100,
			},
			want: `<span class="asset--metadata--exif--fnumber">&#402;</span>/1.4<span class="asset--metadata--exif--seperator">&#124;</span>1/1000<small>s</small><span class="asset--metadata--exif--seperator">&#124;</span>85mm<span class="asset--metadata--exif--seperator">&#124;</span>ISO 100`,
		},
		{
			name: "All fields with focal length having decimals",
			exif: immich.ExifInfo{
				FNumber:      5.6,
				ExposureTime: "1/200",
				FocalLength:  105.5,
				Iso:          3200,
			},
			want: `<span class="asset--metadata--exif--fnumber">&#402;</span>/5.6<span class="asset--metadata--exif--seperator">&#124;</span>1/200<small>s</small><span class="asset--metadata--exif--seperator">&#124;</span>105.5mm<span class="asset--metadata--exif--seperator">&#124;</span>ISO 3200`,
		},
		{
			name: "Focal length edge case - 0.5",
			exif: immich.ExifInfo{
				FocalLength: 0.5,
			},
			want: `0.5mm`,
		},
		{
			name: "Focal length edge case - 1.0",
			exif: immich.ExifInfo{
				FocalLength: 1.0,
			},
			want: `1mm`,
		},
		{
			name: "Long exposure time format",
			exif: immich.ExifInfo{
				ExposureTime: "2.5",
			},
			want: `2.5<small>s</small>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AssetExif(tt.exif)
			if got != tt.want {
				t.Errorf("AssetExif() = %q, want %q", got, tt.want)
			}
		})
	}
}