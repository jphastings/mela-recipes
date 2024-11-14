package crouton

import "github.com/jphastings/recipes/internal/formats"

const recipeExt = "crumb"

var format = formats.Format{
	Name: "Crouton",
	URL:  "https://crouton.app",
	Features: formats.Features{
		ParseRecipe: true,
		WriteRecipe: true,
	},
	Extension: recipeExt,
	New:       newFromInterchange,
}

func init() {
	formats.Register(format)
}
