package formats

import (
	_ "embed"

	"github.com/jphastings/recipes/internal/llm"
	"github.com/jphastings/recipes/utils"
)

//go:embed recipes.schema.json
var RecipesSchema llm.RawJSON

type CollectionDetails struct {
	Name string
	Book utils.Book
}
