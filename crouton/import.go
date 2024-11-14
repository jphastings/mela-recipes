package crouton

import (
	"github.com/jphastings/recipes/internal/formats"
)

func newFromInterchange(ir formats.InterchangeRecipe) (formats.Recipe, error) {
	return &Recipe{
		filename:   ir.Filename,
		RecipeName: ir.Title,
		// TODO: Other fields
	}, nil
}
