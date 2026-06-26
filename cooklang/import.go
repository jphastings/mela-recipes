package cooklang

import (
	"fmt"

	"github.com/jphastings/recipes/internal/formats"
)

func importRecipe(r formats.Recipe) (formats.Recipe, error) {
	if r == nil {
		return nil, fmt.Errorf("provided recipe is nil")
	}
	// If it's already in this format then no conversion is needed.
	if _, ok := r.(*Recipe); ok {
		return r, nil
	}

	ir, err := r.Export()
	if err != nil {
		return nil, err
	}

	return &Recipe{
		filename: formats.WithoutExt(r.Filename()),
		ir:       ir,
	}, nil
}
