package crouton

import (
	"strconv"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/utils"
)

func (r *Recipe) Export() (formats.InterchangeRecipe, error) {
	ir := formats.NewInterchangeRecipe()
	ir.ID = r.UUID.String()
	ir.Title = r.RecipeName
	ir.Source = formats.Source{Name: r.SourceName, URI: string(r.WebLink)}
	ir.Notes = r.Notes
	if r.Serves != 0 {
		ir.Yield = strconv.Itoa(int(r.Serves))
	}
	ir.Ingredients = ingredientsToGroups(r.Ingredients)
	ir.Instructions = stepsToTitledLists(r.Steps)
	ir.TotalTime = minutesToPtr(r.Duration)
	ir.CookTime = minutesToPtr(r.CookingDuration)

	ir.Images = make([]utils.B64Image, len(r.Images))
	for i, img := range r.Images {
		ir.Images[i] = utils.B64Image(img)
	}

	return ir, nil
}
