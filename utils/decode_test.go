package utils

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func sampleImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 40, 24))
	for x := 0; x < 40; x++ {
		for y := 0; y < 24; y++ {
			img.Set(x, y, color.RGBA{uint8(x * 6), uint8(y * 10), 128, 255})
		}
	}
	return img
}

// TestDecodeConfig checks the header decoders report the right format and
// dimensions for JPEG and PNG — the read path deliberately uses the standard
// library rather than the registry (where jpegli's decoder can panic).
func TestDecodeConfig(t *testing.T) {
	var jpg bytes.Buffer
	if err := jpeg.Encode(&jpg, sampleImage(), nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, sampleImage()); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	for _, tc := range []struct {
		name   string
		data   []byte
		format string
	}{
		{"jpeg", jpg.Bytes(), "jpeg"},
		{"png", pngBuf.Bytes(), "png"},
	} {
		cfg, format, err := DecodeConfig(bytes.NewReader(tc.data))
		if err != nil {
			t.Errorf("%s: DecodeConfig: %v", tc.name, err)
			continue
		}
		if format != tc.format {
			t.Errorf("%s: format = %q, want %q", tc.name, format, tc.format)
		}
		if cfg.Width != 40 || cfg.Height != 24 {
			t.Errorf("%s: dims = %dx%d, want 40x24", tc.name, cfg.Width, cfg.Height)
		}

		img, format, err := DecodeImage(bytes.NewReader(tc.data))
		if err != nil {
			t.Errorf("%s: DecodeImage: %v", tc.name, err)
			continue
		}
		if format != tc.format {
			t.Errorf("%s: DecodeImage format = %q, want %q", tc.name, format, tc.format)
		}
		if b := img.Bounds(); b.Dx() != 40 || b.Dy() != 24 {
			t.Errorf("%s: image bounds = %v", tc.name, b)
		}
	}
}
