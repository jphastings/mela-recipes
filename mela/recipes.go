package mela

import "github.com/jphastings/recipes/internal/formats"

// ParseRecipesFile parses a .melarecipes collection (a zip of .melarecipe
// files), streaming its recipes as they are decoded.
func ParseRecipesFile(filename string) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	return formats.ParseZipCollection(filename, recipeExt, decodeRecipe)
}

// NewCollection creates a new .melarecipes collection ready to add recipes into.
func NewCollection(cd formats.CollectionDetails) (formats.CollectionWriter, error) {
	return formats.NewZipCollection(cd, collectionExt, ImportRecipe)
}
