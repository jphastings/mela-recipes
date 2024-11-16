package epub

import (
	"fmt"

	"github.com/jphastings/recipes/internal/formats"
)

func (r RecipeCollection) Export() (formats.InterchangeRecipe, error) {
	return formats.InterchangeRecipe{}, fmt.Errorf("export of epub recipe format not yet implemented")
}
