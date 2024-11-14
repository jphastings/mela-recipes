package crouton

import (
	"fmt"

	"github.com/jphastings/recipes/internal/formats"
)

func (r *Recipe) Export() (formats.InterchangeRecipe, error) {
	return formats.InterchangeRecipe{}, fmt.Errorf("export of crouton recipe format not yet implemented")
}
