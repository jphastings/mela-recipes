package mela

import (
	"fmt"

	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/uuid"
)

func ImportRecipe(r formats.Recipe) (formats.Recipe, error) {
	if r == nil {
		return nil, fmt.Errorf("provided recipe is nil")
	}

	// If its already in this format then no conversion is needed
	if _, ok := r.(*Recipe); ok {
		return r, nil
	}

	ir, err := r.Export()
	if err != nil {
		return nil, err
	}

	id := ir.ID
	if id == "" {
		if newID, err := uuid.NewUUID(ir.Title); err == nil {
			id = newID.String()
		}
	}

	mr := &Recipe{
		filename:     formats.WithoutExt(r.Filename()),
		ID:           id,
		Title:        ir.Title,
		Link:         ir.Source.URI,
		Text:         ir.Description,
		Ingredients:  formats.IngredientsToSectioned(ir.Ingredients),
		Instructions: formats.TitledListsToSectioned(ir.Instructions),
		Images:       ir.Images,

		Categories: importTags(ir.Tags),
		Yield:      PeopleCount(ir.Yield),

		PrepTime:  formats.FormatDuration(ir.PrepTime),
		CookTime:  formats.FormatDuration(ir.CookTime),
		TotalTime: formats.FormatDuration(ir.TotalTime),
	}

	return mr, nil
}

// importTags carries the interchange recipe's tags across as Mela categories,
// always returning a non-nil slice.
func importTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}
