package schemaorg

import "github.com/jphastings/recipes/utils"

// maxFetchImages caps how many image URLs are downloaded per recipe — a
// schema.org "image" value is often the same photo at several aspect ratios.
const maxFetchImages = 4

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

		img, err := utils.FetchImage(u)
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
