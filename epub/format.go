package epub

import (
	"fmt"

	"github.com/jphastings/recipes/internal/formats"
)

const (
	collectionExt = ".epub"
)

var FormatInfo = formats.Format{
	Name: "ePub",
	URL:  "https://en.wikipedia.org/wiki/EPUB",
	Features: formats.Features{
		ParseCollection: true,
	},
	ExtensionCollection: collectionExt,
	Parse: func(formats.Bundle) (formats.Recipe, formats.RecipeCollection, error) {
		return nil, nil, fmt.Errorf("ePub parsing not yet implemented")
	},
	Bundle: formats.BundleByExtension(collectionExt),
}
