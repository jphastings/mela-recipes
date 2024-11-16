package mela

import (
	"github.com/jphastings/recipes/internal/formats"
)

func importRecipe(r formats.Recipe) (formats.Recipe, error) {
	if _, ok := r.(*Recipe); ok {
		return r, nil
	}

	ir, err := r.Export()
	if err != nil {
		return nil, err
	}

	return &Recipe{
		filename: r.Filename(),
		ID:       ir.ID,
		Title:    ir.Title,
		// TODO: Link
		Text:         ir.Description,
		Ingredients:  ingredientsToSecSeq(ir.Ingredients),
		Instructions: instructionsToSecSeq(ir.Instructions),
		// TODO: Images
		Yield: PeopleCount(ir.Yield),

		PrepTime:  MaybeDuration(ir.PrepTime.String()),
		CookTime:  MaybeDuration(ir.CookTime.String()),
		TotalTime: MaybeDuration(ir.TotalTime.String()),
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

func importCollection(name string, rs []formats.Recipe) (formats.RecipeCollection, error) {
	rc := &RecipeCollection{
		name:    name,
		recipes: make([]*Recipe, len(rs)),
	}

	for i, ir := range rs {
		r, err := importRecipe(ir)
		if err != nil {
			return rc, err
		}
		rc.recipes[i] = r.(*Recipe)
	}

	return rc, nil
}
