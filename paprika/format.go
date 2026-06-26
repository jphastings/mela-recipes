package paprika

import "github.com/jphastings/recipes/internal/formats"

const (
	recipeExt     = ".paprikarecipe"
	collectionExt = ".paprikarecipes"
)

var FormatInfo = &formats.Format{
	Name: "Paprika",
	URL:  "https://www.paprikaapp.com",
	Features: formats.Features{
		ParseRecipe:     true,
		WriteRecipe:     true,
		ParseCollection: true,
		WriteCollection: true,
	},
	Extension:           recipeExt,
	ExtensionCollection: collectionExt,
	Import:              ImportRecipe,
	NewCollection:       NewCollection,
	Parse:               Parse,
	Bundle:              formats.BundleByExtension(recipeExt, collectionExt),
}
