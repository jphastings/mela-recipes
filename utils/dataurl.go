package utils

import (
	"encoding/base64"
	"net/http"
	"strings"
)

// DecodeDataURL decodes an RFC 2397 "data:" URL into raw image bytes. Only
// base64-encoded payloads are accepted (text payloads aren't usable images); it
// returns false for non-data URLs or malformed/empty payloads.
func DecodeDataURL(s string) (B64Image, bool) {
	if !strings.HasPrefix(s, "data:") {
		return nil, false
	}
	comma := strings.IndexByte(s, ',')
	if comma < 0 {
		return nil, false
	}
	if !strings.Contains(s[len("data:"):comma], ";base64") {
		return nil, false
	}
	data, err := base64.StdEncoding.DecodeString(s[comma+1:])
	if err != nil || len(data) == 0 {
		return nil, false
	}
	return B64Image(data), true
}

// ImageMIME sniffs the image media type from its bytes, defaulting to
// "image/jpeg" when the content can't be classified as a known image type.
func ImageMIME(data []byte) string {
	if ct := http.DetectContentType(data); strings.HasPrefix(ct, "image/") {
		return ct
	}
	return "image/jpeg"
}

// DataURL renders the image as a base64-encoded "data:" URL, sniffing its MIME
// type from the bytes.
func (i B64Image) DataURL() string {
	return "data:" + ImageMIME(i) + ";base64," + base64.StdEncoding.EncodeToString(i)
}
