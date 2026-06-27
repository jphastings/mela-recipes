package schemaorg

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jphastings/recipes/utils"
)

const (
	// maxFetchImages caps how many image URLs are downloaded per recipe — a
	// schema.org "image" value is often the same photo at several aspect ratios.
	maxFetchImages = 4
	// maxImageBytes bounds a single download so a stray huge file can't blow up
	// memory.
	maxImageBytes = 25 << 20
)

var imageClient = &http.Client{Timeout: 20 * time.Second}

// fetchImages downloads up to maxFetchImages distinct URLs, embedding each as an
// optimised image. Unreachable, oversized, or non-image URLs are skipped.
func fetchImages(urls []string) []utils.B64Image {
	var imgs []utils.B64Image
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
		if img, err := fetchImage(u); err == nil {
			imgs = append(imgs, img)
		}
	}
	return imgs
}

func fetchImage(url string) (utils.B64Image, error) {
	resp, err := imageClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: status %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageBytes))
	if err != nil {
		return nil, err
	}

	img := utils.B64Image(data)
	if opt, changed, err := img.Optimize(); err == nil && changed {
		img = opt
	}
	return img, nil
}
