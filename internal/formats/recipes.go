package formats

import (
	_ "embed"

	"github.com/jphastings/recipes/internal/llm"
	"github.com/jphastings/recipes/utils"
)

//go:embed recipes.schema.json
var RecipesSchema llm.RawJSON

type CollectionDetails struct {
	Name              string
	Filename          string
	Book              utils.Book
	OverwriteExisting bool
}

// Represents a writer for a recipe collection in any format
type CollectionWriter interface {
	// Returns the output filename the collection is being written to
	Filename() string
	// Adds the provided recipe to this collection
	Add(Recipe) error
	// Finishes off writing the recipe collection (no more recipes can be added)
	Close() error
}
