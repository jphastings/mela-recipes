package mela

import (
	"github.com/jphastings/recipes/internal/formats"
	"github.com/jphastings/recipes/internal/uuid"
	"github.com/jphastings/recipes/utils"
)

func ImportRecipe(r formats.Recipe) (formats.Recipe, error) {
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

	return &Recipe{
		filename: r.Filename(),
		ID:       id,
		Title:    ir.Title,
		// TODO: Link (source)
		Text:         ir.Description,
		Ingredients:  ingredientsToSecSeq(ir.Ingredients),
		Instructions: instructionsToSecSeq(ir.Instructions),
		Images:       []utils.B64Image{}, // TODO: Images

		Categories: []string{},
		Yield:      PeopleCount(ir.Yield),

		PrepTime:  ir.PrepTime,
		CookTime:  ir.CookTime,
		TotalTime: ir.TotalTime,
	}, nil
}

func ingredientsToSecSeq(igs []formats.IngredientGroup) SectionedSequence {
	m := make(map[string][]string)
	for _, ig := range igs {
		m[ig.GroupName] = ig.Ingredients
	}

	return NewSectionedSequence(m)
}

func instructionsToSecSeq(igs []formats.InstructionGroup) SectionedSequence {
	m := make(map[string][]string)
	for _, ig := range igs {
		m[ig.GroupName] = ig.Steps
	}

	return NewSectionedSequence(m)
}
