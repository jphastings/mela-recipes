package epub

import (
	"fmt"
	"io"

	"github.com/jphastings/recipes/internal/formats"
)

var _ formats.RecipeCollection = (*RecipeCollection)(nil)

type RecipeCollection struct {
	name     string
	filename string
	// ePub is not an output-capable format, so the internal storage method is actually the interchange format
	recipes []formats.InterchangeRecipe
}

func (rc *RecipeCollection) Filename() string        { return rc.filename + FormatInfo.ExtensionCollection }
func (rc *RecipeCollection) Format() *formats.Format { return FormatInfo }
func (rc *RecipeCollection) Name() string            { return rc.name }
func (rc *RecipeCollection) Recipes() []formats.Recipe {
	out := make([]formats.Recipe, len(rc.recipes))
	for i, r := range rc.recipes {
		out[i] = formats.Recipe(r)
	}
	return out
}

func (rc *RecipeCollection) Add(rs ...formats.Recipe) error {
	return fmt.Errorf("the ePub format doesn't support adding recipes")
}

func (rc *RecipeCollection) Marshal(io.Writer) error {
	return fmt.Errorf("writing recipes to the ePub format is not supported")
}
