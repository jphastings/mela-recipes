package paprika

import "github.com/jphastings/recipes/internal/formats"

// NewCollection creates a new .paprikarecipes collection ready to add recipes
// into.
func NewCollection(cd formats.CollectionDetails) (formats.CollectionWriter, error) {
	return formats.NewZipCollection(cd, collectionExt, ImportRecipe)
}
