package paprika

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

	uid := ir.ID
	if uid == "" {
		if newID, err := uuid.NewUUID(ir.Title); err == nil {
			uid = newID.String()
		}
	}

	pr := &Recipe{
		filename: formats.WithoutExt(r.Filename()),
		UID:      uid,
		Title:    ir.Title,
		// TODO: Source / SourceURL
		Description: ir.Description,
		Notes:       ir.Notes,
		Ingredients: formats.IngredientsToSectioned(ir.Ingredients),
		Directions:  formats.TitledListsToSectioned(ir.Instructions),
		Servings:    ir.Yield,
		Categories:  importTags(ir.Tags),

		PrepTime:  formats.FormatDuration(ir.PrepTime),
		CookTime:  formats.FormatDuration(ir.CookTime),
		TotalTime: formats.FormatDuration(ir.TotalTime),
	}

	// Paprika's export format carries a single primary photo; any further
	// interchange images are dropped on conversion.
	if len(ir.Images) > 0 {
		pr.PhotoData = ir.Images[0]
	}

	return pr, nil
}

// importTags carries the interchange recipe's tags across as Paprika categories,
// always returning a non-nil slice.
func importTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}
