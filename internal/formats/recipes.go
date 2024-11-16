package formats

import (
	_ "embed"

	"github.com/jphastings/recipes/internal/llm"
)

//go:embed recipes.schema.json
var RecipesSchema llm.RawJSON
