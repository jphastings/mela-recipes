package epub

import (
	"github.com/jphastings/recipes/internal/formats"
)

const (
	collectionExt = ".epub"
)

var FormatInfo = &formats.Format{
	Name: "ePub",
	URL:  "https://en.wikipedia.org/wiki/EPUB",
	Features: formats.Features{
		ParseCollection: true,
	},
	ExtensionCollection: collectionExt,

	Parse:  Parse,
	Bundle: formats.BundleByExtension(collectionExt),
}
