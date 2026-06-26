package paprika

import (
	"errors"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/utils"
)

// Export converts a Paprika recipe into the interchange format. It is the
// inverse of ImportRecipe; fields the interchange format cannot represent (eg.
// nutritional_info, source, rating, difficulty) are dropped.
func (r *Recipe) Export() (formats.InterchangeRecipe, error) {
	ir := formats.NewInterchangeRecipe()
	ir.ID = r.UID
	ir.Title = r.Title
	ir.Description = r.Description
	ir.Notes = r.Notes
	ir.Yield = r.Servings
	ir.Ingredients = formats.SectionedToTitledLists(r.Ingredients)
	ir.Instructions = formats.SectionedToTitledLists(r.Directions)

	if r.Categories != nil {
		ir.Tags = r.Categories
	}
	if len(r.PhotoData) > 0 {
		ir.Images = []utils.B64Image{r.PhotoData}
	}

	prep, prepErr := r.PrepTime.Parse()
	cook, cookErr := r.CookTime.Parse()
	total, totalErr := r.TotalTime.Parse()
	ir.PrepTime = prep
	ir.CookTime = cook
	ir.TotalTime = total

	return ir, errors.Join(prepErr, cookErr, totalErr)
}
