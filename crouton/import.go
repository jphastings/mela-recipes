package crouton

import (
	"fmt"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/uuid"
)

func importRecipe(r formats.Recipe) (formats.Recipe, error) {
	if r == nil {
		return nil, fmt.Errorf("provided recipe is nil")
	}
	// If it's already in this format then no conversion is needed
	if _, ok := r.(*Recipe); ok {
		return r, nil
	}

	ir, err := r.Export()
	if err != nil {
		return nil, err
	}

	cr := &Recipe{
		filename:        formats.WithoutExt(r.Filename()),
		UUID:            importUUID(ir),
		RecipeName:      ir.Title,
		SourceName:      ir.Source.Name,
		WebLink:         Link(ir.Source.URI),
		Notes:           ir.Notes,
		Serves:          parseYield(ir.Yield),
		Duration:        ptrToMinutes(ir.TotalTime),
		CookingDuration: ptrToMinutes(ir.CookTime),
	}

	if cr.Ingredients, err = groupsToIngredients(ir.Ingredients); err != nil {
		return nil, err
	}
	if cr.Steps, err = titledListsToSteps(ir.Instructions); err != nil {
		return nil, err
	}

	cr.Images = make([]B64Image, len(ir.Images))
	for i, img := range ir.Images {
		cr.Images[i] = B64Image(img)
	}

	return cr, nil
}

// importUUID reuses the interchange ID when it is a valid UUID (so a crouton recipe
// round-trips exactly), otherwise derives a deterministic one from the ID or title.
func importUUID(ir formats.InterchangeRecipe) uuid.UUID {
	if u, err := uuid.Parse(ir.ID); err == nil {
		return u
	}
	seed := ir.ID
	if seed == "" {
		seed = ir.Title
	}
	u, _ := uuid.NewUUID(seed)
	return u
}
