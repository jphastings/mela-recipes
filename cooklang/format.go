package cooklang

import (
	"fmt"
	"os"

	"github.com/jphastings/recipes/internal/formats"
)

const (
	collectionExt = "epub"
)

var format = formats.Format{
	Name: "Cooklang",
	URL:  "https://cooklang.org/",
	Features: formats.Features{
		ParseRecipe: true,
	},
	ExtensionCollection: collectionExt,
	Parse: func(*os.File) (formats.Recipe, formats.RecipeCollection, error) {
		return nil, nil, fmt.Errorf("cooklang parsing not yet implemented")
	},
}

func init() {
	formats.Register(format)
}
