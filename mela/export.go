package mela

import (
	"errors"

	"github.com/jphastings/recipes/internal/formats"
)

// Export converts a Mela recipe into the interchange format. It is the inverse
// of ImportRecipe; fields the interchange format cannot represent (eg. link,
// nutrition) are dropped. The interchange filename is left unset, as that field
// has no exported setter.
func (r *Recipe) Export() (formats.InterchangeRecipe, error) {
	ir := formats.NewInterchangeRecipe()
	ir.ID = r.ID
	ir.Title = r.Title
	ir.Description = r.Text
	ir.Notes = r.Notes
	ir.Yield = string(r.Yield)
	ir.Ingredients = formats.SectionedToIngredientGroups(r.Ingredients)
	ir.Instructions = formats.SectionedToTitledLists(r.Instructions)

	if r.Categories != nil {
		ir.Tags = r.Categories
	}
	if r.Images != nil {
		ir.Images = r.Images
	}

	prep, prepErr := r.PrepTime.Parse()
	cook, cookErr := r.CookTime.Parse()
	total, totalErr := r.TotalTime.Parse()
	ir.PrepTime = prep
	ir.CookTime = cook
	ir.TotalTime = total

	return ir, errors.Join(prepErr, cookErr, totalErr)
}
