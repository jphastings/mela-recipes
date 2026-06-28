package recipemd

import (
	"errors"
	"fmt"
	"io"

	"github.com/jphastings/recipes/internal/formats"
)

// Parse reads a single .md file and emits the RecipeMD recipe it contains.
// Markdown files that aren't RecipeMD documents are reported as a non-fatal parse
// error, so a directory of mixed Markdown can be converted without aborting.
func Parse(b formats.Bundle, opts formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
	decode := func(r io.Reader, filename string) (formats.Recipe, error) {
		return decodeRecipe(r, filename, opts.AllowNetwork)
	}
	pe, err := formats.ParseRecipeFile(b[0], decode)
	return pe, nil, err
}

func decodeRecipe(r io.Reader, filename string, allowNetwork bool) (formats.Recipe, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	ir, err := parseRecipe(source, allowNetwork)
	if err != nil {
		return nil, err
	}
	if errs := ir.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("incomplete RecipeMD recipe: %w", errors.Join(errs...))
	}

	return &Recipe{filename: filename, ir: ir}, nil
}
