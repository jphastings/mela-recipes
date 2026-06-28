package mealmaster

import "github.com/jphastings/recipes/internal/formats"

const recipeExt = ".mmf"

// FormatInfo describes the MealMaster importer. MealMaster is a 1990s DOS recipe
// program whose plain-text `.mmf` export holds a vast body of BBS-era recipe
// archives. The format is reverse-engineered (no formal spec) and a single file
// usually concatenates many recipes, so it is read-only: writing is not
// supported.
var FormatInfo = &formats.Format{
	Name: "MealMaster",
	URL:  "https://www.ffts.com/mmformat.txt",
	Features: formats.Features{
		ParseRecipe: true,
	},
	Extension: recipeExt,
	Parse:     Parse,
	Bundle:    formats.BundleByExtension(recipeExt),
}
