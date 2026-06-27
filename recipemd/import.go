package recipemd

import "github.com/jphastings/recipes/internal/formats"

// Import converts a recipe from any format into a RecipeMD recipe, ready to be
// written with Marshal.
func Import(r formats.Recipe) (formats.Recipe, error) {
	if rec, ok := r.(*Recipe); ok {
		return rec, nil
	}

	ir, err := r.Export()
	if err != nil {
		return nil, err
	}
	return &Recipe{filename: formats.WithoutExt(r.Filename()), ir: ir}, nil
}
