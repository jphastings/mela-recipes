package schemaorg

import "github.com/jphastings/recipes/internal/formats"

const recipeExt = ".html"

// FormatInfo describes the schema.org recipe importer: it reads structured recipe
// data (JSON-LD, microdata, or the h-recipe microformat) out of saved web pages
// and raw JSON-LD / Nextcloud recipe.json files. It is read-only.
var FormatInfo = &formats.Format{
	Name: "schema.org",
	URL:  "https://schema.org/Recipe",
	Features: formats.Features{
		ParseRecipe: true,
	},
	Extension: recipeExt,
	Parse:     Parse,
	Bundle:    formats.BundleByExtension(".html", ".htm", ".json"),
}
