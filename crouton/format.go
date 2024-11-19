package crouton

import (
	"fmt"

	"github.com/jphastings/recipes/internal/formats"
)

const recipeExt = ".crumb"

var FormatInfo = &formats.Format{
	Name: "Crouton",
	URL:  "https://crouton.app",
	Features: formats.Features{
		ParseRecipe: true,
		WriteRecipe: true,
	},
	Extension: recipeExt,
	Import:    importRecipe,
	Parse: func(formats.Bundle, formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
		return nil, nil, fmt.Errorf("crouton parsing not yet implemented")
	},
	Bundle: formats.BundleByExtension(recipeExt),
}
