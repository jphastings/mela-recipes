package formats

import (
	"fmt"
	"io"
	"time"

	"github.com/jphastings/recipes/internal/standardize"
)

var _ Recipe = (*InterchangeRecipe)(nil)

// A generic and internal structure for recipes that is used for conversion
// ⚠️ This struct is highly likely to change subtly with each new recipe format added to this library.
type InterchangeRecipe struct {
	filename     string
	ID           string             `json:"-"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Yield        string             `json:"yield"`
	Ingredients  []IngredientGroup  `json:"ingredients"`
	Instructions []InstructionGroup `json:"instructions"`

	// TODO: Handle these
	PrepTime  time.Duration `json:"-"`
	CookTime  time.Duration `json:"-"`
	TotalTime time.Duration `json:"-"`
}

type IngredientGroup struct {
	GroupName   string   `json:"groupName"`
	Ingredients []string `json:"ingredients"`
}

type InstructionGroup struct {
	GroupName string   `json:"groupName"`
	Steps     []string `json:"steps"`
}

func (ir InterchangeRecipe) Filename() string                   { return ir.filename }
func (ir InterchangeRecipe) Format() *Format                    { return nil }
func (ir InterchangeRecipe) Export() (InterchangeRecipe, error) { return ir, nil }
func (ir InterchangeRecipe) Name() string                       { return ir.Title }
func (ir InterchangeRecipe) Standardize() ([]standardize.Std, error) {
	return nil, fmt.Errorf("standardising a recipe in the interchange format is not supported yet")
}

func (ir InterchangeRecipe) Marshal(io.Writer) error {
	return fmt.Errorf("writing a recipe in the interchange format is not supported")
}
