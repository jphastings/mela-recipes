package recipemd

import "github.com/jphastings/recipes/internal/formats"

const recipeExt = ".md"

// FormatInfo describes the RecipeMD importer/exporter: an open, CommonMark-based
// Markdown recipe standard that is human-writable and git-friendly, making it a
// good import source and a stable round-trip storage target.
var FormatInfo = &formats.Format{
	Name: "RecipeMD",
	URL:  "https://recipemd.org",
	Features: formats.Features{
		ParseRecipe: true,
		WriteRecipe: true,
	},
	Extension: recipeExt,
	Import:    Import,
	Parse:     Parse,
	Bundle:    formats.BundleByExtension(recipeExt),
}
