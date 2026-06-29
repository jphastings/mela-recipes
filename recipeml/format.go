package recipeml

import "github.com/jphastings/recipes/internal/formats"

const recipeExt = ".xml"

// FormatInfo describes the RecipeML importer. RecipeML is an open, published XML
// recipe format from the early 2000s with a clean structure (separate quantity /
// unit / item nodes). The ecosystem is dead but sizeable `.xml` archives remain,
// so it is read-only: writing is not supported.
//
// `.xml` is a generic extension, so the bundle claims every `.xml` file eagerly
// (no other format reads `.xml`); Parse confirms the root is <recipeml> and
// reports a clear non-fatal error for any other XML.
var FormatInfo = &formats.Format{
	Name: "RecipeML",
	URL:  "http://www.formatdata.com/recipeml/",
	Features: formats.Features{
		ParseRecipe: true,
	},
	Extension: recipeExt,
	Parse:     Parse,
	Bundle:    formats.BundleByExtension(recipeExt),
}
