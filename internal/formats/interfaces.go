package formats

import (
	"io"

	"github.com/jphastings/recipes/internal/standardize"
)

// Represents a recipe in a specific format, held in memory
type Recipe interface {
	// The name of the recipe
	Name() string
	// The format this recipe is in
	Format() Format
	// Returns the stored filename (with extension), or an appropriate generated one
	Filename() string
	// Destructively applies all standardizations to this recipe, returning
	// Typically this includes:
	// - Setting the filename from the recipe title
	// - Optimising any images attached
	// - Extracting and standardising the ISBN & physical book data found in the notes field, if one is present
	Standardize() ([]standardize.Std, error)
	// Mutates the recipe in this format into the interchange format.
	Export() (InterchangeRecipe, error)
	// Marshals the recipe (in this format) to the provided writer
	Marshal(io.Writer) error
}

// Represents a collection of recipes held in a specific format, held in memory
type RecipeCollection interface {
	// The name of the recipe collection
	Name() string
	// The format this recipe collection is in
	Format() Format
	// Returns the stored filename (with extension), or an appropriate generated one
	Filename() string
	// Adds the provided recipe to this collection
	Add(...Recipe) error
	// Allows iteration through each of the recipes in the collection
	Recipes() []Recipe
	// Marshals the recipe (in this format) to the provided writer
	Marshal(io.Writer) error
}
