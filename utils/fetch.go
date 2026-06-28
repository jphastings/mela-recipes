package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// maxFetchBytes bounds a single image download so a stray huge file can't blow
// up memory.
const maxFetchBytes = 25 << 20

// maxFetchImages caps how many image URLs are downloaded per recipe — a source
// often lists the same photo at several aspect ratios.
const maxFetchImages = 4

var fetchClient = &http.Client{Timeout: 20 * time.Second}

// FetchImages downloads up to maxFetchImages distinct URLs, embedding each as an
// optimised image. Unreachable, oversized, or non-image URLs are skipped.
func FetchImages(urls []string) []B64Image {
	var imgs []B64Image
	seen := make(map[string]bool, len(urls))
	attempts := 0
	for _, u := range urls {
		if seen[u] {
			continue
		}
		seen[u] = true
		if attempts++; attempts > maxFetchImages {
			break
		}

		img, err := FetchImage(u)
		if err != nil {
			continue
		}
		if opt, changed, err := img.Optimize(); err == nil && changed {
			img = opt
		}
		imgs = append(imgs, img)
	}
	return imgs
}

// FetchImage downloads an image from url and returns its raw bytes. The download
// is time-limited, size-limited, and rejected if the server reports a non-image
// content type. Callers that want the image resized or recompressed should call
// Optimize on the result.
func FetchImage(url string) (B64Image, error) {
	resp, err := fetchClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: status %d", url, resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "image/") {
		return nil, fmt.Errorf("fetching %s: unexpected content type %q", url, ct)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchBytes))
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("fetching %s: empty response", url)
	}
	return B64Image(data), nil
}
