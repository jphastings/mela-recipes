package recipeml

import (
	"errors"
	"io"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/standardize"
)

var _ formats.Recipe = (*Recipe)(nil)

// Recipe is a single recipe read out of a RecipeML document, held as its mapped
// interchange form. The format is read-only: emitting RecipeML output is not
// supported.
type Recipe struct {
	filename string
	ir       formats.InterchangeRecipe
}

func (r *Recipe) Name() string            { return r.ir.Title }
func (r *Recipe) Format() *formats.Format { return FormatInfo }

func (r *Recipe) Filename() string {
	name := r.filename
	if name == "" {
		name = standardize.StringToFilename(r.ir.Title)
	}
	return name + FormatInfo.Extension
}

func (r *Recipe) Standardize() ([]standardize.Std, error) {
	if r.filename == "" && r.ir.Title != "" {
		r.filename = standardize.StringToFilename(r.ir.Title)
		return []standardize.Std{standardize.StdFilename}, nil
	}
	return nil, nil
}

func (r *Recipe) Export() (formats.InterchangeRecipe, error) { return r.ir, nil }

func (r *Recipe) Marshal(io.Writer) error {
	return errors.New("writing RecipeML recipes is not supported")
}
