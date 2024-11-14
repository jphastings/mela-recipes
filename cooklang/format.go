package cooklang

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/jphastings/recipes/internal/formats"
)

const (
	collectionExt = ".cook"
)

var FormatInfo = formats.Format{
	Name: "Cooklang",
	URL:  "https://cooklang.org",
	Features: formats.Features{
		ParseRecipe: true,
	},
	ExtensionCollection: collectionExt,
	Parse: func(formats.Bundle) (formats.Recipe, formats.RecipeCollection, error) {
		return nil, nil, fmt.Errorf("cooklang parsing not yet implemented")
	},
	Bundle: bundle,
}

var bundleExts = []string{collectionExt, ".jpg", ".jpeg", ".png"}
var sectionSuffix = regexp.MustCompile(`\.\d+$`)

func bundle(files []string) (bundles []formats.Bundle, unused []string) {
	idx := make(map[string][]string)

	for _, f := range files {
		ext := path.Ext(f)
		if !slices.Contains(bundleExts, ext) {
			unused = append(unused, f)
			continue
		}
		k := strings.TrimSuffix(f, ext)
		if ext != collectionExt {
			k = sectionSuffix.ReplaceAllString(k, "")
		}
		idx[k] = append(idx[k], f)
	}

	for _, b := range idx {
		bundles = append(bundles, b)
	}
	return
}
