package mela

import (
	"github.com/jphastings/recipes/internal/formats"
)

func newFromInterchange(ir formats.InterchangeRecipe) (formats.Recipe, error) {
	return &Recipe{
		filename: ir.Filename,
		ID:       ir.ID,
		Title:    ir.Title,
		Text:     ir.Description,
		// TODO: Other fields
	}, nil
}

func newFromInterchangeCollection(name string, rs []formats.InterchangeRecipe) (formats.RecipeCollection, error) {
	rc := &RecipeCollection{
		name:    name,
		recipes: make([]*Recipe, len(rs)),
	}

	for i, ir := range rs {
		r, err := newFromInterchange(ir)
		if err != nil {
			return rc, err
		}
		rc.recipes[i] = r.(*Recipe)
	}

	return rc, nil
}
