package epub

import (
	"fmt"
	"os"

	"github.com/jphastings/recipes/internal/formats"
)

const (
	collectionExt = "epub"
)

var format = formats.Format{
	Name: "ePub cookbook",
	URL:  "https://en.wikipedia.org/wiki/EPUB",
	Features: formats.Features{
		ParseCollection: true,
	},
	ExtensionCollection: collectionExt,
	Parse: func(*os.File) (formats.Recipe, formats.RecipeCollection, error) {
		return nil, nil, fmt.Errorf("epub parsing not yet implemented")
	},
}

func init() {
	formats.Register(format)
}
