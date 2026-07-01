package utils

import (
	"bufio"
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

var (
	jpegMagic = []byte{0xff, 0xd8}
	pngMagic  = []byte("\x89PNG\r\n\x1a\n")
)

// DecodeConfig reads an image header, sending JPEG and PNG straight to the
// standard-library decoders instead of through the global image registry. The
// jpegli package (imported for its faster encoder) registers a WASM-backed JPEG
// decoder that panics on some otherwise-valid JPEGs, so decoding those ourselves
// keeps the robust standard-library decoder on the read path.
func DecodeConfig(r io.Reader) (image.Config, string, error) {
	br := bufio.NewReader(r)
	switch magic, _ := br.Peek(8); {
	case bytes.HasPrefix(magic, jpegMagic):
		cfg, err := jpeg.DecodeConfig(br)
		return cfg, "jpeg", err
	case bytes.HasPrefix(magic, pngMagic):
		cfg, err := png.DecodeConfig(br)
		return cfg, "png", err
	default:
		return image.DecodeConfig(br)
	}
}

// DecodeImage decodes a whole image, sending JPEG and PNG straight to the
// standard-library decoders (see DecodeConfig for why).
func DecodeImage(r io.Reader) (image.Image, string, error) {
	br := bufio.NewReader(r)
	switch magic, _ := br.Peek(8); {
	case bytes.HasPrefix(magic, jpegMagic):
		img, err := jpeg.Decode(br)
		return img, "jpeg", err
	case bytes.HasPrefix(magic, pngMagic):
		img, err := png.Decode(br)
		return img, "png", err
	default:
		return image.Decode(br)
	}
}
