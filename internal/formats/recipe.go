package formats

import (
	"fmt"
	"io"
	"time"

	"github.com/jphastings/recipes/internal/ingredients"
	"github.com/jphastings/recipes/internal/standardize"
	"github.com/jphastings/recipes/utils"
)

// Represents a recipe in a specific format, held in memory
type Recipe interface {
	// The name of the recipe
	Name() string
	// The format this recipe is in
	Format() *Format
	// Returns the stored filename (without extension), or an appropriate generated one
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

var _ Recipe = (*InterchangeRecipe)(nil)

// A generic and internal structure for recipes that is used for conversion
// ⚠️ This struct is highly likely to change subtly with each new recipe format added to this library.
type InterchangeRecipe struct {
	filename     string
	ID           string
	Title        string
	Description  string
	Yield        string
	Ingredients  []IngredientGroup
	Instructions []TitledList
	Notes        string
	Source       Source

	Images []utils.B64Image

	PrepTime  *time.Duration
	CookTime  *time.Duration
	TotalTime *time.Duration

	Tags []string
}

func NewInterchangeRecipe() InterchangeRecipe {
	return InterchangeRecipe{
		Images: []utils.B64Image{},
		Tags:   []string{},
	}
}

type TitledList struct {
	Title string
	List  []string
}

// IngredientGroup is a titled section of structured ingredients. Items before
// the first heading live under the empty-string Title.
type IngredientGroup struct {
	Title string
	Items []ingredients.IngredientUse
}

// Source describes where a recipe came from. URI may be a web address
// (https://…) or a URN such as urn:isbn:9781234567890.
type Source struct {
	Name string
	URI  string
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

func (r InterchangeRecipe) Validate() []error {
	var errs []error

	if r.Title == "" {
		errs = append(errs, fmt.Errorf("recipe has no title"))
	}

	items := 0
	for _, g := range r.Ingredients {
		items += len(g.Items)
	}
	if items == 0 {
		errs = append(errs, fmt.Errorf("recipe has no ingredients"))
	}

	if len(r.Instructions) == 0 {
		errs = append(errs, fmt.Errorf("recipe has no instructions"))
	}

	return errs
}
