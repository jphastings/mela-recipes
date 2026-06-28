package crouton

import (
	"fmt"

	"github.com/jphastings/recipes/internal/formats"
)

const recipeExt = ".crumb"

var FormatInfo = &formats.Format{
	Name: "Crouton",
	URL:  "https://crouton.app",
	Features: formats.Features{
		ParseRecipe: true,
		WriteRecipe: true,
	},
	Lossiness: formats.Lossiness{
		// Crouton -> interchange loses only volatile metadata (regenerated UUIDs,
		// senderName), which is not content, so OnExport is empty. Interchange ->
		// Crouton drops content Crouton's data model can't hold:
		OnImport: []formats.LossyField{
			{
				Field:   "Description",
				Reason:  "Crouton has no description field, only notes",
				Present: func(ir formats.InterchangeRecipe) bool { return ir.Description != "" },
			},
			{
				Field:   "Tags",
				Reason:  "Crouton groups recipes by folder, not by tag name",
				Present: func(ir formats.InterchangeRecipe) bool { return len(ir.Tags) > 0 },
			},
			{
				Field:   "PrepTime",
				Reason:  "Crouton records only total and cooking durations",
				Present: func(ir formats.InterchangeRecipe) bool { return ir.PrepTime != nil },
			},
			{
				Field:   "Yield",
				Reason:  "Crouton servings are a whole number; extra detail is dropped",
				Present: func(ir formats.InterchangeRecipe) bool { return isFreeFormYield(ir.Yield) },
			},
		},
	},
	Extension: recipeExt,
	Import:    importRecipe,
	Parse: func(formats.Bundle, formats.ParseOptions) (<-chan formats.ParseEvent, *formats.CollectionDetails, error) {
		return nil, nil, fmt.Errorf("crouton parsing not yet implemented")
	},
	Bundle: formats.BundleByExtension(recipeExt),
}
