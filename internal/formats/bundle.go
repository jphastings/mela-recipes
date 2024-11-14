package formats

import (
	"path"
	"slices"
)

type Bundle []string
type Bundler func(files []string) (bundles []Bundle, unused []string)

func BundleByExtension(extensions ...string) Bundler {
	return func(files []string) (bundles []Bundle, unused []string) {
		for _, f := range files {
			if slices.Contains(extensions, path.Ext(f)) {
				bundles = append(bundles, []string{f})
			} else {
				unused = append(unused, f)
			}
		}
		return
	}
}
